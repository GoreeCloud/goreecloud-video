package playback

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestFileTransformJobRepositoryCreateLoadAndSave(t *testing.T) {
	now := time.Date(2026, 9, 7, 22, 0, 0, 0, time.UTC)
	repository, err := NewFileTransformJobRepository(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	request := TransformRequest{sourceID: "source-1", requirements: []TransformRequirement{TransformMediaTranscode}}
	job, err := NewTransformJob("job/../../one", request, now)
	if err != nil {
		t.Fatal(err)
	}

	created, err := repository.Create(context.Background(), job)
	if err != nil || created.Revision() != 0 {
		t.Fatalf("created=%+v err=%v", created, err)
	}
	path := repository.recordPath(job.ID())
	if filepath.Dir(path) != repository.root || strings.Contains(filepath.Base(path), "job") {
		t.Fatalf("unsafe transform job path %q", path)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0o077 != 0 {
		t.Fatalf("transform job permissions=%#o", info.Mode().Perm())
	}

	loaded, found, err := repository.Load(context.Background(), job.ID())
	if err != nil || !found || !sameTransformJobRecord(loaded, created) {
		t.Fatalf("loaded=%+v found=%v err=%v", loaded, found, err)
	}

	running, err := job.Transition(TransformJobRunning, now.Add(time.Second), "")
	if err != nil {
		t.Fatal(err)
	}
	saved, err := repository.Save(context.Background(), 0, running)
	if err != nil || saved.Revision() != 1 {
		t.Fatalf("saved=%+v err=%v", saved, err)
	}
	restored, err := saved.Restore()
	if err != nil || restored.State() != TransformJobRunning || restored.Attempts() != 1 {
		t.Fatalf("restored=%+v err=%v", restored, err)
	}
}

func TestFileTransformJobRepositoryRejectsDuplicateAndStaleWrites(t *testing.T) {
	now := time.Date(2026, 9, 7, 22, 0, 0, 0, time.UTC)
	repository, _ := NewFileTransformJobRepository(t.TempDir())
	request := TransformRequest{sourceID: "source-1", requirements: []TransformRequirement{TransformMediaTranscode}}
	job, _ := NewTransformJob("job-1", request, now)
	if _, err := repository.Create(context.Background(), job); err != nil {
		t.Fatal(err)
	}
	if _, err := repository.Create(context.Background(), job); !errors.Is(err, ErrTransformJobAlreadyExists) {
		t.Fatalf("duplicate create error=%v", err)
	}

	running, _ := job.Transition(TransformJobRunning, now.Add(time.Second), "")
	current, err := repository.Save(context.Background(), 0, running)
	if err != nil {
		t.Fatal(err)
	}
	failed, _ := running.Transition(TransformJobFailed, now.Add(2*time.Second), "worker-exit")
	returned, err := repository.Save(context.Background(), 0, failed)
	if !errors.Is(err, ErrStaleTransformJobRevision) || returned.Revision() != current.Revision() {
		t.Fatalf("returned=%+v error=%v", returned, err)
	}
}

func TestFileTransformJobRepositoryRejectsMalformedAndWrongIdentityFiles(t *testing.T) {
	now := time.Date(2026, 9, 7, 22, 0, 0, 0, time.UTC)
	repository, _ := NewFileTransformJobRepository(t.TempDir())
	request := TransformRequest{sourceID: "source-1", requirements: []TransformRequirement{TransformMediaTranscode}}
	job, _ := NewTransformJob("job-1", request, now)
	record, _ := NewTransformJobRecord(job)
	path := repository.recordPath("job-1")

	if err := os.WriteFile(path, []byte(`{"revision":0,"payload":"not-base64"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := repository.Load(context.Background(), "job-1"); !errors.Is(err, ErrInvalidTransformJobRecord) {
		t.Fatalf("malformed error=%v", err)
	}

	wrongJob, _ := NewTransformJob("job-2", request, now)
	wrongRecord, _ := NewTransformJobRecord(wrongJob)
	encoded, _ := encodeTransformJobFileEnvelope(wrongRecord)
	if err := os.WriteFile(path, encoded, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := repository.Load(context.Background(), "job-1"); !errors.Is(err, ErrInvalidTransformJobRecord) {
		t.Fatalf("wrong identity error=%v", err)
	}

	canonical, _ := encodeTransformJobFileEnvelope(record)
	if err := os.WriteFile(path, append(canonical, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := repository.Load(context.Background(), "job-1"); !errors.Is(err, ErrInvalidTransformJobRecord) {
		t.Fatalf("noncanonical error=%v", err)
	}
}

func TestFileTransformJobRepositoryNotFoundCancellationAndBounds(t *testing.T) {
	if _, err := NewFileTransformJobRepository("   "); !errors.Is(err, ErrInvalidTransformJobFileRepository) {
		t.Fatalf("blank root error=%v", err)
	}
	repository, _ := NewFileTransformJobRepository(t.TempDir())
	if _, found, err := repository.Load(context.Background(), "missing"); err != nil || found {
		t.Fatalf("found=%v err=%v", found, err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	request := TransformRequest{sourceID: "source-1", requirements: []TransformRequirement{TransformMediaTranscode}}
	job, _ := NewTransformJob("job-1", request, time.Now().UTC())
	if _, err := repository.Create(ctx, job); !errors.Is(err, context.Canceled) {
		t.Fatalf("create cancellation=%v", err)
	}
	if _, _, err := repository.Load(ctx, "job-1"); !errors.Is(err, context.Canceled) {
		t.Fatalf("load cancellation=%v", err)
	}

	path := repository.recordPath("oversized")
	if err := os.WriteFile(path, make([]byte, MaxTransformJobFileBytes+1), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := repository.Load(context.Background(), "oversized"); !errors.Is(err, ErrInvalidTransformJobRecord) {
		t.Fatalf("oversized error=%v", err)
	}
}

func TestFileTransformJobRepositoryRunsWorkerIterationAcrossDurableRevisions(t *testing.T) {
	now := time.Date(2026, 9, 7, 22, 0, 0, 0, time.UTC)
	repository, _ := NewFileTransformJobRepository(t.TempDir())
	request := TransformRequest{sourceID: "source-1", requirements: []TransformRequirement{TransformMediaTranscode}}
	job, _ := NewTransformJob("job-worker", request, now)
	if _, err := InitializeStoredTransformJob(context.Background(), repository, job); err != nil {
		t.Fatal(err)
	}
	resolver := &fakeFFmpegExecutionResolver{result: FFmpegExecutionRequest{
		Transform:  request,
		InputPath:  "/srv/video/input.mkv",
		OutputPath: "/srv/video/output.mp4",
	}}
	executor := &fakeTransformExecutionExecutor{}
	clock := &sequenceTransformWorkerClock{times: []time.Time{now.Add(time.Second), now.Add(2 * time.Second)}}
	iteration, err := RunStoredTransformWorkerIteration(
		context.Background(),
		repository,
		resolver,
		executor,
		clock,
		"job-worker",
		"worker-1",
		WorkerCapabilities{MediaTranscode: true},
		time.Minute,
	)
	if err != nil || iteration.Completed.Revision() != 2 || iteration.Outcome != TransformJobOutcomeSucceeded {
		t.Fatalf("iteration=%+v err=%v", iteration, err)
	}
	loaded, found, err := repository.Load(context.Background(), "job-worker")
	if err != nil || !found || loaded.Revision() != 2 {
		t.Fatalf("loaded=%+v found=%v err=%v", loaded, found, err)
	}
	stored, err := loaded.Restore()
	if err != nil || stored.State() != TransformJobSucceeded {
		t.Fatalf("stored=%+v err=%v", stored, err)
	}
}

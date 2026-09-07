package playback

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestClaimStoredTransformJobPersistsRunningRevisionBeforeReturningLease(t *testing.T) {
	now := time.Date(2026, 9, 7, 20, 10, 0, 0, time.UTC)
	job := testTransformJob(t, "job-claim-1", now)
	record, err := NewTransformJobRecord(job)
	if err != nil {
		t.Fatal(err)
	}
	repository := &fakeTransformJobRepository{record: record, found: true}

	saved, lease, err := ClaimStoredTransformJob(
		context.Background(),
		repository,
		"job-claim-1",
		"worker-1",
		WorkerCapabilities{MediaTranscode: true},
		now.Add(time.Second),
		5*time.Minute,
	)
	if err != nil || repository.saves != 1 || saved.Revision() != 1 {
		t.Fatalf("saved=%+v saves=%d err=%v", saved, repository.saves, err)
	}
	stored, err := saved.Restore()
	if err != nil || stored.State() != TransformJobRunning || stored.Attempts() != 1 || !lease.Matches(stored, "worker-1") {
		t.Fatalf("stored=%+v lease=%+v err=%v", stored, lease, err)
	}
}

func TestClaimStoredTransformJobRejectsUnsupportedWorkerWithoutSaving(t *testing.T) {
	now := time.Date(2026, 9, 7, 20, 10, 0, 0, time.UTC)
	job := testTransformJob(t, "job-claim-1", now)
	record, err := NewTransformJobRecord(job)
	if err != nil {
		t.Fatal(err)
	}
	repository := &fakeTransformJobRepository{record: record, found: true}

	returned, lease, err := ClaimStoredTransformJob(
		context.Background(),
		repository,
		"job-claim-1",
		"worker-1",
		WorkerCapabilities{},
		now.Add(time.Second),
		time.Minute,
	)
	if !errors.Is(err, ErrUnsupportedTransformJobWorker) || repository.saves != 0 || returned.Revision() != 0 || lease.JobID() != "" {
		t.Fatalf("returned=%+v lease=%+v saves=%d err=%v", returned, lease, repository.saves, err)
	}
}

func TestClaimStoredTransformJobNeverLeaksLeaseWhenSaveFails(t *testing.T) {
	now := time.Date(2026, 9, 7, 20, 10, 0, 0, time.UTC)
	job := testTransformJob(t, "job-claim-1", now)
	record, err := NewTransformJobRecord(job)
	if err != nil {
		t.Fatal(err)
	}
	want := errors.New("save failed")
	repository := &fakeTransformJobRepository{record: record, found: true, saveErr: want}

	returned, lease, err := ClaimStoredTransformJob(
		context.Background(),
		repository,
		"job-claim-1",
		"worker-1",
		WorkerCapabilities{MediaTranscode: true},
		now.Add(time.Second),
		time.Minute,
	)
	if !errors.Is(err, want) || returned.Revision() != 0 || lease.JobID() != "" {
		t.Fatalf("returned=%+v lease=%+v err=%v", returned, lease, err)
	}
}

func TestClaimStoredTransformJobPropagatesNotFoundAndCanceledContext(t *testing.T) {
	repository := &fakeTransformJobRepository{}
	if _, lease, err := ClaimStoredTransformJob(
		context.Background(),
		repository,
		"job-claim-1",
		"worker-1",
		WorkerCapabilities{MediaTranscode: true},
		time.Now(),
		time.Minute,
	); !errors.Is(err, ErrTransformJobRecordNotFound) || lease.JobID() != "" {
		t.Fatalf("lease=%+v err=%v", lease, err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, lease, err := ClaimStoredTransformJob(
		ctx,
		repository,
		"job-claim-1",
		"worker-1",
		WorkerCapabilities{MediaTranscode: true},
		time.Now(),
		time.Minute,
	); !errors.Is(err, context.Canceled) || lease.JobID() != "" {
		t.Fatalf("lease=%+v err=%v", lease, err)
	}
}

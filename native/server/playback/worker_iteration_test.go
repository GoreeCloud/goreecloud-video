package playback

import (
	"context"
	"errors"
	"testing"
	"time"
)

type sequenceTransformWorkerClock struct {
	times []time.Time
	index int
}

func (c *sequenceTransformWorkerClock) Now() time.Time {
	if c == nil || c.index >= len(c.times) {
		return time.Time{}
	}
	value := c.times[c.index]
	c.index++
	return value
}

type fakeFFmpegExecutionResolver struct {
	result FFmpegExecutionRequest
	err    error
	calls  int
	seen   TransformRequest
	hook   func()
}

func (f *fakeFFmpegExecutionResolver) ResolveExecution(_ context.Context, request TransformRequest) (FFmpegExecutionRequest, error) {
	f.calls++
	f.seen = cloneTransformRequest(request)
	if f.hook != nil {
		f.hook()
	}
	return f.result, f.err
}

type fakeTransformExecutionExecutor struct {
	err   error
	calls int
	seen  FFmpegExecutionRequest
}

func (f *fakeTransformExecutionExecutor) Execute(_ context.Context, request FFmpegExecutionRequest) error {
	f.calls++
	f.seen = request
	return f.err
}

func TestRunStoredTransformWorkerIterationPersistsSuccess(t *testing.T) {
	started := time.Date(2026, 9, 7, 21, 45, 0, 0, time.UTC)
	request := TransformRequest{sourceID: "source-1", requirements: []TransformRequirement{TransformMediaTranscode}}
	repository := storedWorkerTestRepository(t, "job-1", request, started)
	resolver := &fakeFFmpegExecutionResolver{result: FFmpegExecutionRequest{
		Transform:  request,
		InputPath:  "/srv/goreecloud/video/input.mkv",
		OutputPath: "/srv/goreecloud/video/output.mp4",
	}}
	executor := &fakeTransformExecutionExecutor{}
	clock := &sequenceTransformWorkerClock{times: []time.Time{started.Add(time.Second), started.Add(2 * time.Second)}}

	iteration, err := RunStoredTransformWorkerIteration(
		context.Background(),
		repository,
		resolver,
		executor,
		clock,
		"job-1",
		"worker-1",
		WorkerCapabilities{MediaTranscode: true},
		time.Minute,
	)
	if err != nil {
		t.Fatal(err)
	}
	if repository.saves != 2 || resolver.calls != 1 || executor.calls != 1 || !iteration.Executed || iteration.Outcome != TransformJobOutcomeSucceeded || iteration.FailureCode != "" {
		t.Fatalf("saves=%d resolver=%d executor=%d iteration=%+v", repository.saves, resolver.calls, executor.calls, iteration)
	}
	if iteration.Claimed.Revision() != 1 || iteration.Completed.Revision() != 2 || resolver.seen.SourceID() != "source-1" {
		t.Fatalf("claimed=%d completed=%d seen=%q", iteration.Claimed.Revision(), iteration.Completed.Revision(), resolver.seen.SourceID())
	}
	stored, err := iteration.Completed.Restore()
	if err != nil || stored.State() != TransformJobSucceeded || stored.Attempts() != 1 {
		t.Fatalf("stored=%+v err=%v", stored, err)
	}
}

func TestRunStoredTransformWorkerIterationPersistsResolutionFailureWithoutExecuting(t *testing.T) {
	started := time.Date(2026, 9, 7, 21, 45, 0, 0, time.UTC)
	request := TransformRequest{sourceID: "source-1", requirements: []TransformRequirement{TransformMediaTranscode}}
	repository := storedWorkerTestRepository(t, "job-1", request, started)
	resolver := &fakeFFmpegExecutionResolver{err: errors.New("storage unavailable")}
	executor := &fakeTransformExecutionExecutor{}
	clock := &sequenceTransformWorkerClock{times: []time.Time{started.Add(time.Second), started.Add(2 * time.Second)}}

	iteration, err := RunStoredTransformWorkerIteration(
		context.Background(), repository, resolver, executor, clock, "job-1", "worker-1",
		WorkerCapabilities{MediaTranscode: true}, time.Minute,
	)
	if err != nil {
		t.Fatal(err)
	}
	if executor.calls != 0 || repository.saves != 2 || iteration.Executed || iteration.Outcome != TransformJobOutcomeFailed || iteration.FailureCode != TransformWorkerFailureResolution {
		t.Fatalf("executor=%d saves=%d iteration=%+v", executor.calls, repository.saves, iteration)
	}
	stored, _ := iteration.Completed.Restore()
	if stored.State() != TransformJobFailed || stored.FailureCode() != TransformWorkerFailureResolution {
		t.Fatalf("stored=%+v", stored)
	}
}

func TestRunStoredTransformWorkerIterationRejectsResolverTransformSubstitution(t *testing.T) {
	started := time.Date(2026, 9, 7, 21, 45, 0, 0, time.UTC)
	request := TransformRequest{sourceID: "source-1", requirements: []TransformRequirement{TransformMediaTranscode}}
	repository := storedWorkerTestRepository(t, "job-1", request, started)
	resolver := &fakeFFmpegExecutionResolver{result: FFmpegExecutionRequest{
		Transform: TransformRequest{sourceID: "source-2", requirements: []TransformRequirement{TransformMediaTranscode}},
		InputPath: "/srv/goreecloud/video/input.mkv", OutputPath: "/srv/goreecloud/video/output.mp4",
	}}
	executor := &fakeTransformExecutionExecutor{}
	clock := &sequenceTransformWorkerClock{times: []time.Time{started.Add(time.Second), started.Add(2 * time.Second)}}

	iteration, err := RunStoredTransformWorkerIteration(
		context.Background(), repository, resolver, executor, clock, "job-1", "worker-1",
		WorkerCapabilities{MediaTranscode: true}, time.Minute,
	)
	if err != nil || executor.calls != 0 || iteration.FailureCode != TransformWorkerFailureResolution {
		t.Fatalf("executor=%d iteration=%+v err=%v", executor.calls, iteration, err)
	}
}

func TestRunStoredTransformWorkerIterationPersistsExecutionFailure(t *testing.T) {
	started := time.Date(2026, 9, 7, 21, 45, 0, 0, time.UTC)
	request := TransformRequest{sourceID: "source-1", requirements: []TransformRequirement{TransformMediaTranscode}}
	repository := storedWorkerTestRepository(t, "job-1", request, started)
	resolver := &fakeFFmpegExecutionResolver{result: FFmpegExecutionRequest{
		Transform: request, InputPath: "/srv/goreecloud/video/input.mkv", OutputPath: "/srv/goreecloud/video/output.mp4",
	}}
	executor := &fakeTransformExecutionExecutor{err: ErrFFmpegExecutionFailed}
	clock := &sequenceTransformWorkerClock{times: []time.Time{started.Add(time.Second), started.Add(2 * time.Second)}}

	iteration, err := RunStoredTransformWorkerIteration(
		context.Background(), repository, resolver, executor, clock, "job-1", "worker-1",
		WorkerCapabilities{MediaTranscode: true}, time.Minute,
	)
	if err != nil || executor.calls != 1 || !iteration.Executed || iteration.FailureCode != TransformWorkerFailureExecution {
		t.Fatalf("executor=%d iteration=%+v err=%v", executor.calls, iteration, err)
	}
	stored, _ := iteration.Completed.Restore()
	if stored.State() != TransformJobFailed || stored.FailureCode() != TransformWorkerFailureExecution {
		t.Fatalf("stored=%+v", stored)
	}
}

func TestRunStoredTransformWorkerIterationLeavesClaimForLeaseRecoveryWhenContextCancels(t *testing.T) {
	started := time.Date(2026, 9, 7, 21, 45, 0, 0, time.UTC)
	request := TransformRequest{sourceID: "source-1", requirements: []TransformRequirement{TransformMediaTranscode}}
	repository := storedWorkerTestRepository(t, "job-1", request, started)
	ctx, cancel := context.WithCancel(context.Background())
	resolver := &fakeFFmpegExecutionResolver{
		result: FFmpegExecutionRequest{Transform: request, InputPath: "/srv/video/in.mkv", OutputPath: "/srv/video/out.mp4"},
		hook:   cancel,
	}
	executor := &fakeTransformExecutionExecutor{}
	clock := &sequenceTransformWorkerClock{times: []time.Time{started.Add(time.Second)}}

	iteration, err := RunStoredTransformWorkerIteration(
		ctx, repository, resolver, executor, clock, "job-1", "worker-1",
		WorkerCapabilities{MediaTranscode: true}, time.Minute,
	)
	if !errors.Is(err, context.Canceled) || repository.saves != 1 || executor.calls != 0 || iteration.Claimed.Revision() != 1 || iteration.Completed.Revision() != 0 {
		t.Fatalf("saves=%d executor=%d iteration=%+v err=%v", repository.saves, executor.calls, iteration, err)
	}
	stored, _ := repository.record.Restore()
	if stored.State() != TransformJobRunning {
		t.Fatalf("stored=%+v", stored)
	}
}

func TestRunStoredTransformWorkerIterationFailsClosedOnExpiredLeaseAtCompletion(t *testing.T) {
	started := time.Date(2026, 9, 7, 21, 45, 0, 0, time.UTC)
	request := TransformRequest{sourceID: "source-1", requirements: []TransformRequirement{TransformMediaTranscode}}
	repository := storedWorkerTestRepository(t, "job-1", request, started)
	resolver := &fakeFFmpegExecutionResolver{result: FFmpegExecutionRequest{
		Transform: request, InputPath: "/srv/video/in.mkv", OutputPath: "/srv/video/out.mp4",
	}}
	executor := &fakeTransformExecutionExecutor{}
	clock := &sequenceTransformWorkerClock{times: []time.Time{started.Add(time.Second), started.Add(3 * time.Second)}}

	iteration, err := RunStoredTransformWorkerIteration(
		context.Background(), repository, resolver, executor, clock, "job-1", "worker-1",
		WorkerCapabilities{MediaTranscode: true}, time.Second,
	)
	if !errors.Is(err, ErrExpiredTransformJobLease) || repository.saves != 1 || executor.calls != 1 || !iteration.Executed {
		t.Fatalf("saves=%d executor=%d iteration=%+v err=%v", repository.saves, executor.calls, iteration, err)
	}
}

func TestRunStoredTransformWorkerIterationValidatesDependenciesAndCapabilities(t *testing.T) {
	started := time.Date(2026, 9, 7, 21, 45, 0, 0, time.UTC)
	request := TransformRequest{sourceID: "source-1", requirements: []TransformRequirement{TransformMediaTranscode}}
	repository := storedWorkerTestRepository(t, "job-1", request, started)
	resolver := &fakeFFmpegExecutionResolver{}
	executor := &fakeTransformExecutionExecutor{}
	clock := &sequenceTransformWorkerClock{times: []time.Time{started.Add(time.Second)}}

	if _, err := RunStoredTransformWorkerIteration(context.Background(), repository, nil, executor, clock, "job-1", "worker-1", WorkerCapabilities{MediaTranscode: true}, time.Minute); !errors.Is(err, ErrInvalidTransformWorkerIteration) {
		t.Fatalf("nil resolver error=%v", err)
	}
	if _, err := RunStoredTransformWorkerIteration(context.Background(), repository, resolver, executor, clock, "job-1", "worker-1", WorkerCapabilities{}, time.Minute); !errors.Is(err, ErrUnsupportedTransformJobWorker) {
		t.Fatalf("unsupported worker error=%v", err)
	}
}

func storedWorkerTestRepository(t *testing.T, id string, request TransformRequest, now time.Time) *fakeTransformJobRepository {
	t.Helper()
	job, err := NewTransformJob(id, request, now)
	if err != nil {
		t.Fatal(err)
	}
	record, err := NewTransformJobRecord(job)
	if err != nil {
		t.Fatal(err)
	}
	return &fakeTransformJobRepository{record: record, found: true}
}

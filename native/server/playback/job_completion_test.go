package playback

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestCompleteStoredTransformJobPersistsSuccess(t *testing.T) {
	now := time.Date(2026, 9, 7, 20, 30, 0, 0, time.UTC)
	repository, lease := claimedTransformJob(t, now, 10*time.Minute)

	record, err := CompleteStoredTransformJob(
		context.Background(),
		repository,
		"job-complete-1",
		"worker-1",
		lease,
		TransformJobOutcomeSucceeded,
		now.Add(2*time.Second),
		"",
	)
	if err != nil || record.Revision() != 2 {
		t.Fatalf("record=%+v err=%v", record, err)
	}
	stored, err := record.Restore()
	if err != nil || stored.State() != TransformJobSucceeded || stored.Attempts() != 1 {
		t.Fatalf("stored=%+v err=%v", stored, err)
	}
}

func TestCompleteStoredTransformJobPersistsCanonicalFailure(t *testing.T) {
	now := time.Date(2026, 9, 7, 20, 30, 0, 0, time.UTC)
	repository, lease := claimedTransformJob(t, now, 10*time.Minute)

	record, err := CompleteStoredTransformJob(
		context.Background(),
		repository,
		"job-complete-1",
		"worker-1",
		lease,
		TransformJobOutcomeFailed,
		now.Add(2*time.Second),
		"encoder-exit",
	)
	if err != nil {
		t.Fatal(err)
	}
	stored, err := record.Restore()
	if err != nil || stored.State() != TransformJobFailed || stored.FailureCode() != "encoder-exit" {
		t.Fatalf("stored=%+v err=%v", stored, err)
	}

	repository, lease = claimedTransformJob(t, now.Add(time.Hour), 10*time.Minute)
	if _, err := CompleteStoredTransformJob(
		context.Background(),
		repository,
		"job-complete-1",
		"worker-1",
		lease,
		TransformJobOutcomeFailed,
		now.Add(time.Hour+2*time.Second),
		" encoder-exit ",
	); !errors.Is(err, ErrInvalidTransformJobCompletion) {
		t.Fatalf("error=%v", err)
	}
}

func TestCompleteStoredTransformJobRejectsExpiredOrWrongWorkerLease(t *testing.T) {
	now := time.Date(2026, 9, 7, 20, 30, 0, 0, time.UTC)
	repository, lease := claimedTransformJob(t, now, time.Minute)
	before := repository.saves

	if _, err := CompleteStoredTransformJob(
		context.Background(),
		repository,
		"job-complete-1",
		"worker-1",
		lease,
		TransformJobOutcomeSucceeded,
		now.Add(time.Minute+time.Second),
		"",
	); !errors.Is(err, ErrExpiredTransformJobLease) || repository.saves != before {
		t.Fatalf("error=%v saves=%d", err, repository.saves)
	}

	if _, err := CompleteStoredTransformJob(
		context.Background(),
		repository,
		"job-complete-1",
		"worker-2",
		lease,
		TransformJobOutcomeSucceeded,
		now.Add(30*time.Second),
		"",
	); !errors.Is(err, ErrInvalidTransformJobCompletion) || repository.saves != before {
		t.Fatalf("error=%v saves=%d", err, repository.saves)
	}
}

func TestOldLeaseCannotCompleteLaterRetryAttempt(t *testing.T) {
	now := time.Date(2026, 9, 7, 20, 30, 0, 0, time.UTC)
	repository, oldLease := claimedTransformJob(t, now, 30*time.Minute)

	if _, err := MutateStoredTransformJob(context.Background(), repository, "job-complete-1", func(current TransformJob) (TransformJob, error) {
		return current.Transition(TransformJobFailed, now.Add(2*time.Second), "worker-failed")
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := MutateStoredTransformJob(context.Background(), repository, "job-complete-1", func(current TransformJob) (TransformJob, error) {
		return current.Transition(TransformJobQueued, now.Add(3*time.Second), "")
	}); err != nil {
		t.Fatal(err)
	}
	_, newLease, err := ClaimStoredTransformJob(
		context.Background(),
		repository,
		"job-complete-1",
		"worker-1",
		WorkerCapabilities{MediaTranscode: true},
		now.Add(4*time.Second),
		30*time.Minute,
	)
	if err != nil {
		t.Fatal(err)
	}
	if oldLease.Attempt() != 1 || newLease.Attempt() != 2 {
		t.Fatalf("old attempt=%d new attempt=%d", oldLease.Attempt(), newLease.Attempt())
	}

	before := repository.saves
	if _, err := CompleteStoredTransformJob(
		context.Background(),
		repository,
		"job-complete-1",
		"worker-1",
		oldLease,
		TransformJobOutcomeSucceeded,
		now.Add(5*time.Second),
		"",
	); !errors.Is(err, ErrInvalidTransformJobCompletion) || repository.saves != before {
		t.Fatalf("error=%v saves=%d", err, repository.saves)
	}

	if _, err := CompleteStoredTransformJob(
		context.Background(),
		repository,
		"job-complete-1",
		"worker-1",
		newLease,
		TransformJobOutcomeSucceeded,
		now.Add(5*time.Second),
		"",
	); err != nil {
		t.Fatal(err)
	}
}

func TestCompleteStoredTransformJobRejectsInvalidOutcomeAndSaveFailure(t *testing.T) {
	now := time.Date(2026, 9, 7, 20, 30, 0, 0, time.UTC)
	repository, lease := claimedTransformJob(t, now, 10*time.Minute)
	before := repository.saves
	if _, err := CompleteStoredTransformJob(
		context.Background(),
		repository,
		"job-complete-1",
		"worker-1",
		lease,
		TransformJobOutcome("unknown"),
		now.Add(2*time.Second),
		"",
	); !errors.Is(err, ErrInvalidTransformJobCompletion) || repository.saves != before {
		t.Fatalf("error=%v saves=%d", err, repository.saves)
	}

	want := errors.New("save failed")
	repository.saveErr = want
	if _, err := CompleteStoredTransformJob(
		context.Background(),
		repository,
		"job-complete-1",
		"worker-1",
		lease,
		TransformJobOutcomeCanceled,
		now.Add(2*time.Second),
		"",
	); !errors.Is(err, want) {
		t.Fatalf("error=%v", err)
	}
}

func claimedTransformJob(t *testing.T, now time.Time, ttl time.Duration) (*fakeTransformJobRepository, TransformJobLease) {
	t.Helper()
	job := testTransformJob(t, "job-complete-1", now)
	record, err := NewTransformJobRecord(job)
	if err != nil {
		t.Fatal(err)
	}
	repository := &fakeTransformJobRepository{record: record, found: true}
	_, lease, err := ClaimStoredTransformJob(
		context.Background(),
		repository,
		"job-complete-1",
		"worker-1",
		WorkerCapabilities{MediaTranscode: true},
		now.Add(time.Second),
		ttl,
	)
	if err != nil {
		t.Fatal(err)
	}
	return repository, lease
}

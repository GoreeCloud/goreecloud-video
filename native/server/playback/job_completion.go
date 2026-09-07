package playback

import (
	"context"
	"errors"
	"strings"
	"time"
)

type TransformJobOutcome string

const (
	TransformJobOutcomeSucceeded TransformJobOutcome = "succeeded"
	TransformJobOutcomeFailed    TransformJobOutcome = "failed"
	TransformJobOutcomeCanceled  TransformJobOutcome = "canceled"
)

var (
	ErrInvalidTransformJobCompletion = errors.New("invalid transform job completion")
	ErrExpiredTransformJobLease      = errors.New("transform job lease expired")
)

// CompleteStoredTransformJob applies one terminal worker result only when the
// supplied lease still matches the exact persisted running attempt. The helper
// persists the canonical next job record through the existing optimistic
// repository boundary; it does not execute or inspect media work.
func CompleteStoredTransformJob(
	ctx context.Context,
	repository TransformJobRepository,
	id string,
	workerID string,
	lease TransformJobLease,
	outcome TransformJobOutcome,
	now time.Time,
	failureCode string,
) (TransformJobRecord, error) {
	if repository == nil || !validTransformJobID(id) || !validTransformWorkerID(workerID) || !validTransformJobLease(lease) || now.IsZero() {
		return TransformJobRecord{}, ErrInvalidTransformJobCompletion
	}
	if lease.JobID() != id || lease.WorkerID() != workerID {
		return TransformJobRecord{}, ErrInvalidTransformJobCompletion
	}
	expired, err := lease.Expired(now)
	if err != nil {
		return TransformJobRecord{}, ErrInvalidTransformJobCompletion
	}
	if expired {
		return TransformJobRecord{}, ErrExpiredTransformJobLease
	}

	canonicalFailureCode := strings.TrimSpace(failureCode)
	if canonicalFailureCode != failureCode {
		return TransformJobRecord{}, ErrInvalidTransformJobCompletion
	}
	failureCode = canonicalFailureCode
	if outcome == TransformJobOutcomeFailed {
		if failureCode == "" {
			return TransformJobRecord{}, ErrInvalidTransformJobCompletion
		}
	} else if failureCode != "" {
		return TransformJobRecord{}, ErrInvalidTransformJobCompletion
	}

	return MutateStoredTransformJob(ctx, repository, id, func(current TransformJob) (TransformJob, error) {
		if current.State() != TransformJobRunning || !lease.Matches(current, workerID) {
			return current, ErrInvalidTransformJobCompletion
		}
		switch outcome {
		case TransformJobOutcomeSucceeded:
			return current.Transition(TransformJobSucceeded, now, "")
		case TransformJobOutcomeFailed:
			return current.Transition(TransformJobFailed, now, failureCode)
		case TransformJobOutcomeCanceled:
			return current.Transition(TransformJobCanceled, now, "")
		default:
			return current, ErrInvalidTransformJobCompletion
		}
	})
}

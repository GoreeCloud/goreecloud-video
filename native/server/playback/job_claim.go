package playback

import (
	"context"
	"time"
)

// ClaimStoredTransformJob performs one optimistic persisted worker claim. The
// returned lease is exposed only after the repository has accepted and returned
// the exact canonical running-job revision requested by the playback core.
func ClaimStoredTransformJob(
	ctx context.Context,
	repository TransformJobRepository,
	id string,
	workerID string,
	capabilities WorkerCapabilities,
	now time.Time,
	ttl time.Duration,
) (TransformJobRecord, TransformJobLease, error) {
	var claimedLease TransformJobLease
	record, err := MutateStoredTransformJob(
		ctx,
		repository,
		id,
		func(current TransformJob) (TransformJob, error) {
			running, lease, err := ClaimTransformJob(current, workerID, capabilities, now, ttl)
			if err != nil {
				return current, err
			}
			claimedLease = lease
			return running, nil
		},
	)
	if err != nil {
		return record, TransformJobLease{}, err
	}

	stored, err := record.Restore()
	if err != nil || stored.State() != TransformJobRunning || !claimedLease.Matches(stored, workerID) {
		return TransformJobRecord{}, TransformJobLease{}, ErrInvalidTransformJobRepositoryResult
	}
	return record, claimedLease, nil
}

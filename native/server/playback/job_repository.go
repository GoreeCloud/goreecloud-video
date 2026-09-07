package playback

import (
	"bytes"
	"context"
	"errors"
)

var (
	ErrInvalidTransformJobRepository       = errors.New("invalid transform job repository")
	ErrTransformJobRecordNotFound          = errors.New("transform job record not found")
	ErrInvalidTransformJobMutation         = errors.New("invalid transform job mutation")
	ErrInvalidTransformJobRepositoryResult = errors.New("invalid transform job repository result")
)

// TransformJobRepository is the storage-neutral persistence boundary for
// worker-neutral transform jobs. Implementations choose the actual database or
// queue storage and must honor TransformJobRecord revision semantics.
type TransformJobRepository interface {
	Load(ctx context.Context, id string) (TransformJobRecord, bool, error)
	Create(ctx context.Context, job TransformJob) (TransformJobRecord, error)
	Save(ctx context.Context, expected TransformJobRevision, job TransformJob) (TransformJobRecord, error)
}

// InitializeStoredTransformJob validates both the requested initial job and the
// repository's returned revision/payload before accepting persisted state.
func InitializeStoredTransformJob(
	ctx context.Context,
	repository TransformJobRepository,
	job TransformJob,
) (TransformJobRecord, error) {
	if repository == nil || !validTransformJob(job) {
		return TransformJobRecord{}, ErrInvalidTransformJobRepository
	}
	if err := ctx.Err(); err != nil {
		return TransformJobRecord{}, err
	}
	expected, err := NewTransformJobRecord(job)
	if err != nil {
		return TransformJobRecord{}, err
	}
	created, err := repository.Create(ctx, job)
	if err != nil {
		return TransformJobRecord{}, err
	}
	if !sameTransformJobRecord(created, expected) {
		return TransformJobRecord{}, ErrInvalidTransformJobRepositoryResult
	}
	return created, nil
}

// MutateStoredTransformJob performs one optimistic load-mutate-save cycle and
// verifies that the adapter returned the exact canonical next record requested
// by the core. The callback cannot change immutable job identity.
func MutateStoredTransformJob(
	ctx context.Context,
	repository TransformJobRepository,
	id string,
	mutate func(TransformJob) (TransformJob, error),
) (TransformJobRecord, error) {
	if repository == nil || mutate == nil || !validTransformJobID(id) {
		return TransformJobRecord{}, ErrInvalidTransformJobMutation
	}
	if err := ctx.Err(); err != nil {
		return TransformJobRecord{}, err
	}

	record, found, err := repository.Load(ctx, id)
	if err != nil {
		return TransformJobRecord{}, err
	}
	if !found {
		return TransformJobRecord{}, ErrTransformJobRecordNotFound
	}
	current, err := record.Restore()
	if err != nil || current.ID() != id {
		return TransformJobRecord{}, ErrInvalidTransformJobRepositoryResult
	}

	nextJob, err := mutate(current)
	if err != nil {
		return record, err
	}
	expected, err := record.Update(record.Revision(), nextJob)
	if err != nil {
		return record, err
	}
	if err := ctx.Err(); err != nil {
		return record, err
	}

	saved, err := repository.Save(ctx, record.Revision(), nextJob)
	if err != nil {
		return record, err
	}
	if !sameTransformJobRecord(saved, expected) {
		return record, ErrInvalidTransformJobRepositoryResult
	}
	return saved, nil
}

func sameTransformJobRecord(left, right TransformJobRecord) bool {
	return left.Revision() == right.Revision() && bytes.Equal(left.Payload(), right.Payload())
}

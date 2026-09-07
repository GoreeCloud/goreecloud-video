package playback

import (
	"context"
	"errors"
	"time"
)

var ErrInvalidTransformWorkerIteration = errors.New("invalid transform worker iteration")

const (
	TransformWorkerFailureResolution = "execution-resolution-failed"
	TransformWorkerFailureExecution  = "ffmpeg-execution-failed"
)

// FFmpegExecutionResolver translates one already-authorized opaque transform
// request into separately authorized filesystem paths. It is the storage
// authority boundary for a worker iteration; SourceID itself is never treated
// as a path or URL by the playback package.
type FFmpegExecutionResolver interface {
	ResolveExecution(ctx context.Context, request TransformRequest) (FFmpegExecutionRequest, error)
}

// TransformExecutionExecutor is implemented by FFmpegExecutor and keeps worker
// orchestration independent from process construction in tests.
type TransformExecutionExecutor interface {
	Execute(ctx context.Context, request FFmpegExecutionRequest) error
}

type TransformWorkerClock interface {
	Now() time.Time
}

type SystemTransformWorkerClock struct{}

func (SystemTransformWorkerClock) Now() time.Time { return time.Now().UTC() }

type TransformWorkerIteration struct {
	Claimed     TransformJobRecord
	Completed   TransformJobRecord
	Lease       TransformJobLease
	Outcome     TransformJobOutcome
	Executed    bool
	FailureCode string
}

// RunStoredTransformWorkerIteration performs exactly one persisted transform
// attempt: optimistic claim, separately authorized path resolution, one media
// execution, then lease-checked terminal persistence. It never polls, renews a
// lease, sleeps, or retries. If the caller context is canceled after the claim,
// the running job is intentionally left for lease-expiry/recovery handling
// rather than trying to persist terminal state through a canceled context.
func RunStoredTransformWorkerIteration(
	ctx context.Context,
	repository TransformJobRepository,
	resolver FFmpegExecutionResolver,
	executor TransformExecutionExecutor,
	clock TransformWorkerClock,
	jobID string,
	workerID string,
	capabilities WorkerCapabilities,
	leaseTTL time.Duration,
) (TransformWorkerIteration, error) {
	if repository == nil || resolver == nil || executor == nil || clock == nil || !validTransformJobID(jobID) || !validTransformWorkerID(workerID) {
		return TransformWorkerIteration{}, ErrInvalidTransformWorkerIteration
	}
	if err := ctx.Err(); err != nil {
		return TransformWorkerIteration{}, err
	}

	claimTime := clock.Now()
	if claimTime.IsZero() {
		return TransformWorkerIteration{}, ErrInvalidTransformWorkerIteration
	}
	claimed, lease, err := ClaimStoredTransformJob(
		ctx,
		repository,
		jobID,
		workerID,
		capabilities,
		claimTime,
		leaseTTL,
	)
	if err != nil {
		return TransformWorkerIteration{}, err
	}
	iteration := TransformWorkerIteration{Claimed: claimed, Lease: lease}
	job, err := claimed.Restore()
	if err != nil || job.State() != TransformJobRunning || !lease.Matches(job, workerID) {
		return iteration, ErrInvalidTransformWorkerIteration
	}

	execution, resolveErr := resolver.ResolveExecution(ctx, job.Request())
	if contextErr := ctx.Err(); contextErr != nil {
		return iteration, contextErr
	}
	if resolveErr != nil || !sameTransformRequest(execution.Transform, job.Request()) {
		return completeTransformWorkerFailure(
			ctx,
			repository,
			clock,
			iteration,
			jobID,
			workerID,
			lease,
			TransformWorkerFailureResolution,
		)
	}

	iteration.Executed = true
	executeErr := executor.Execute(ctx, execution)
	if contextErr := ctx.Err(); contextErr != nil {
		return iteration, contextErr
	}
	if executeErr != nil {
		return completeTransformWorkerFailure(
			ctx,
			repository,
			clock,
			iteration,
			jobID,
			workerID,
			lease,
			TransformWorkerFailureExecution,
		)
	}

	completionTime := clock.Now()
	if completionTime.IsZero() {
		return iteration, ErrInvalidTransformWorkerIteration
	}
	completed, err := CompleteStoredTransformJob(
		ctx,
		repository,
		jobID,
		workerID,
		lease,
		TransformJobOutcomeSucceeded,
		completionTime,
		"",
	)
	if err != nil {
		return iteration, err
	}
	iteration.Completed = completed
	iteration.Outcome = TransformJobOutcomeSucceeded
	return iteration, nil
}

func completeTransformWorkerFailure(
	ctx context.Context,
	repository TransformJobRepository,
	clock TransformWorkerClock,
	iteration TransformWorkerIteration,
	jobID string,
	workerID string,
	lease TransformJobLease,
	failureCode string,
) (TransformWorkerIteration, error) {
	completionTime := clock.Now()
	if completionTime.IsZero() {
		return iteration, ErrInvalidTransformWorkerIteration
	}
	completed, err := CompleteStoredTransformJob(
		ctx,
		repository,
		jobID,
		workerID,
		lease,
		TransformJobOutcomeFailed,
		completionTime,
		failureCode,
	)
	if err != nil {
		return iteration, err
	}
	iteration.Completed = completed
	iteration.Outcome = TransformJobOutcomeFailed
	iteration.FailureCode = failureCode
	return iteration, nil
}

func sameTransformRequest(left, right TransformRequest) bool {
	if left.SourceID() != right.SourceID() {
		return false
	}
	leftRequirements := left.Requirements()
	rightRequirements := right.Requirements()
	if len(leftRequirements) != len(rightRequirements) {
		return false
	}
	for index := range leftRequirements {
		if leftRequirements[index] != rightRequirements[index] {
			return false
		}
	}
	return true
}

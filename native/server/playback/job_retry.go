package playback

import (
	"errors"
	"time"
)

var ErrInvalidTransformRetryPolicy = errors.New("invalid transform retry policy")

type TransformRetryPolicy struct {
	BaseDelay   time.Duration
	MaxDelay    time.Duration
	MaxAttempts int
}

type TransformRetryPlan struct {
	RetryAt   time.Time
	Delay     time.Duration
	Ready     bool
	Exhausted bool
}

// PlanTransformJobRetry computes the retry observation for one failed transform
// job. The delay is capped exponential backoff based on completed attempts. This
// function never changes job state, sleeps, schedules work, or claims a worker.
func PlanTransformJobRetry(job TransformJob, policy TransformRetryPolicy, now time.Time) (TransformRetryPlan, error) {
	if !validTransformJob(job) || job.State() != TransformJobFailed || !validTransformRetryPolicy(policy) {
		return TransformRetryPlan{}, ErrInvalidTransformRetryPolicy
	}
	if job.Attempts() >= policy.MaxAttempts {
		return TransformRetryPlan{Exhausted: true}, nil
	}
	if now.IsZero() {
		return TransformRetryPlan{}, ErrInvalidTransformRetryPolicy
	}

	delay := retryDelayForAttempt(policy, job.Attempts())
	retryAt := job.UpdatedAt().Add(delay)
	now = now.UTC()
	if now.Before(job.UpdatedAt()) {
		return TransformRetryPlan{}, ErrInvalidTransformRetryPolicy
	}
	return TransformRetryPlan{
		RetryAt: retryAt,
		Delay:   delay,
		Ready:   !now.Before(retryAt),
	}, nil
}

func validTransformRetryPolicy(policy TransformRetryPolicy) bool {
	return policy.BaseDelay > 0 &&
		policy.MaxDelay >= policy.BaseDelay &&
		policy.MaxAttempts > 0
}

func retryDelayForAttempt(policy TransformRetryPolicy, attempts int) time.Duration {
	delay := policy.BaseDelay
	for attempt := 1; attempt < attempts; attempt++ {
		if delay >= policy.MaxDelay || delay > policy.MaxDelay/2 {
			return policy.MaxDelay
		}
		delay *= 2
	}
	if delay > policy.MaxDelay {
		return policy.MaxDelay
	}
	return delay
}

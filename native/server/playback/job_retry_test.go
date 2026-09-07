package playback

import (
	"errors"
	"testing"
	"time"
)

func TestPlanTransformJobRetryUsesCappedExponentialBackoff(t *testing.T) {
	now := time.Date(2026, 9, 7, 20, 20, 0, 0, time.UTC)
	job := failedTransformAttempt(t, now, 1)
	policy := TransformRetryPolicy{BaseDelay: 10 * time.Second, MaxDelay: time.Minute, MaxAttempts: 4}

	plan, err := PlanTransformJobRetry(job, policy, job.UpdatedAt().Add(5*time.Second))
	if err != nil || plan.Delay != 10*time.Second || plan.Ready || plan.Exhausted || !plan.RetryAt.Equal(job.UpdatedAt().Add(10*time.Second)) {
		t.Fatalf("plan=%+v err=%v", plan, err)
	}
	plan, err = PlanTransformJobRetry(job, policy, job.UpdatedAt().Add(10*time.Second))
	if err != nil || !plan.Ready {
		t.Fatalf("ready plan=%+v err=%v", plan, err)
	}

	job = failedTransformAttempt(t, now.Add(time.Hour), 2)
	plan, err = PlanTransformJobRetry(job, policy, job.UpdatedAt())
	if err != nil || plan.Delay != 20*time.Second {
		t.Fatalf("second-attempt plan=%+v err=%v", plan, err)
	}

	job = failedTransformAttempt(t, now.Add(2*time.Hour), 3)
	policy.MaxDelay = 25 * time.Second
	plan, err = PlanTransformJobRetry(job, policy, job.UpdatedAt())
	if err != nil || plan.Delay != 25*time.Second {
		t.Fatalf("capped plan=%+v err=%v", plan, err)
	}
}

func TestPlanTransformJobRetryReportsExhaustionWithoutClock(t *testing.T) {
	now := time.Date(2026, 9, 7, 20, 20, 0, 0, time.UTC)
	job := failedTransformAttempt(t, now, 2)
	plan, err := PlanTransformJobRetry(
		job,
		TransformRetryPolicy{BaseDelay: time.Second, MaxDelay: time.Minute, MaxAttempts: 2},
		time.Time{},
	)
	if err != nil || !plan.Exhausted || plan.Ready || !plan.RetryAt.IsZero() || plan.Delay != 0 {
		t.Fatalf("plan=%+v err=%v", plan, err)
	}
}

func TestPlanTransformJobRetryRejectsInvalidPolicyStateAndClock(t *testing.T) {
	now := time.Date(2026, 9, 7, 20, 20, 0, 0, time.UTC)
	queued := testTransformJob(t, "job-retry-invalid", now)
	policy := TransformRetryPolicy{BaseDelay: time.Second, MaxDelay: time.Minute, MaxAttempts: 3}
	if _, err := PlanTransformJobRetry(queued, policy, now); !errors.Is(err, ErrInvalidTransformRetryPolicy) {
		t.Fatalf("queued error=%v", err)
	}

	failed := failedTransformAttempt(t, now.Add(time.Hour), 1)
	invalidPolicies := []TransformRetryPolicy{
		{},
		{BaseDelay: 0, MaxDelay: time.Minute, MaxAttempts: 3},
		{BaseDelay: time.Minute, MaxDelay: time.Second, MaxAttempts: 3},
		{BaseDelay: time.Second, MaxDelay: time.Minute, MaxAttempts: 0},
	}
	for _, invalid := range invalidPolicies {
		if _, err := PlanTransformJobRetry(failed, invalid, failed.UpdatedAt()); !errors.Is(err, ErrInvalidTransformRetryPolicy) {
			t.Fatalf("policy=%+v error=%v", invalid, err)
		}
	}
	if _, err := PlanTransformJobRetry(failed, policy, time.Time{}); !errors.Is(err, ErrInvalidTransformRetryPolicy) {
		t.Fatalf("missing clock error=%v", err)
	}
	if _, err := PlanTransformJobRetry(failed, policy, failed.UpdatedAt().Add(-time.Nanosecond)); !errors.Is(err, ErrInvalidTransformRetryPolicy) {
		t.Fatalf("backward clock error=%v", err)
	}
}

func failedTransformAttempt(t *testing.T, start time.Time, attempts int) TransformJob {
	t.Helper()
	job := testTransformJob(t, "job-retry", start)
	clock := start
	for attempt := 1; attempt <= attempts; attempt++ {
		clock = clock.Add(time.Second)
		running, err := job.Transition(TransformJobRunning, clock, "")
		if err != nil {
			t.Fatal(err)
		}
		clock = clock.Add(time.Second)
		job, err = running.Transition(TransformJobFailed, clock, "worker-failed")
		if err != nil {
			t.Fatal(err)
		}
		if attempt < attempts {
			clock = clock.Add(time.Second)
			job, err = job.Transition(TransformJobQueued, clock, "")
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	return job
}

package playback

import (
	"errors"
	"strings"
	"time"
)

const (
	MaxTransformWorkerIDLength = 128
	MaxTransformLeaseDuration  = 24 * time.Hour
)

var (
	ErrInvalidTransformJobLease      = errors.New("invalid transform job lease")
	ErrUnsupportedTransformJobWorker = errors.New("worker cannot satisfy transform job requirements")
)

// TransformJobLease is a scheduler-neutral claim tying one queued transform job
// to one capability-compatible worker for a bounded time window. It does not
// dispatch processes, reserve hardware, persist itself, or execute media work.
type TransformJobLease struct {
	jobID      string
	workerID   string
	acquiredAt time.Time
	expiresAt  time.Time
}

func (l TransformJobLease) JobID() string      { return l.jobID }
func (l TransformJobLease) WorkerID() string   { return l.workerID }
func (l TransformJobLease) AcquiredAt() time.Time { return l.acquiredAt }
func (l TransformJobLease) ExpiresAt() time.Time  { return l.expiresAt }

// ClaimTransformJob validates worker capability against the job's immutable
// transform requirements, transitions the job from queued to running exactly
// once, and returns a bounded lease describing that claim.
func ClaimTransformJob(
	job TransformJob,
	workerID string,
	capabilities WorkerCapabilities,
	now time.Time,
	ttl time.Duration,
) (TransformJob, TransformJobLease, error) {
	if !validTransformJob(job) || job.State() != TransformJobQueued || !validTransformWorkerID(workerID) || !validTransformLeaseWindow(now, ttl) {
		return TransformJob{}, TransformJobLease{}, ErrInvalidTransformJobLease
	}

	decision, err := EvaluateTransformWorker(
		TransformPlan{Requirements: append([]TransformRequirement(nil), job.request.requirements...)},
		capabilities,
	)
	if err != nil {
		return TransformJob{}, TransformJobLease{}, ErrInvalidTransformJobLease
	}
	if !decision.Supported {
		return TransformJob{}, TransformJobLease{}, ErrUnsupportedTransformJobWorker
	}

	now = now.UTC()
	running, err := job.Transition(TransformJobRunning, now, "")
	if err != nil {
		return TransformJob{}, TransformJobLease{}, ErrInvalidTransformJobLease
	}
	return running, TransformJobLease{
		jobID:      running.ID(),
		workerID:   workerID,
		acquiredAt: now,
		expiresAt:  now.Add(ttl),
	}, nil
}

func (l TransformJobLease) Expired(now time.Time) (bool, error) {
	if !validTransformJobLease(l) || now.IsZero() {
		return false, ErrInvalidTransformJobLease
	}
	now = now.UTC()
	if now.Before(l.acquiredAt) {
		return false, ErrInvalidTransformJobLease
	}
	return !now.Before(l.expiresAt), nil
}

// Renew extends a still-active lease from the supplied current time. The job
// and worker identities are immutable; an expired lease must be reacquired by
// a higher-level scheduler rather than silently revived here.
func (l TransformJobLease) Renew(now time.Time, ttl time.Duration) (TransformJobLease, error) {
	if !validTransformJobLease(l) || !validTransformLeaseWindow(now, ttl) {
		return TransformJobLease{}, ErrInvalidTransformJobLease
	}
	now = now.UTC()
	if now.Before(l.acquiredAt) || !now.Before(l.expiresAt) {
		return TransformJobLease{}, ErrInvalidTransformJobLease
	}
	l.expiresAt = now.Add(ttl)
	return l, nil
}

func (l TransformJobLease) Matches(job TransformJob, workerID string) bool {
	return validTransformJobLease(l) && validTransformJob(job) && l.jobID == job.ID() && l.workerID == workerID
}

func validTransformWorkerID(id string) bool {
	return id != "" && len(id) <= MaxTransformWorkerIDLength && strings.TrimSpace(id) == id
}

func validTransformLeaseWindow(now time.Time, ttl time.Duration) bool {
	return !now.IsZero() && ttl > 0 && ttl <= MaxTransformLeaseDuration
}

func validTransformJobLease(lease TransformJobLease) bool {
	return validTransformJobID(lease.jobID) &&
		validTransformWorkerID(lease.workerID) &&
		!lease.acquiredAt.IsZero() &&
		!lease.expiresAt.IsZero() &&
		lease.acquiredAt.Equal(lease.acquiredAt.UTC()) &&
		lease.expiresAt.Equal(lease.expiresAt.UTC()) &&
		lease.expiresAt.After(lease.acquiredAt) &&
		lease.expiresAt.Sub(lease.acquiredAt) <= MaxTransformLeaseDuration
}

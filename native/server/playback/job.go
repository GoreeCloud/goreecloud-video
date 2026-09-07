package playback

import (
	"errors"
	"strings"
	"time"
)

const MaxTransformJobIDLength = 128

type TransformJobState string

const (
	TransformJobQueued    TransformJobState = "queued"
	TransformJobRunning   TransformJobState = "running"
	TransformJobSucceeded TransformJobState = "succeeded"
	TransformJobFailed    TransformJobState = "failed"
	TransformJobCanceled  TransformJobState = "canceled"
)

var ErrInvalidTransformJob = errors.New("invalid transform job")

// TransformJob is a worker-neutral lifecycle record for one already-validated
// TransformRequest. It does not schedule or execute FFmpeg/media work.
type TransformJob struct {
	id          string
	request     TransformRequest
	state       TransformJobState
	updatedAt   time.Time
	attempts    int
	failureCode string
}

func NewTransformJob(id string, request TransformRequest, now time.Time) (TransformJob, error) {
	if !validTransformJobID(id) || !validTransformRequest(request) || now.IsZero() {
		return TransformJob{}, ErrInvalidTransformJob
	}
	return TransformJob{id: id, request: cloneTransformRequest(request), state: TransformJobQueued, updatedAt: now.UTC()}, nil
}

func (j TransformJob) ID() string                { return j.id }
func (j TransformJob) Request() TransformRequest { return cloneTransformRequest(j.request) }
func (j TransformJob) State() TransformJobState  { return j.state }
func (j TransformJob) UpdatedAt() time.Time      { return j.updatedAt }
func (j TransformJob) Attempts() int             { return j.attempts }
func (j TransformJob) FailureCode() string       { return j.failureCode }

func (j TransformJob) Transition(next TransformJobState, now time.Time, failureCode string) (TransformJob, error) {
	if !validTransformJob(j) || !allowedTransformJobTransition(j.state, next) || now.IsZero() {
		return TransformJob{}, ErrInvalidTransformJob
	}
	now = now.UTC()
	if now.Before(j.updatedAt) {
		return TransformJob{}, ErrInvalidTransformJob
	}
	failureCode = strings.TrimSpace(failureCode)
	if next == TransformJobFailed && failureCode == "" {
		return TransformJob{}, ErrInvalidTransformJob
	}
	if next != TransformJobFailed {
		failureCode = ""
	}
	if next == TransformJobRunning {
		j.attempts++
	}
	j.state = next
	j.updatedAt = now
	j.failureCode = failureCode
	return j, nil
}

func validTransformJobID(id string) bool {
	return id != "" && len(id) <= MaxTransformJobIDLength && strings.TrimSpace(id) == id
}

func validTransformRequest(request TransformRequest) bool {
	if request.sourceID == "" || len(request.sourceID) > MaxTransformSourceIDLength || strings.TrimSpace(request.sourceID) != request.sourceID || len(request.requirements) == 0 {
		return false
	}
	seen := make(map[TransformRequirement]struct{}, len(request.requirements))
	for _, requirement := range request.requirements {
		if _, duplicate := seen[requirement]; duplicate {
			return false
		}
		seen[requirement] = struct{}{}
		switch requirement {
		case TransformMediaTranscode, TransformSubtitleBurnIn, TransformHDRToneMap:
		default:
			return false
		}
	}
	return true
}

func cloneTransformRequest(request TransformRequest) TransformRequest {
	return TransformRequest{sourceID: request.sourceID, requirements: append([]TransformRequirement(nil), request.requirements...)}
}

func validTransformJob(job TransformJob) bool {
	if !validTransformJobID(job.id) || !validTransformRequest(job.request) || job.updatedAt.IsZero() || job.attempts < 0 {
		return false
	}
	if job.state == TransformJobFailed {
		return strings.TrimSpace(job.failureCode) != "" && job.failureCode == strings.TrimSpace(job.failureCode)
	}
	return job.failureCode == "" && (job.state == TransformJobQueued || job.state == TransformJobRunning || job.state == TransformJobSucceeded || job.state == TransformJobCanceled)
}

func allowedTransformJobTransition(current, next TransformJobState) bool {
	switch current {
	case TransformJobQueued:
		return next == TransformJobRunning || next == TransformJobCanceled
	case TransformJobRunning:
		return next == TransformJobSucceeded || next == TransformJobFailed || next == TransformJobCanceled
	case TransformJobFailed:
		return next == TransformJobQueued || next == TransformJobCanceled
	default:
		return false
	}
}

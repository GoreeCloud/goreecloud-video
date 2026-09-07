package playback

import (
	"errors"
	"strings"
)

const MaxTransformSourceIDLength = 256

var (
	ErrInvalidTransformSourceID = errors.New("invalid transform source id")
	ErrTransformNotRequired     = errors.New("playback session does not require transformation")
)

// TransformRequest is an opaque worker-facing request. SourceID identifies an
// already-authorized media source outside the policy package; it is not a path,
// URL, credential, or storage authority interpreted by this package.
type TransformRequest struct {
	sourceID     string
	requirements []TransformRequirement
}

func BuildTransformRequest(sourceID string, session SessionPlan) (TransformRequest, error) {
	if sourceID == "" || len(sourceID) > MaxTransformSourceIDLength || strings.TrimSpace(sourceID) != sourceID {
		return TransformRequest{}, ErrInvalidTransformSourceID
	}
	plan, err := PlanTransforms(session)
	if err != nil {
		return TransformRequest{}, err
	}
	if !plan.RequiresTranscoder() {
		return TransformRequest{}, ErrTransformNotRequired
	}
	return TransformRequest{
		sourceID:     sourceID,
		requirements: append([]TransformRequirement(nil), plan.Requirements...),
	}, nil
}

func (r TransformRequest) SourceID() string {
	return r.sourceID
}

func (r TransformRequest) Requirements() []TransformRequirement {
	return append([]TransformRequirement(nil), r.requirements...)
}

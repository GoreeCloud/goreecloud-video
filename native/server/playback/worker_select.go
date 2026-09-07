package playback

import (
	"errors"
	"strings"
)

const MaxTransformWorkerIDLength = 128

var ErrInvalidTransformWorker = errors.New("invalid transform worker")

type TransformWorkerCandidate struct {
	ID           string
	Capabilities WorkerCapabilities
	Available    bool
}

// SelectTransformWorker chooses the first available worker that can satisfy an
// already-built transform request. Candidate order is therefore the caller's
// explicit scheduling preference. This function does not reserve capacity,
// launch jobs, inspect processes, or execute media transformations.
func SelectTransformWorker(
	request TransformRequest,
	candidates []TransformWorkerCandidate,
) (TransformWorkerCandidate, bool, error) {
	if request.sourceID == "" || len(request.sourceID) > MaxTransformSourceIDLength || strings.TrimSpace(request.sourceID) != request.sourceID {
		return TransformWorkerCandidate{}, false, ErrInvalidTransformSourceID
	}
	if len(request.requirements) == 0 {
		return TransformWorkerCandidate{}, false, ErrTransformNotRequired
	}

	seen := make(map[string]struct{}, len(candidates))
	plan := TransformPlan{Requirements: append([]TransformRequirement(nil), request.requirements...)}
	for _, candidate := range candidates {
		if candidate.ID == "" || len(candidate.ID) > MaxTransformWorkerIDLength || strings.TrimSpace(candidate.ID) != candidate.ID {
			return TransformWorkerCandidate{}, false, ErrInvalidTransformWorker
		}
		if _, duplicate := seen[candidate.ID]; duplicate {
			return TransformWorkerCandidate{}, false, ErrInvalidTransformWorker
		}
		seen[candidate.ID] = struct{}{}

		decision, err := EvaluateTransformWorker(plan, candidate.Capabilities)
		if err != nil {
			return TransformWorkerCandidate{}, false, err
		}
		if candidate.Available && decision.Supported {
			return candidate, true, nil
		}
	}
	return TransformWorkerCandidate{}, false, nil
}

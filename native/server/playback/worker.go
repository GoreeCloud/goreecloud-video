package playback

import "errors"

type WorkerCapabilities struct {
	MediaTranscode bool
	SubtitleBurnIn bool
	HDRToneMap     bool
}

type WorkerDecision struct {
	Supported bool
	Missing   []TransformRequirement
}

// EvaluateTransformWorker determines whether one worker can satisfy every
// requirement in an already-validated transform plan. It does not schedule or
// execute a worker.
func EvaluateTransformWorker(plan TransformPlan, capabilities WorkerCapabilities) (WorkerDecision, error) {
	missing := make([]TransformRequirement, 0, len(plan.Requirements))
	seen := make(map[TransformRequirement]struct{}, len(plan.Requirements))
	for _, requirement := range plan.Requirements {
		if _, duplicate := seen[requirement]; duplicate {
			return WorkerDecision{}, errors.New("duplicate transform requirement")
		}
		seen[requirement] = struct{}{}

		supported := false
		switch requirement {
		case TransformMediaTranscode:
			supported = capabilities.MediaTranscode
		case TransformSubtitleBurnIn:
			supported = capabilities.SubtitleBurnIn
		case TransformHDRToneMap:
			supported = capabilities.HDRToneMap
		default:
			return WorkerDecision{}, errors.New("unknown transform requirement")
		}
		if !supported {
			missing = append(missing, requirement)
		}
	}
	return WorkerDecision{Supported: len(missing) == 0, Missing: missing}, nil
}

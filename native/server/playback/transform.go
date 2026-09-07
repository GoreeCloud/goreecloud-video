package playback

import "errors"

type TransformRequirement string

const (
	TransformMediaTranscode  TransformRequirement = "media-transcode"
	TransformSubtitleBurnIn TransformRequirement = "subtitle-burn-in"
	TransformHDRToneMap     TransformRequirement = "hdr-tone-map"
)

type TransformPlan struct {
	Requirements []TransformRequirement
}

func (p TransformPlan) RequiresTranscoder() bool {
	return len(p.Requirements) > 0
}

// PlanTransforms converts an already-decided playback session into a
// deterministic worker-facing transformation contract. It validates that the
// supplied SessionPlan is internally coherent but does not execute FFmpeg or
// any other media worker.
func PlanTransforms(session SessionPlan) (TransformPlan, error) {
	if session.Mode == ModeDenied {
		return TransformPlan{}, errors.New("denied playback session cannot produce a transform plan")
	}

	requirements := make([]TransformRequirement, 0, 3)
	if session.Media.Mode == ModeTranscode {
		requirements = append(requirements, TransformMediaTranscode)
	}
	if session.Subtitle.Mode == SubtitleBurnIn {
		requirements = append(requirements, TransformSubtitleBurnIn)
	}
	if session.HDR.Mode == HDRToneMap {
		requirements = append(requirements, TransformHDRToneMap)
	}

	requiresTransform := len(requirements) > 0
	if requiresTransform != session.RequiresTranscoding {
		return TransformPlan{}, errors.New("inconsistent session transcoding requirement")
	}
	if requiresTransform && session.Mode != ModeTranscode {
		return TransformPlan{}, errors.New("transforming session must use transcode mode")
	}
	if !requiresTransform && session.Mode == ModeTranscode {
		return TransformPlan{}, errors.New("transcode mode requires at least one transformation")
	}

	return TransformPlan{Requirements: requirements}, nil
}

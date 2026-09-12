package playback

import (
	"reflect"
	"testing"
)

func TestTransformPlanDirectSessionNeedsNoWorker(t *testing.T) {
	plan, err := PlanTransforms(SessionPlan{
		Mode:     ModeDirectPlay,
		Media:    Decision{Mode: ModeDirectPlay},
		Subtitle: SubtitleDecision{Mode: SubtitleDirect},
		HDR:      HDRDecision{Mode: HDROff},
	})
	if err != nil {
		t.Fatal(err)
	}
	if plan.RequiresTranscoder() || len(plan.Requirements) != 0 {
		t.Fatalf("unexpected transform plan: %+v", plan)
	}
}

func TestTransformPlanIncludesPrimaryMediaTranscode(t *testing.T) {
	plan, err := PlanTransforms(SessionPlan{
		Mode:                ModeTranscode,
		Media:               Decision{Mode: ModeTranscode},
		Subtitle:            SubtitleDecision{Mode: SubtitleOff},
		HDR:                 HDRDecision{Mode: HDROff},
		RequiresTranscoding: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(plan.Requirements, []TransformRequirement{TransformMediaTranscode}) {
		t.Fatalf("requirements = %v", plan.Requirements)
	}
}

func TestTransformPlanOrdersCombinedSubtitleAndHDRRequirements(t *testing.T) {
	plan, err := PlanTransforms(SessionPlan{
		Mode:                ModeTranscode,
		Media:               Decision{Mode: ModeDirectPlay},
		Subtitle:            SubtitleDecision{Mode: SubtitleBurnIn},
		HDR:                 HDRDecision{Mode: HDRToneMap},
		RequiresTranscoding: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []TransformRequirement{TransformSubtitleBurnIn, TransformHDRToneMap}
	if !reflect.DeepEqual(plan.Requirements, want) {
		t.Fatalf("requirements = %v, want %v", plan.Requirements, want)
	}
}

func TestTransformPlanRejectsDeniedSession(t *testing.T) {
	if _, err := PlanTransforms(SessionPlan{Mode: ModeDenied}); err == nil {
		t.Fatal("expected denied session to reject transform planning")
	}
}

func TestTransformPlanRejectsInconsistentSessionFlags(t *testing.T) {
	cases := []SessionPlan{
		{
			Mode:                ModeDirectPlay,
			Media:               Decision{Mode: ModeDirectPlay},
			Subtitle:            SubtitleDecision{Mode: SubtitleBurnIn},
			HDR:                 HDRDecision{Mode: HDROff},
			RequiresTranscoding: false,
		},
		{
			Mode:                ModeDirectPlay,
			Media:               Decision{Mode: ModeDirectPlay},
			Subtitle:            SubtitleDecision{Mode: SubtitleBurnIn},
			HDR:                 HDRDecision{Mode: HDROff},
			RequiresTranscoding: true,
		},
		{
			Mode:                ModeTranscode,
			Media:               Decision{Mode: ModeDirectPlay},
			Subtitle:            SubtitleDecision{Mode: SubtitleOff},
			HDR:                 HDRDecision{Mode: HDROff},
			RequiresTranscoding: false,
		},
	}
	for _, session := range cases {
		if _, err := PlanTransforms(session); err == nil {
			t.Fatalf("expected inconsistent session %+v to be rejected", session)
		}
	}
}

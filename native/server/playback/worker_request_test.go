package playback

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestBuildTransformRequestCapturesDeterministicRequirements(t *testing.T) {
	session := SessionPlan{
		Mode:                ModeTranscode,
		Media:               Decision{Mode: ModeTranscode},
		Subtitle:            SubtitleDecision{Mode: SubtitleBurnIn},
		HDR:                 HDRDecision{Mode: HDRToneMap},
		RequiresTranscoding: true,
	}
	request, err := BuildTransformRequest("source-1", session)
	if err != nil {
		t.Fatal(err)
	}
	want := []TransformRequirement{
		TransformMediaTranscode,
		TransformSubtitleBurnIn,
		TransformHDRToneMap,
	}
	if request.SourceID() != "source-1" || !reflect.DeepEqual(request.Requirements(), want) {
		t.Fatalf("unexpected request: source=%q requirements=%v", request.SourceID(), request.Requirements())
	}

	copyRequirements := request.Requirements()
	copyRequirements[0] = TransformHDRToneMap
	if !reflect.DeepEqual(request.Requirements(), want) {
		t.Fatal("returned requirements mutated request state")
	}
}

func TestBuildTransformRequestRejectsInvalidSourceIDs(t *testing.T) {
	session := SessionPlan{
		Mode:                ModeTranscode,
		Media:               Decision{Mode: ModeTranscode},
		RequiresTranscoding: true,
	}
	for _, sourceID := range []string{"", " source", strings.Repeat("x", MaxTransformSourceIDLength+1)} {
		if _, err := BuildTransformRequest(sourceID, session); !errors.Is(err, ErrInvalidTransformSourceID) {
			t.Fatalf("source id %q error = %v, want invalid source id", sourceID, err)
		}
	}
}

func TestBuildTransformRequestRejectsNoTransformSession(t *testing.T) {
	session := SessionPlan{Mode: ModeDirectPlay, Media: Decision{Mode: ModeDirectPlay}}
	if _, err := BuildTransformRequest("source", session); !errors.Is(err, ErrTransformNotRequired) {
		t.Fatalf("error = %v, want transform not required", err)
	}
}

func TestBuildTransformRequestRejectsDeniedSession(t *testing.T) {
	session := SessionPlan{Mode: ModeDenied, Media: Decision{Mode: ModeDenied}}
	if _, err := BuildTransformRequest("source", session); err == nil {
		t.Fatal("expected denied session to fail")
	}
}

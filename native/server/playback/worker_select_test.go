package playback

import (
	"errors"
	"strings"
	"testing"
)

func transformRequestForWorkerTests(t *testing.T) TransformRequest {
	t.Helper()
	request, err := BuildTransformRequest("media-1", SessionPlan{
		Mode:                ModeTranscode,
		Media:               Decision{Mode: ModeTranscode},
		Subtitle:            SubtitleDecision{Mode: SubtitleOff},
		HDR:                 HDRDecision{Mode: HDROff},
		RequiresTranscoding: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	return request
}

func TestSelectTransformWorkerUsesFirstAvailableCompatibleCandidate(t *testing.T) {
	request := transformRequestForWorkerTests(t)
	candidate, ok, err := SelectTransformWorker(request, []TransformWorkerCandidate{
		{ID: "busy", Capabilities: WorkerCapabilities{MediaTranscode: true}, Available: false},
		{ID: "incompatible", Capabilities: WorkerCapabilities{}, Available: true},
		{ID: "selected", Capabilities: WorkerCapabilities{MediaTranscode: true}, Available: true},
		{ID: "later", Capabilities: WorkerCapabilities{MediaTranscode: true}, Available: true},
	})
	if err != nil || !ok || candidate.ID != "selected" {
		t.Fatalf("candidate=%+v ok=%v err=%v", candidate, ok, err)
	}
}

func TestSelectTransformWorkerReturnsNoMatchWithoutError(t *testing.T) {
	request := transformRequestForWorkerTests(t)
	candidate, ok, err := SelectTransformWorker(request, []TransformWorkerCandidate{
		{ID: "worker-1", Capabilities: WorkerCapabilities{}, Available: true},
	})
	if err != nil || ok || candidate.ID != "" {
		t.Fatalf("candidate=%+v ok=%v err=%v", candidate, ok, err)
	}
}

func TestSelectTransformWorkerRejectsInvalidAndDuplicateIDs(t *testing.T) {
	request := transformRequestForWorkerTests(t)
	cases := [][]TransformWorkerCandidate{
		{{ID: " worker", Capabilities: WorkerCapabilities{MediaTranscode: true}, Available: true}},
		{{ID: strings.Repeat("w", MaxTransformWorkerIDLength+1), Capabilities: WorkerCapabilities{MediaTranscode: true}, Available: true}},
		{
			{ID: "worker", Capabilities: WorkerCapabilities{MediaTranscode: true}, Available: false},
			{ID: "worker", Capabilities: WorkerCapabilities{MediaTranscode: true}, Available: true},
		},
		{
			{ID: "worker", Capabilities: WorkerCapabilities{MediaTranscode: true}, Available: true},
			{ID: "worker", Capabilities: WorkerCapabilities{MediaTranscode: true}, Available: false},
		},
	}
	for i, candidates := range cases {
		if _, _, err := SelectTransformWorker(request, candidates); !errors.Is(err, ErrInvalidTransformWorker) {
			t.Fatalf("case %d error=%v", i, err)
		}
	}
}

func TestSelectTransformWorkerRejectsMalformedRequest(t *testing.T) {
	if _, _, err := SelectTransformWorker(TransformRequest{}, nil); !errors.Is(err, ErrInvalidTransformSourceID) {
		t.Fatalf("error=%v", err)
	}
}

func TestSelectTransformWorkerRejectsUnknownRequirement(t *testing.T) {
	request := transformRequestForWorkerTests(t)
	request.requirements = []TransformRequirement{"future-transform"}
	_, _, err := SelectTransformWorker(request, []TransformWorkerCandidate{
		{ID: "worker", Available: true},
	})
	if err == nil {
		t.Fatal("expected unknown transform requirement to fail closed")
	}
}

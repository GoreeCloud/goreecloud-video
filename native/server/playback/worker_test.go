package playback

import "testing"

func TestEvaluateTransformWorker(t *testing.T) {
	plan := TransformPlan{Requirements: []TransformRequirement{TransformMediaTranscode, TransformSubtitleBurnIn, TransformHDRToneMap}}
	decision, err := EvaluateTransformWorker(plan, WorkerCapabilities{MediaTranscode: true, SubtitleBurnIn: false, HDRToneMap: true})
	if err != nil {
		t.Fatal(err)
	}
	if decision.Supported || len(decision.Missing) != 1 || decision.Missing[0] != TransformSubtitleBurnIn {
		t.Fatalf("decision=%+v", decision)
	}
	full, err := EvaluateTransformWorker(plan, WorkerCapabilities{true, true, true})
	if err != nil || !full.Supported || len(full.Missing) != 0 {
		t.Fatalf("full=%+v err=%v", full, err)
	}
}

func TestEvaluateTransformWorkerRejectsMalformedPlan(t *testing.T) {
	if _, err := EvaluateTransformWorker(TransformPlan{Requirements: []TransformRequirement{TransformMediaTranscode, TransformMediaTranscode}}, WorkerCapabilities{}); err == nil {
		t.Fatal("expected duplicate rejection")
	}
	if _, err := EvaluateTransformWorker(TransformPlan{Requirements: []TransformRequirement{TransformRequirement("unknown")}}, WorkerCapabilities{}); err == nil {
		t.Fatal("expected unknown rejection")
	}
}

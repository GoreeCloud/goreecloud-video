package playback

import (
	"testing"
	"time"
)

func TestTransformJobLifecycleCountsAttemptsAndClearsFailure(t *testing.T) {
	now := time.Date(2026, 9, 7, 19, 0, 0, 0, time.UTC)
	request := TransformRequest{sourceID: "source-1", requirements: []TransformRequirement{TransformMediaTranscode}}
	job, err := NewTransformJob("job-1", request, now)
	if err != nil || job.State() != TransformJobQueued {
		t.Fatalf("job=%+v err=%v", job, err)
	}

	job, err = job.Transition(TransformJobRunning, now.Add(time.Second), "")
	if err != nil || job.Attempts() != 1 {
		t.Fatalf("job=%+v err=%v", job, err)
	}
	job, err = job.Transition(TransformJobFailed, now.Add(2*time.Second), "worker-exit")
	if err != nil || job.FailureCode() != "worker-exit" {
		t.Fatalf("job=%+v err=%v", job, err)
	}
	job, err = job.Transition(TransformJobQueued, now.Add(3*time.Second), "")
	if err != nil || job.FailureCode() != "" {
		t.Fatalf("job=%+v err=%v", job, err)
	}
	job, err = job.Transition(TransformJobRunning, now.Add(4*time.Second), "")
	if err != nil || job.Attempts() != 2 {
		t.Fatalf("job=%+v err=%v", job, err)
	}
	job, err = job.Transition(TransformJobSucceeded, now.Add(5*time.Second), "")
	if err != nil || job.State() != TransformJobSucceeded {
		t.Fatalf("job=%+v err=%v", job, err)
	}
	if _, err := job.Transition(TransformJobRunning, now.Add(6*time.Second), ""); err == nil {
		t.Fatal("expected succeeded job to be terminal")
	}
}

func TestTransformJobRejectsInvalidStateAndTimestamp(t *testing.T) {
	now := time.Date(2026, 9, 7, 19, 0, 0, 0, time.UTC)
	request := TransformRequest{sourceID: "source-1", requirements: []TransformRequirement{TransformMediaTranscode}}
	if _, err := NewTransformJob(" bad ", request, now); err == nil {
		t.Fatal("expected invalid job id rejection")
	}
	job, err := NewTransformJob("job-1", request, now)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := job.Transition(TransformJobFailed, now.Add(time.Second), ""); err == nil {
		t.Fatal("expected failed state to require failure code")
	}
	if _, err := job.Transition(TransformJobRunning, now.Add(-time.Second), ""); err == nil {
		t.Fatal("expected backwards timestamp rejection")
	}
}

func TestTransformJobRequestGetterDefensivelyCopiesRequirements(t *testing.T) {
	now := time.Date(2026, 9, 7, 19, 0, 0, 0, time.UTC)
	request := TransformRequest{sourceID: "source-1", requirements: []TransformRequirement{TransformMediaTranscode}}
	job, err := NewTransformJob("job-1", request, now)
	if err != nil {
		t.Fatal(err)
	}
	copy := job.Request()
	copy.requirements[0] = TransformHDRToneMap
	if job.Request().Requirements()[0] != TransformMediaTranscode {
		t.Fatal("job request requirements were mutated through getter")
	}
}

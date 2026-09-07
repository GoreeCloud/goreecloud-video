package playback

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestTransformJobSnapshotRoundTrip(t *testing.T) {
	now := time.Date(2026, 9, 7, 19, 15, 0, 123, time.UTC)
	request := TransformRequest{
		sourceID:     "source-1",
		requirements: []TransformRequirement{TransformMediaTranscode, TransformSubtitleBurnIn},
	}
	job, err := NewTransformJob("job-1", request, now)
	if err != nil {
		t.Fatal(err)
	}
	job, err = job.Transition(TransformJobRunning, now.Add(time.Second), "")
	if err != nil {
		t.Fatal(err)
	}
	job, err = job.Transition(TransformJobFailed, now.Add(2*time.Second), "worker_failed")
	if err != nil {
		t.Fatal(err)
	}

	data, err := EncodeTransformJobSnapshot(job)
	if err != nil {
		t.Fatal(err)
	}
	restored, err := DecodeTransformJobSnapshot(data)
	if err != nil {
		t.Fatal(err)
	}
	if restored.ID() != job.ID() || restored.State() != job.State() || restored.Attempts() != 1 || restored.FailureCode() != "worker_failed" || !restored.UpdatedAt().Equal(job.UpdatedAt()) {
		t.Fatalf("restored=%+v job=%+v", restored, job)
	}
	if restored.Request().SourceID() != "source-1" || len(restored.Request().Requirements()) != 2 {
		t.Fatalf("request=%+v", restored.Request())
	}
}

func TestTransformJobSnapshotRejectsUnsupportedUnknownAndTrailingData(t *testing.T) {
	cases := []struct {
		data []byte
		want error
	}{
		{[]byte(`{"schemaVersion":2,"id":"job","sourceId":"source","requirements":["media-transcode"],"state":"queued","updatedAt":"2026-09-07T19:15:00Z","attempts":0,"failureCode":""}`), ErrUnsupportedTransformJobSnapshot},
		{[]byte(`{"schemaVersion":1,"id":"job","sourceId":"source","requirements":["media-transcode"],"state":"queued","updatedAt":"2026-09-07T19:15:00Z","attempts":0,"failureCode":"","extra":true}`), ErrInvalidTransformJobSnapshot},
		{[]byte(`{"schemaVersion":1,"id":"job","sourceId":"source","requirements":["media-transcode"],"state":"queued","updatedAt":"2026-09-07T19:15:00Z","attempts":0,"failureCode":""} {}`), ErrInvalidTransformJobSnapshot},
	}
	for _, tc := range cases {
		if _, err := DecodeTransformJobSnapshot(tc.data); !errors.Is(err, tc.want) {
			t.Fatalf("error=%v want=%v", err, tc.want)
		}
	}
}

func TestTransformJobSnapshotRejectsMalformedLifecycleState(t *testing.T) {
	cases := [][]byte{
		[]byte(`{"schemaVersion":1,"id":"job","sourceId":"source","requirements":["media-transcode"],"state":"running","updatedAt":"2026-09-07T19:15:00Z","attempts":0,"failureCode":""}`),
		[]byte(`{"schemaVersion":1,"id":"job","sourceId":"source","requirements":["media-transcode"],"state":"failed","updatedAt":"2026-09-07T19:15:00Z","attempts":1,"failureCode":""}`),
		[]byte(`{"schemaVersion":1,"id":"job","sourceId":"source","requirements":["media-transcode","media-transcode"],"state":"queued","updatedAt":"2026-09-07T19:15:00Z","attempts":0,"failureCode":""}`),
		[]byte(`{"schemaVersion":1,"id":"job","sourceId":"source","requirements":["media-transcode"],"state":"queued","updatedAt":"2026-09-07T19:15:00+00:00","attempts":0,"failureCode":""}`),
	}
	for _, data := range cases {
		if _, err := DecodeTransformJobSnapshot(data); !errors.Is(err, ErrInvalidTransformJobSnapshot) {
			t.Fatalf("payload=%s error=%v", data, err)
		}
	}
}

func TestTransformJobSnapshotRejectsOversizedPayload(t *testing.T) {
	data := []byte(strings.Repeat("x", MaxTransformJobSnapshotBytes+1))
	if _, err := DecodeTransformJobSnapshot(data); !errors.Is(err, ErrInvalidTransformJobSnapshot) {
		t.Fatalf("error=%v", err)
	}
}

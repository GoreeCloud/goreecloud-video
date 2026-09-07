package playback

import (
	"errors"
	"math"
	"testing"
	"time"
)

func TestTransformJobRecordUpdateAdvancesRevisionAndPreservesIdentity(t *testing.T) {
	now := time.Date(2026, 9, 7, 19, 30, 0, 0, time.UTC)
	job := testTransformJob(t, "job-1", now)
	record, err := NewTransformJobRecord(job)
	if err != nil {
		t.Fatal(err)
	}
	if record.Revision() != 0 {
		t.Fatalf("revision=%d", record.Revision())
	}

	running, err := job.Transition(TransformJobRunning, now.Add(time.Second), "")
	if err != nil {
		t.Fatal(err)
	}
	next, err := record.Update(0, running)
	if err != nil {
		t.Fatal(err)
	}
	if next.Revision() != 1 {
		t.Fatalf("revision=%d", next.Revision())
	}
	restored, err := next.Restore()
	if err != nil || restored.ID() != "job-1" || restored.State() != TransformJobRunning || restored.Attempts() != 1 {
		t.Fatalf("restored=%+v err=%v", restored, err)
	}
}

func TestTransformJobRecordRejectsStaleWriterAndIdentityReplacement(t *testing.T) {
	now := time.Date(2026, 9, 7, 19, 30, 0, 0, time.UTC)
	record, err := NewTransformJobRecord(testTransformJob(t, "job-1", now))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := record.Update(1, testTransformJob(t, "job-1", now)); !errors.Is(err, ErrStaleTransformJobRevision) {
		t.Fatalf("error=%v", err)
	}
	if _, err := record.Update(0, testTransformJob(t, "job-2", now)); !errors.Is(err, ErrInvalidTransformJobRecord) {
		t.Fatalf("error=%v", err)
	}
}

func TestTransformJobRecordRestoresExternalPayloadAndCopiesBytes(t *testing.T) {
	now := time.Date(2026, 9, 7, 19, 30, 0, 0, time.UTC)
	job := testTransformJob(t, "job-1", now)
	payload, err := EncodeTransformJobSnapshot(job)
	if err != nil {
		t.Fatal(err)
	}
	record, err := RestoreTransformJobRecord(7, payload)
	if err != nil || record.Revision() != 7 {
		t.Fatalf("record=%+v err=%v", record, err)
	}

	external := record.Payload()
	external[0] = 'x'
	if _, err := record.Restore(); err != nil {
		t.Fatal("payload copy mutation escaped into record")
	}
	payload[0] = 'x'
	if _, err := record.Restore(); err != nil {
		t.Fatal("input payload mutation escaped into record")
	}
}

func TestTransformJobRecordRejectsMalformedStoredPayloadAndRevisionOverflow(t *testing.T) {
	if _, err := RestoreTransformJobRecord(1, nil); !errors.Is(err, ErrInvalidTransformJobRecord) {
		t.Fatalf("error=%v", err)
	}
	if _, err := RestoreTransformJobRecord(1, []byte(`{"schemaVersion":1}`)); !errors.Is(err, ErrInvalidTransformJobRecord) {
		t.Fatalf("error=%v", err)
	}

	now := time.Date(2026, 9, 7, 19, 30, 0, 0, time.UTC)
	job := testTransformJob(t, "job-1", now)
	record, err := NewTransformJobRecord(job)
	if err != nil {
		t.Fatal(err)
	}
	record.revision = TransformJobRevision(math.MaxUint64)
	if _, err := record.Update(record.revision, job); !errors.Is(err, ErrTransformJobRevisionExhausted) {
		t.Fatalf("error=%v", err)
	}
}

func testTransformJob(t *testing.T, id string, now time.Time) TransformJob {
	t.Helper()
	job, err := NewTransformJob(
		id,
		TransformRequest{
			sourceID:     "source-1",
			requirements: []TransformRequirement{TransformMediaTranscode},
		},
		now,
	)
	if err != nil {
		t.Fatal(err)
	}
	return job
}

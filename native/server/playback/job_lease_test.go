package playback

import (
	"errors"
	"testing"
	"time"
)

func TestClaimTransformJobCreatesRunningLease(t *testing.T) {
	now := time.Date(2026, 9, 7, 20, 0, 0, 0, time.UTC)
	job := testTransformJob(t, "job-lease-1", now)

	running, lease, err := ClaimTransformJob(
		job,
		"worker-1",
		WorkerCapabilities{MediaTranscode: true},
		now.Add(time.Second),
		5*time.Minute,
	)
	if err != nil {
		t.Fatal(err)
	}
	if running.State() != TransformJobRunning || running.Attempts() != 1 {
		t.Fatalf("state=%s attempts=%d", running.State(), running.Attempts())
	}
	if lease.JobID() != running.ID() || lease.WorkerID() != "worker-1" || !lease.Matches(running, "worker-1") {
		t.Fatalf("lease=%+v", lease)
	}
	if !lease.AcquiredAt().Equal(now.Add(time.Second)) || !lease.ExpiresAt().Equal(now.Add(time.Second+5*time.Minute)) {
		t.Fatalf("acquired=%v expires=%v", lease.AcquiredAt(), lease.ExpiresAt())
	}
}

func TestClaimTransformJobRejectsUnsupportedOrAlreadyRunningJob(t *testing.T) {
	now := time.Date(2026, 9, 7, 20, 0, 0, 0, time.UTC)
	job := testTransformJob(t, "job-lease-1", now)

	if _, _, err := ClaimTransformJob(job, "worker-1", WorkerCapabilities{}, now.Add(time.Second), time.Minute); !errors.Is(err, ErrUnsupportedTransformJobWorker) {
		t.Fatalf("error=%v", err)
	}

	running, _, err := ClaimTransformJob(job, "worker-1", WorkerCapabilities{MediaTranscode: true}, now.Add(time.Second), time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := ClaimTransformJob(running, "worker-2", WorkerCapabilities{MediaTranscode: true}, now.Add(2*time.Second), time.Minute); !errors.Is(err, ErrInvalidTransformJobLease) {
		t.Fatalf("error=%v", err)
	}
}

func TestTransformJobLeaseExpiryAndRenewal(t *testing.T) {
	now := time.Date(2026, 9, 7, 20, 0, 0, 0, time.UTC)
	job := testTransformJob(t, "job-lease-1", now)
	_, lease, err := ClaimTransformJob(job, "worker-1", WorkerCapabilities{MediaTranscode: true}, now, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	expired, err := lease.Expired(now.Add(59 * time.Minute))
	if err != nil || expired {
		t.Fatalf("expired=%v err=%v", expired, err)
	}
	expired, err = lease.Expired(now.Add(time.Hour))
	if err != nil || !expired {
		t.Fatalf("expired=%v err=%v", expired, err)
	}

	renewed, err := lease.Renew(now.Add(30*time.Minute), MaxTransformLeaseDuration)
	if err != nil {
		t.Fatal(err)
	}
	if !renewed.ExpiresAt().Equal(now.Add(30*time.Minute + MaxTransformLeaseDuration)) {
		t.Fatalf("expires=%v", renewed.ExpiresAt())
	}
	if _, err := lease.Renew(now.Add(time.Hour), time.Minute); !errors.Is(err, ErrInvalidTransformJobLease) {
		t.Fatalf("error=%v", err)
	}
}

func TestTransformJobLeaseRejectsInvalidWorkerAndWindow(t *testing.T) {
	now := time.Date(2026, 9, 7, 20, 0, 0, 0, time.UTC)
	job := testTransformJob(t, "job-lease-1", now)
	capabilities := WorkerCapabilities{MediaTranscode: true}

	cases := []struct {
		workerID string
		at       time.Time
		ttl      time.Duration
	}{
		{"", now, time.Minute},
		{" worker-1", now, time.Minute},
		{"worker-1", time.Time{}, time.Minute},
		{"worker-1", now, 0},
		{"worker-1", now, MaxTransformLeaseDuration + time.Nanosecond},
	}
	for i, tc := range cases {
		if _, _, err := ClaimTransformJob(job, tc.workerID, capabilities, tc.at, tc.ttl); !errors.Is(err, ErrInvalidTransformJobLease) {
			t.Fatalf("case=%d error=%v", i, err)
		}
	}
}

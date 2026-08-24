package library

import "testing"

func TestNewScanCandidateAcceptsVideo(t *testing.T) {
	candidate, ok := NewScanCandidate("/media/Family.MKV")
	if !ok || candidate.Path != "/media/Family.MKV" {
		t.Fatalf("expected MKV candidate, got %+v ok=%v", candidate, ok)
	}
}

func TestNewScanCandidateRejectsAudioAndImages(t *testing.T) {
	for _, path := range []string{"/media/song.flac", "/media/photo.jpg", "/media/no-extension"} {
		if _, ok := NewScanCandidate(path); ok {
			t.Fatalf("expected %q to be rejected", path)
		}
	}
}

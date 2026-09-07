package playback

import "testing"

func TestDecideSubtitleOffWhenNoTrackSelected(t *testing.T) {
	decision := DecideSubtitle(nil, SubtitleCapabilities{})
	if decision.Mode != SubtitleOff {
		t.Fatalf("mode = %q, want %q", decision.Mode, SubtitleOff)
	}
}

func TestDecideSubtitleDirectWhenCodecSupported(t *testing.T) {
	decision := DecideSubtitle(&SubtitleTrack{Codec: " SRT "}, SubtitleCapabilities{
		Codecs: map[string]bool{"srt": true},
	})
	if decision.Mode != SubtitleDirect || decision.ReasonCode != SubtitleReasonCodecSupported {
		t.Fatalf("decision = %+v, want direct codec-supported", decision)
	}
}

func TestDecideSubtitleBurnInWhenRequiredAndAllowed(t *testing.T) {
	decision := DecideSubtitle(&SubtitleTrack{Codec: "pgs"}, SubtitleCapabilities{AllowBurnIn: true})
	if decision.Mode != SubtitleBurnIn || decision.ReasonCode != SubtitleReasonBurnInRequired {
		t.Fatalf("decision = %+v, want burn-in", decision)
	}
}

func TestDecideSubtitleDeniedWhenBurnInUnavailable(t *testing.T) {
	decision := DecideSubtitle(&SubtitleTrack{Codec: "pgs"}, SubtitleCapabilities{})
	if decision.Mode != SubtitleDenied || decision.ReasonCode != SubtitleReasonBurnInUnavailable {
		t.Fatalf("decision = %+v, want denied burn-in-unavailable", decision)
	}
}

func TestDecideSubtitleRejectsBlankCodec(t *testing.T) {
	decision := DecideSubtitle(&SubtitleTrack{Codec: "  "}, SubtitleCapabilities{AllowBurnIn: true})
	if decision.Mode != SubtitleDenied || decision.ReasonCode != SubtitleReasonInvalidTrack {
		t.Fatalf("decision = %+v, want invalid track denial", decision)
	}
}

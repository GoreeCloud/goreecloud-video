package playback

import "testing"

func TestDecideHDROffWithoutProfile(t *testing.T) {
	decision := DecideHDR(nil, HDRCapabilities{})
	if decision.Mode != HDROff || decision.ReasonCode != "" {
		t.Fatalf("unexpected HDR decision: %+v", decision)
	}
}

func TestDecideHDRDirectNormalizesFormat(t *testing.T) {
	decision := DecideHDR(
		&HDRProfile{Format: " HDR10 "},
		HDRCapabilities{Formats: map[string]bool{"hdr10": true}},
	)
	if decision.Mode != HDRDirect || decision.ReasonCode != HDRReasonFormatSupported {
		t.Fatalf("unexpected HDR decision: %+v", decision)
	}
}

func TestDecideHDRToneMap(t *testing.T) {
	decision := DecideHDR(
		&HDRProfile{Format: "dolby-vision"},
		HDRCapabilities{AllowToneMapping: true},
	)
	if decision.Mode != HDRToneMap || decision.ReasonCode != HDRReasonToneMappingRequired {
		t.Fatalf("unexpected HDR decision: %+v", decision)
	}
}

func TestDecideHDRDeniedWithoutToneMapping(t *testing.T) {
	decision := DecideHDR(&HDRProfile{Format: "hlg"}, HDRCapabilities{})
	if decision.Mode != HDRDenied || decision.ReasonCode != HDRReasonToneMappingUnavailable {
		t.Fatalf("unexpected HDR decision: %+v", decision)
	}
}

func TestDecideHDRRejectsBlankProfile(t *testing.T) {
	decision := DecideHDR(
		&HDRProfile{Format: "   "},
		HDRCapabilities{AllowToneMapping: true},
	)
	if decision.Mode != HDRDenied || decision.ReasonCode != HDRReasonInvalidProfile {
		t.Fatalf("unexpected HDR decision: %+v", decision)
	}
}

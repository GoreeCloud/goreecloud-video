package playback

import (
	"reflect"
	"testing"
)

func TestDecideDirectPlay(t *testing.T) {
	decision := Decide(MediaProfile{Container: "mkv", VideoCodec: "h264", AudioCodec: "aac", Width: 1920, Height: 1080, Bitrate: 8_000_000}, ClientCapabilities{
		Containers: map[string]bool{"mkv": true}, VideoCodecs: map[string]bool{"h264": true}, AudioCodecs: map[string]bool{"aac": true}, MaxWidth: 3840, MaxHeight: 2160, MaxBitrate: 20_000_000,
	})
	if decision.Mode != ModeDirectPlay {
		t.Fatalf("mode = %q, want %q", decision.Mode, ModeDirectPlay)
	}
	if len(decision.ReasonCodes) != 0 {
		t.Fatalf("direct-play reason codes = %v, want none", decision.ReasonCodes)
	}
}

func TestDecideRemuxWhenOnlyContainerDiffers(t *testing.T) {
	decision := Decide(MediaProfile{Container: "mkv", VideoCodec: "h264", AudioCodec: "aac"}, ClientCapabilities{
		Containers: map[string]bool{"mp4": true}, VideoCodecs: map[string]bool{"h264": true}, AudioCodecs: map[string]bool{"aac": true},
	})
	if decision.Mode != ModeRemux {
		t.Fatalf("mode = %q, want %q", decision.Mode, ModeRemux)
	}
	if !reflect.DeepEqual(decision.ReasonCodes, []ReasonCode{ReasonContainerUnsupported}) {
		t.Fatalf("reason codes = %v", decision.ReasonCodes)
	}
}

func TestDecideTranscodeWhenCodecUnsupported(t *testing.T) {
	decision := Decide(MediaProfile{Container: "mkv", VideoCodec: "hevc", AudioCodec: "aac"}, ClientCapabilities{
		Containers: map[string]bool{"mkv": true}, VideoCodecs: map[string]bool{"h264": true}, AudioCodecs: map[string]bool{"aac": true}, AllowTranscoding: true,
	})
	if decision.Mode != ModeTranscode {
		t.Fatalf("mode = %q, want %q", decision.Mode, ModeTranscode)
	}
	if !reflect.DeepEqual(decision.ReasonCodes, []ReasonCode{ReasonVideoCodecUnsupported}) {
		t.Fatalf("reason codes = %v", decision.ReasonCodes)
	}
}

func TestDecideReportsAllTranscodeConstraints(t *testing.T) {
	decision := Decide(MediaProfile{Container: "mkv", VideoCodec: "hevc", AudioCodec: "ac3", Width: 3840, Height: 2160, Bitrate: 20_000_000}, ClientCapabilities{
		Containers: map[string]bool{"mkv": true}, VideoCodecs: map[string]bool{"h264": true}, AudioCodecs: map[string]bool{"aac": true}, MaxWidth: 1920, MaxHeight: 1080, MaxBitrate: 8_000_000, AllowTranscoding: true,
	})
	want := []ReasonCode{ReasonVideoCodecUnsupported, ReasonAudioCodecUnsupported, ReasonResolutionUnsupported, ReasonBitrateUnsupported}
	if decision.Mode != ModeTranscode || !reflect.DeepEqual(decision.ReasonCodes, want) {
		t.Fatalf("decision = %+v, want reason codes %v", decision, want)
	}
}

func TestDecideDeniedWhenTransformationDisabled(t *testing.T) {
	decision := Decide(MediaProfile{Container: "mkv", VideoCodec: "hevc"}, ClientCapabilities{
		Containers: map[string]bool{"mkv": true}, VideoCodecs: map[string]bool{"h264": true}, AllowTranscoding: false,
	})
	if decision.Mode != ModeDenied {
		t.Fatalf("mode = %q, want %q", decision.Mode, ModeDenied)
	}
	want := []ReasonCode{ReasonVideoCodecUnsupported, ReasonTranscodingDisabled}
	if !reflect.DeepEqual(decision.ReasonCodes, want) {
		t.Fatalf("reason codes = %v, want %v", decision.ReasonCodes, want)
	}
}

func TestDecideRejectsInvalidSourceProfile(t *testing.T) {
	tests := []MediaProfile{
		{VideoCodec: "h264"},
		{Container: "mkv"},
		{Container: "mkv", VideoCodec: "h264", Width: -1},
		{Container: "mkv", VideoCodec: "h264", Height: -1},
		{Container: "mkv", VideoCodec: "h264", Bitrate: -1},
	}

	for _, media := range tests {
		decision := Decide(media, ClientCapabilities{AllowTranscoding: true})
		if decision.Mode != ModeDenied {
			t.Fatalf("media %+v mode = %q, want %q", media, decision.Mode, ModeDenied)
		}
		if !reflect.DeepEqual(decision.ReasonCodes, []ReasonCode{ReasonInvalidSource}) {
			t.Fatalf("media %+v reason codes = %v", media, decision.ReasonCodes)
		}
	}
}

func TestDecideRejectsInvalidClientLimits(t *testing.T) {
	tests := []ClientCapabilities{
		{MaxWidth: -1, AllowTranscoding: true},
		{MaxHeight: -1, AllowTranscoding: true},
		{MaxBitrate: -1, AllowTranscoding: true},
	}
	media := MediaProfile{Container: "mkv", VideoCodec: "h264"}

	for _, client := range tests {
		decision := Decide(media, client)
		if decision.Mode != ModeDenied {
			t.Fatalf("client %+v mode = %q, want %q", client, decision.Mode, ModeDenied)
		}
		if !reflect.DeepEqual(decision.ReasonCodes, []ReasonCode{ReasonInvalidClientLimits}) {
			t.Fatalf("client %+v reason codes = %v", client, decision.ReasonCodes)
		}
	}
}

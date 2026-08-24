package playback

import "testing"

func TestDecideDirectPlay(t *testing.T) {
	decision := Decide(MediaProfile{Container: "mkv", VideoCodec: "h264", AudioCodec: "aac", Width: 1920, Height: 1080, Bitrate: 8_000_000}, ClientCapabilities{
		Containers: map[string]bool{"mkv": true}, VideoCodecs: map[string]bool{"h264": true}, AudioCodecs: map[string]bool{"aac": true}, MaxWidth: 3840, MaxHeight: 2160, MaxBitrate: 20_000_000,
	})
	if decision.Mode != ModeDirectPlay {
		t.Fatalf("mode = %q, want %q", decision.Mode, ModeDirectPlay)
	}
}

func TestDecideRemuxWhenOnlyContainerDiffers(t *testing.T) {
	decision := Decide(MediaProfile{Container: "mkv", VideoCodec: "h264", AudioCodec: "aac"}, ClientCapabilities{
		Containers: map[string]bool{"mp4": true}, VideoCodecs: map[string]bool{"h264": true}, AudioCodecs: map[string]bool{"aac": true},
	})
	if decision.Mode != ModeRemux {
		t.Fatalf("mode = %q, want %q", decision.Mode, ModeRemux)
	}
}

func TestDecideTranscodeWhenCodecUnsupported(t *testing.T) {
	decision := Decide(MediaProfile{Container: "mkv", VideoCodec: "hevc", AudioCodec: "aac"}, ClientCapabilities{
		Containers: map[string]bool{"mkv": true}, VideoCodecs: map[string]bool{"h264": true}, AudioCodecs: map[string]bool{"aac": true}, AllowTranscoding: true,
	})
	if decision.Mode != ModeTranscode {
		t.Fatalf("mode = %q, want %q", decision.Mode, ModeTranscode)
	}
}

func TestDecideDeniedWhenTransformationDisabled(t *testing.T) {
	decision := Decide(MediaProfile{Container: "mkv", VideoCodec: "hevc"}, ClientCapabilities{
		Containers: map[string]bool{"mkv": true}, VideoCodecs: map[string]bool{"h264": true}, AllowTranscoding: false,
	})
	if decision.Mode != ModeDenied {
		t.Fatalf("mode = %q, want %q", decision.Mode, ModeDenied)
	}
}

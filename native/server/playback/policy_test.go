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
	}
}

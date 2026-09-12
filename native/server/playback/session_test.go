package playback

import "testing"

func sessionTestMedia() MediaProfile {
	return MediaProfile{
		Container:  "mp4",
		VideoCodec: "h264",
		AudioCodec: "aac",
		Width:      1920,
		Height:     1080,
		Bitrate:    5_000_000,
	}
}

func sessionTestClient(allowTranscoding bool) ClientCapabilities {
	return ClientCapabilities{
		Containers:       map[string]bool{"mp4": true},
		VideoCodecs:      map[string]bool{"h264": true},
		AudioCodecs:      map[string]bool{"aac": true},
		AllowTranscoding: allowTranscoding,
	}
}

func TestDecideSessionKeepsDirectPlayback(t *testing.T) {
	plan := DecideSession(
		sessionTestMedia(),
		sessionTestClient(true),
		nil,
		SubtitleCapabilities{},
		nil,
		HDRCapabilities{},
	)
	if plan.Mode != ModeDirectPlay || plan.RequiresTranscoding {
		t.Fatalf("unexpected direct session plan: %+v", plan)
	}
}

func TestDecideSessionUpgradesBurnInAndToneMapToTranscode(t *testing.T) {
	subtitle := &SubtitleTrack{Codec: "ass"}
	hdr := &HDRProfile{Format: "hdr10"}
	plan := DecideSession(
		sessionTestMedia(),
		sessionTestClient(true),
		subtitle,
		SubtitleCapabilities{AllowBurnIn: true},
		hdr,
		HDRCapabilities{AllowToneMapping: true},
	)
	if plan.Mode != ModeTranscode || !plan.RequiresTranscoding {
		t.Fatalf("unexpected transformed session plan: %+v", plan)
	}
}

func TestDecideSessionDeniesTransformationWhenPrimaryTranscodingDisabled(t *testing.T) {
	subtitle := &SubtitleTrack{Codec: "ass"}
	plan := DecideSession(
		sessionTestMedia(),
		sessionTestClient(false),
		subtitle,
		SubtitleCapabilities{AllowBurnIn: true},
		nil,
		HDRCapabilities{},
	)
	if plan.Mode != ModeDenied {
		t.Fatalf("unexpected session plan: %+v", plan)
	}
}

func TestDecideSessionPropagatesSubsystemDenial(t *testing.T) {
	subtitle := &SubtitleTrack{Codec: "ass"}
	plan := DecideSession(
		sessionTestMedia(),
		sessionTestClient(true),
		subtitle,
		SubtitleCapabilities{},
		nil,
		HDRCapabilities{},
	)
	if plan.Mode != ModeDenied {
		t.Fatalf("unexpected session plan: %+v", plan)
	}
}

package playback

import "strings"

type Mode string

const (
	ModeDirectPlay Mode = "direct-play"
	ModeRemux      Mode = "remux"
	ModeTranscode  Mode = "transcode"
	ModeDenied     Mode = "denied"
)

type ReasonCode string

const (
	ReasonInvalidSource         ReasonCode = "invalid-source"
	ReasonInvalidClientLimits   ReasonCode = "invalid-client-limits"
	ReasonContainerUnsupported  ReasonCode = "container-unsupported"
	ReasonVideoCodecUnsupported ReasonCode = "video-codec-unsupported"
	ReasonAudioCodecUnsupported ReasonCode = "audio-codec-unsupported"
	ReasonResolutionUnsupported ReasonCode = "resolution-unsupported"
	ReasonBitrateUnsupported    ReasonCode = "bitrate-unsupported"
	ReasonTranscodingDisabled   ReasonCode = "transcoding-disabled"
)

type MediaProfile struct {
	Container, VideoCodec, AudioCodec string
	Width, Height                     int
	Bitrate                           int64
}

type ClientCapabilities struct {
	Containers, VideoCodecs, AudioCodecs map[string]bool
	MaxWidth, MaxHeight                  int
	MaxBitrate                           int64
	AllowTranscoding                     bool
}

type Decision struct {
	Mode        Mode
	Reason      string
	ReasonCodes []ReasonCode
}

func Decide(media MediaProfile, client ClientCapabilities) Decision {
	if !validMediaProfile(media) {
		return Decision{Mode: ModeDenied, Reason: "invalid source media profile", ReasonCodes: []ReasonCode{ReasonInvalidSource}}
	}
	if !validClientCapabilities(client) {
		return Decision{Mode: ModeDenied, Reason: "invalid client capability limits", ReasonCodes: []ReasonCode{ReasonInvalidClientLimits}}
	}

	reasons := compatibilityReasons(media, client)
	if len(reasons) == 0 {
		return Decision{Mode: ModeDirectPlay, Reason: "client supports source container, codecs, and limits"}
	}
	if len(reasons) == 1 && reasons[0] == ReasonContainerUnsupported {
		return Decision{Mode: ModeRemux, Reason: "codecs are compatible but container is not", ReasonCodes: reasons}
	}
	if client.AllowTranscoding {
		return Decision{Mode: ModeTranscode, Reason: "source exceeds direct-play or remux capabilities", ReasonCodes: reasons}
	}
	reasons = append(reasons, ReasonTranscodingDisabled)
	return Decision{Mode: ModeDenied, Reason: "client cannot play the source and transcoding is disabled", ReasonCodes: reasons}
}

func compatibilityReasons(media MediaProfile, client ClientCapabilities) []ReasonCode {
	reasons := make([]ReasonCode, 0, 5)
	if !client.Containers[normalize(media.Container)] {
		reasons = append(reasons, ReasonContainerUnsupported)
	}
	if !client.VideoCodecs[normalize(media.VideoCodec)] {
		reasons = append(reasons, ReasonVideoCodecUnsupported)
	}
	if media.AudioCodec != "" && !client.AudioCodecs[normalize(media.AudioCodec)] {
		reasons = append(reasons, ReasonAudioCodecUnsupported)
	}
	if !withinLimit(media.Width, client.MaxWidth) || !withinLimit(media.Height, client.MaxHeight) {
		reasons = append(reasons, ReasonResolutionUnsupported)
	}
	if !withinBitrate(media.Bitrate, client.MaxBitrate) {
		reasons = append(reasons, ReasonBitrateUnsupported)
	}
	return reasons
}

func validMediaProfile(media MediaProfile) bool {
	return normalize(media.Container) != "" && normalize(media.VideoCodec) != "" && media.Width >= 0 && media.Height >= 0 && media.Bitrate >= 0
}

func validClientCapabilities(client ClientCapabilities) bool {
	return client.MaxWidth >= 0 && client.MaxHeight >= 0 && client.MaxBitrate >= 0
}

func normalize(value string) string         { return strings.ToLower(strings.TrimSpace(value)) }
func withinLimit(value, limit int) bool     { return limit == 0 || value <= limit }
func withinBitrate(value, limit int64) bool { return limit == 0 || value <= limit }

package playback

import "strings"

type Mode string

const (
	ModeDirectPlay Mode = "direct-play"
	ModeRemux      Mode = "remux"
	ModeTranscode  Mode = "transcode"
	ModeDenied     Mode = "denied"
)

type MediaProfile struct {
	Container  string
	VideoCodec string
	AudioCodec string
	Width      int
	Height     int
	Bitrate    int64
}

type ClientCapabilities struct {
	Containers       map[string]bool
	VideoCodecs      map[string]bool
	AudioCodecs      map[string]bool
	MaxWidth         int
	MaxHeight        int
	MaxBitrate       int64
	AllowTranscoding bool
}

type Decision struct {
	Mode   Mode
	Reason string
}

func Decide(media MediaProfile, client ClientCapabilities) Decision {
	if !validMediaProfile(media) {
		return Decision{Mode: ModeDenied, Reason: "invalid source media profile"}
	}
	if !validClientCapabilities(client) {
		return Decision{Mode: ModeDenied, Reason: "invalid client capability limits"}
	}

	containerOK := client.Containers[normalize(media.Container)]
	videoOK := client.VideoCodecs[normalize(media.VideoCodec)]
	audioOK := media.AudioCodec == "" || client.AudioCodecs[normalize(media.AudioCodec)]
	limitsOK := withinLimit(media.Width, client.MaxWidth) && withinLimit(media.Height, client.MaxHeight) && withinBitrate(media.Bitrate, client.MaxBitrate)

	if containerOK && videoOK && audioOK && limitsOK {
		return Decision{Mode: ModeDirectPlay, Reason: "client supports source container, codecs, and limits"}
	}
	if videoOK && audioOK && limitsOK {
		return Decision{Mode: ModeRemux, Reason: "codecs are compatible but container is not"}
	}
	if client.AllowTranscoding {
		return Decision{Mode: ModeTranscode, Reason: "source exceeds direct-play or remux capabilities"}
	}
	return Decision{Mode: ModeDenied, Reason: "client cannot play the source and transcoding is disabled"}
}

func validMediaProfile(media MediaProfile) bool {
	return normalize(media.Container) != "" &&
		normalize(media.VideoCodec) != "" &&
		media.Width >= 0 &&
		media.Height >= 0 &&
		media.Bitrate >= 0
}

func validClientCapabilities(client ClientCapabilities) bool {
	return client.MaxWidth >= 0 && client.MaxHeight >= 0 && client.MaxBitrate >= 0
}

func normalize(value string) string { return strings.ToLower(strings.TrimSpace(value)) }

func withinLimit(value, limit int) bool     { return limit == 0 || value <= limit }
func withinBitrate(value, limit int64) bool { return limit == 0 || value <= limit }

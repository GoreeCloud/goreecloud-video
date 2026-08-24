package playback

import "strings"

type Mode string

const (
	ModeDirectPlay Mode = "direct-play"
	ModeRemux      Mode = "remux"
	ModeTranscode  Mode = "transcode"
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
	return Decision{Mode: ModeDirectPlay, Reason: "no supported transformation is authorized"}
}

func normalize(value string) string { return strings.ToLower(strings.TrimSpace(value)) }

func withinLimit(value, limit int) bool { return limit <= 0 || value <= limit }
func withinBitrate(value, limit int64) bool { return limit <= 0 || value <= limit }

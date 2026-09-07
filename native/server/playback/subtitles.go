package playback

type SubtitleMode string

const (
	SubtitleOff    SubtitleMode = "off"
	SubtitleDirect SubtitleMode = "direct"
	SubtitleBurnIn SubtitleMode = "burn-in"
	SubtitleDenied SubtitleMode = "denied"
)

type SubtitleReasonCode string

const (
	SubtitleReasonInvalidTrack      SubtitleReasonCode = "invalid-track"
	SubtitleReasonCodecSupported    SubtitleReasonCode = "codec-supported"
	SubtitleReasonBurnInRequired    SubtitleReasonCode = "burn-in-required"
	SubtitleReasonBurnInUnavailable SubtitleReasonCode = "burn-in-unavailable"
)

type SubtitleTrack struct {
	Codec string
}

type SubtitleCapabilities struct {
	Codecs      map[string]bool
	AllowBurnIn bool
}

type SubtitleDecision struct {
	Mode       SubtitleMode
	Reason     string
	ReasonCode SubtitleReasonCode
}

func DecideSubtitle(track *SubtitleTrack, client SubtitleCapabilities) SubtitleDecision {
	if track == nil {
		return SubtitleDecision{Mode: SubtitleOff, Reason: "no subtitle track selected"}
	}
	codec := normalize(track.Codec)
	if codec == "" {
		return SubtitleDecision{Mode: SubtitleDenied, Reason: "invalid subtitle track", ReasonCode: SubtitleReasonInvalidTrack}
	}
	if client.Codecs[codec] {
		return SubtitleDecision{Mode: SubtitleDirect, Reason: "client supports selected subtitle codec", ReasonCode: SubtitleReasonCodecSupported}
	}
	if client.AllowBurnIn {
		return SubtitleDecision{Mode: SubtitleBurnIn, Reason: "selected subtitle codec requires burn-in", ReasonCode: SubtitleReasonBurnInRequired}
	}
	return SubtitleDecision{Mode: SubtitleDenied, Reason: "selected subtitle codec is unsupported and burn-in is disabled", ReasonCode: SubtitleReasonBurnInUnavailable}
}

package playback

type SessionPlan struct {
	Mode                Mode
	Media               Decision
	Subtitle            SubtitleDecision
	HDR                 HDRDecision
	RequiresTranscoding bool
	Reason              string
}

// DecideSession combines the primary media, subtitle, and HDR policies into one
// playback decision. The planner may require transcoding but does not execute a
// transcoder, subtitle renderer, or tone-mapping worker.
func DecideSession(
	media MediaProfile,
	client ClientCapabilities,
	subtitle *SubtitleTrack,
	subtitleClient SubtitleCapabilities,
	hdr *HDRProfile,
	hdrClient HDRCapabilities,
) SessionPlan {
	mediaDecision := Decide(media, client)
	subtitleDecision := DecideSubtitle(subtitle, subtitleClient)
	hdrDecision := DecideHDR(hdr, hdrClient)
	plan := SessionPlan{
		Mode:     mediaDecision.Mode,
		Media:    mediaDecision,
		Subtitle: subtitleDecision,
		HDR:      hdrDecision,
	}

	if mediaDecision.Mode == ModeDenied || subtitleDecision.Mode == SubtitleDenied || hdrDecision.Mode == HDRDenied {
		plan.Mode = ModeDenied
		plan.Reason = "one or more playback policies denied the session"
		return plan
	}

	needsTransform := mediaDecision.Mode == ModeTranscode || subtitleDecision.Mode == SubtitleBurnIn || hdrDecision.Mode == HDRToneMap
	if needsTransform && !client.AllowTranscoding {
		plan.Mode = ModeDenied
		plan.Reason = "session requires media transformation but transcoding is disabled"
		return plan
	}
	if needsTransform {
		plan.Mode = ModeTranscode
		plan.RequiresTranscoding = true
		plan.Reason = "session requires transcoding or media transformation"
		return plan
	}

	plan.Reason = "session can use the primary media playback mode without transformation"
	return plan
}

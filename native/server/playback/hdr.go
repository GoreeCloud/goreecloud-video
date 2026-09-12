package playback

type HDRMode string

const (
	HDROff     HDRMode = "off"
	HDRDirect  HDRMode = "direct"
	HDRToneMap HDRMode = "tone-map"
	HDRDenied  HDRMode = "denied"
)

type HDRReasonCode string

const (
	HDRReasonInvalidProfile         HDRReasonCode = "invalid-profile"
	HDRReasonFormatSupported        HDRReasonCode = "format-supported"
	HDRReasonToneMappingRequired    HDRReasonCode = "tone-mapping-required"
	HDRReasonToneMappingUnavailable HDRReasonCode = "tone-mapping-unavailable"
)

type HDRProfile struct {
	Format string
}

type HDRCapabilities struct {
	Formats          map[string]bool
	AllowToneMapping bool
}

type HDRDecision struct {
	Mode       HDRMode
	Reason     string
	ReasonCode HDRReasonCode
}

// DecideHDR determines whether an HDR source may be sent directly, requires
// tone mapping, or must be denied. A nil profile represents SDR/no HDR metadata.
// This policy does not execute a transcoder or tone-mapping worker.
func DecideHDR(profile *HDRProfile, client HDRCapabilities) HDRDecision {
	if profile == nil {
		return HDRDecision{Mode: HDROff, Reason: "source is SDR or has no HDR profile"}
	}
	format := normalize(profile.Format)
	if format == "" {
		return HDRDecision{Mode: HDRDenied, Reason: "invalid HDR source profile", ReasonCode: HDRReasonInvalidProfile}
	}
	if client.Formats[format] {
		return HDRDecision{Mode: HDRDirect, Reason: "client supports source HDR format", ReasonCode: HDRReasonFormatSupported}
	}
	if client.AllowToneMapping {
		return HDRDecision{Mode: HDRToneMap, Reason: "source HDR format requires tone mapping", ReasonCode: HDRReasonToneMappingRequired}
	}
	return HDRDecision{Mode: HDRDenied, Reason: "client does not support source HDR format and tone mapping is disabled", ReasonCode: HDRReasonToneMappingUnavailable}
}

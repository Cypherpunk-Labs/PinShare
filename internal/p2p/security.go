package p2p

import "pinshare/internal/types"

// SecurityCapability is re-exported from the types package for backwards compatibility.
// This allows existing code using p2p.SecurityCapability to continue working.
type SecurityCapability = types.SecurityCapability

// Re-export security capability constants for backwards compatibility.
const (
	SecurityCapabilityNone          = types.SecurityCapabilityNone
	SecurityCapabilityP2PSec        = types.SecurityCapabilityP2PSec
	SecurityCapabilityVirusTotal    = types.SecurityCapabilityVirusTotal
	SecurityCapabilityClamAV        = types.SecurityCapabilityClamAV
	SecurityCapabilityVirusTotalBrowser = types.SecurityCapabilityVirusTotalBrowser
)

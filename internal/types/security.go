// Package types provides shared type definitions used across multiple packages.
package types

// SecurityCapability defines the available security scanning backends.
// These constants should be used instead of magic numbers when checking
// or setting the security capability level.
type SecurityCapability int

const (
	// SecurityCapabilityNone indicates no scanning backend is available.
	// The application will fail to start with this capability.
	SecurityCapabilityNone SecurityCapability = 0

	// SecurityCapabilityP2PSec uses the P2P-Sec service on port 36939.
	SecurityCapabilityP2PSec SecurityCapability = 1

	// SecurityCapabilityVirusTotal uses the VirusTotal API with VT_TOKEN env var.
	SecurityCapabilityVirusTotal SecurityCapability = 2

	// SecurityCapabilityClamAV uses ClamAV (clamscan must be in PATH).
	SecurityCapabilityClamAV SecurityCapability = 3

	// SecurityCapabilityVirusTotalBrowser uses VirusTotal via browser automation.
	// Requires Chromium to be installed.
	SecurityCapabilityVirusTotalBrowser SecurityCapability = 4
)

// UsesClamAV returns true if the security capability uses ClamAV for scanning.
// Capabilities 1, 2, and 3 all use ClamAV as the primary scanner.
func (sc SecurityCapability) UsesClamAV() bool {
	return sc >= SecurityCapabilityP2PSec && sc <= SecurityCapabilityClamAV
}

// UsesVirusTotalBrowser returns true if the security capability uses VirusTotal via browser.
func (sc SecurityCapability) UsesVirusTotalBrowser() bool {
	return sc == SecurityCapabilityVirusTotalBrowser
}

package p2p

import (
	"fmt"
	"pinshare/internal/psfs"
	"pinshare/internal/store"
)

func ProcessDownload(metadata store.BaseMetadata) (bool, error) {
	if appconfInstance.SecurityCapability == int(SecurityCapabilityNone) {
		fmt.Println("[ERROR] No security capability configured")
		return false, nil
	}

	fmt.Println("[INFO] File Security checking CID: " + metadata.IPFSCID + " with SHA256: " + metadata.FileSHA256)

	fresult, err := performSecurityScan(metadata)
	if err != nil {
		return false, err
	}

	if !fresult {
		fmt.Println("[ERROR] File Security check failed for CID: " + metadata.IPFSCID + " with SHA256: " + metadata.FileSHA256)
		return false, nil
	}

	// Validate file type
	ftype, err := psfs.ValidateFileType(appconfInstance.CacheFolder + "/" + metadata.IPFSCID + "." + metadata.FileType)
	if err != nil {
		return false, err
	}

	fmt.Println("[INFO] File Security type check passed for CID: " + metadata.IPFSCID + "." + metadata.FileType)

	if !ftype {
		return false, nil
	}

	psfs.PinFileIPFS(metadata.IPFSCID)
	fmt.Println("[INFO] IPFS Pinned for CID: " + metadata.IPFSCID)
	return true, nil
}

// performSecurityScan handles the security scanning based on the configured capability.
func performSecurityScan(metadata store.BaseMetadata) (bool, error) {
	capability := SecurityCapability(appconfInstance.SecurityCapability)
	cachePath := appconfInstance.CacheFolder + "/" + metadata.IPFSCID + "." + metadata.FileType

	// Skip all security scanning if FFSkipVT is enabled
	if appconfInstance.FFSkipVT {
		fmt.Println("[INFO] Virus scanning disabled (FFSkipVT=true), skipping security check")
		fmt.Println("[INFO] Fetching CID: " + metadata.IPFSCID)
		psfs.GetFileIPFS(metadata.IPFSCID, cachePath)
		return true, nil
	}

	switch {
	case capability.UsesClamAV():
		// SecurityCapability 1, 2, 3: Use ClamAV
		fmt.Println("[INFO] Fetching CID: " + metadata.IPFSCID)
		psfs.GetFileIPFS(metadata.IPFSCID, cachePath)
		return psfs.ClamScanFileClean(cachePath)

	case capability.UsesVirusTotalBrowser():
		// SecurityCapability 4: Use VirusTotal via browser
		result, err := psfs.GetVirusTotalWSVerdictByHash(metadata.FileSHA256)
		if err != nil {
			return false, err
		}
		fmt.Println("[INFO] Fetching CID: " + metadata.IPFSCID)
		psfs.GetFileIPFS(metadata.IPFSCID, cachePath)
		return result, nil

	default:
		fmt.Println("[ERROR] Unknown security capability: ", appconfInstance.SecurityCapability)
		return false, nil
	}
}

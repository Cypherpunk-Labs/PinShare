package p2p

import (
	"fmt"
	"path/filepath"

	"pinshare/internal/psfs"
	"pinshare/internal/store"
)

func ProcessDownload(metadata store.BaseMetadata) (bool, error) {
	capability := appconfInstance.SecurityCapability
	if capability == SecurityCapabilityNone {
		fmt.Println("[ERROR] No security capability configured")
		return false, nil
	}

	fmt.Printf("[INFO] File Security checking CID: %s with SHA256: %s\n", metadata.IPFSCID, metadata.FileSHA256)

	fresult, err := performSecurityScan(metadata)
	if err != nil {
		return false, err
	}

	if !fresult {
		fmt.Printf("[ERROR] File Security check failed for CID: %s with SHA256: %s\n", metadata.IPFSCID, metadata.FileSHA256)
		return false, nil
	}

	// Validate file type using OS-agnostic path construction
	cacheFilePath := filepath.Join(appconfInstance.CacheFolder, metadata.IPFSCID+"."+metadata.FileType)
	ftype, err := psfs.ValidateFileType(cacheFilePath)
	if err != nil {
		return false, err
	}

	fmt.Printf("[INFO] File Security type check passed for CID: %s.%s\n", metadata.IPFSCID, metadata.FileType)

	if !ftype {
		return false, nil
	}

	psfs.PinFileIPFS(metadata.IPFSCID)
	fmt.Printf("[INFO] IPFS Pinned for CID: %s\n", metadata.IPFSCID)
	return true, nil
}

// performSecurityScan handles the security scanning based on the configured capability.
func performSecurityScan(metadata store.BaseMetadata) (bool, error) {
	capability := appconfInstance.SecurityCapability
	// Use OS-agnostic path construction
	cachePath := filepath.Join(appconfInstance.CacheFolder, metadata.IPFSCID+"."+metadata.FileType)

	// Skip all security scanning if FFSkipVT is enabled
	if appconfInstance.FFSkipVT {
		fmt.Println("[INFO] Virus scanning disabled (FFSkipVT=true), skipping security check")
		fmt.Printf("[INFO] Fetching CID: %s\n", metadata.IPFSCID)
		psfs.GetFileIPFS(metadata.IPFSCID, cachePath)
		return true, nil
	}

	switch {
	case capability.UsesClamAV():
		// SecurityCapability 1, 2, 3: Use ClamAV
		fmt.Printf("[INFO] Fetching CID: %s\n", metadata.IPFSCID)
		psfs.GetFileIPFS(metadata.IPFSCID, cachePath)
		return psfs.ClamScanFileClean(cachePath)

	case capability.UsesVirusTotalBrowser():
		// SecurityCapability 4: Use VirusTotal via browser
		result, err := psfs.GetVirusTotalWSVerdictByHash(metadata.FileSHA256)
		if err != nil {
			return false, err
		}
		fmt.Printf("[INFO] Fetching CID: %s\n", metadata.IPFSCID)
		psfs.GetFileIPFS(metadata.IPFSCID, cachePath)
		return result, nil

	default:
		fmt.Printf("[ERROR] Unknown security capability: %d\n", appconfInstance.SecurityCapability)
		return false, nil
	}
}

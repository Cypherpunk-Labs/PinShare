package p2p

import (
	"fmt"
	"pinshare/internal/psfs"
	"pinshare/internal/store"
)

func ProcessDownload(metadata store.BaseMetadata) (bool, error) {
	returnValue := false

	var fresult bool
	if appconfInstance.SecurityCapability > 0 {
		fmt.Println("[INFO] File Security checking CID: " + metadata.IPFSCID + " with SHA256: " + metadata.FileSHA256)

		// Skip all security scanning if FFSkipVT is enabled
		if appconfInstance.FFSkipVT {
			fmt.Println("[INFO] Virus scanning disabled (FFSkipVT=true), skipping security check")
			fmt.Println("[INFO] Fetching CID: " + metadata.IPFSCID)
			psfs.GetFileIPFS(metadata.IPFSCID, appconfInstance.CacheFolder+"/"+metadata.IPFSCID+"."+metadata.FileType)
			fresult = true
		} else if appconfInstance.SecurityCapability <= 3 {
			// SecurityCapability 1, 2, 3: Use ClamAV
			fmt.Println("[INFO] Fetching CID: " + metadata.IPFSCID)
			// ipfs get
			psfs.GetFileIPFS(metadata.IPFSCID, appconfInstance.CacheFolder+"/"+metadata.IPFSCID+"."+metadata.FileType)

			result, err := psfs.ClamScanFileClean(appconfInstance.CacheFolder + "/" + metadata.IPFSCID + "." + metadata.FileType)
			if err != nil {
				return returnValue, err
			}
			fresult = result
		} else if appconfInstance.SecurityCapability == 4 {
			// SecurityCapability 4: Use VirusTotal via browser
			result, err := psfs.GetVirusTotalWSVerdictByHash(metadata.FileSHA256) // true == safe
			if err != nil {
				return returnValue, err
			}
			// fmt.Println("[INFO] File Security check verdict for CID: " + metadata.IPFSCID + " with SHA256: " + metadata.FileSHA256)
			fresult = result
			fmt.Println("[INFO] Fetching CID: " + metadata.IPFSCID)
			// ipfs get
			psfs.GetFileIPFS(metadata.IPFSCID, appconfInstance.CacheFolder+"/"+metadata.IPFSCID+"."+metadata.FileType)
		}
	}
	if fresult {
		// check file type
		ftype, err := psfs.ValidateFileType(appconfInstance.CacheFolder + "/" + metadata.IPFSCID + "." + metadata.FileType)
		if err != nil {
			return returnValue, err
		}
		fmt.Println("[INFO] File Security type check passed for CID: " + metadata.IPFSCID + "." + metadata.FileType)
		if ftype {
			psfs.PinFileIPFS(metadata.IPFSCID)
			fmt.Println("[INFO] IPFS Pinned for CID: " + metadata.IPFSCID)
			returnValue = true
		}
	} else {
		fmt.Println("[ERROR] File Security check failed for CID: " + metadata.IPFSCID + " with SHA256: " + metadata.FileSHA256)
	}
	return returnValue, nil
}

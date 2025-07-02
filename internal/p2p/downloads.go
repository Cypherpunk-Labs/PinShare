package p2p

import (
	"fmt"
	"pinshare/internal/psfs"
	"pinshare/internal/store"
)

func ProcessDownload(metadata store.BaseMetadata) (bool, error) {
	returnValue := false

	var fresult bool
	if appconfInstance.FFSkipVT {
		fresult = true
	} else {
		fmt.Println("[INFO] File Security checking CID: " + metadata.IPFSCID + " with SHA256: " + metadata.FileSHA256)
		result, err := psfs.GetVirusTotalVerdictByHash(metadata.FileSHA256) // true == safe
		if err != nil {
			return returnValue, err
		}
		// fmt.Println("[INFO] File Security check verdict for CID: " + metadata.IPFSCID + " with SHA256: " + metadata.FileSHA256)
		fresult = result
	}
	if fresult {
		fmt.Println("[INFO] Fetching CID: " + metadata.IPFSCID)
		// ipfs get
		psfs.GetFileIPFS(metadata.IPFSCID, appconfInstance.CacheFolder+"/"+metadata.IPFSCID+"."+metadata.FileType)
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

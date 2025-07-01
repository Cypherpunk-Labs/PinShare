package p2p

import (
	"fmt"
	"pinshare/internal/config"
	"pinshare/internal/psfs"
	"pinshare/internal/store"
)

func ProcessDownload(metadata store.BaseMetadata) (bool, error) {
	returnValue := false

	var fresult bool
	if config.FF_skip_vt {
		fresult = true
	} else {
		result, err := psfs.GetVirusTotalVerdictByHash(metadata.FileSHA256) // true == safe
		if err != nil {
			return returnValue, err
		}
		fmt.Println("[INFO] File Security check passed for CID: " + metadata.IPFSCID + " with SHA256: " + metadata.FileSHA256)
		fresult = result
	}
	if fresult {
		// ipfs get
		psfs.GetFileIPFS(metadata.IPFSCID, config.CacheFolder+"/"+metadata.IPFSCID+"."+metadata.FileType)
		// check file type
		ftype, err := psfs.ValidateFileType(config.CacheFolder + "/" + metadata.IPFSCID + "." + metadata.FileType)
		if err != nil {
			return returnValue, err
		}
		if ftype {
			psfs.PinFileIPFS(metadata.IPFSCID)
			returnValue = true
		}
	} else {
		fmt.Println("[ERROR] File Security check failed for CID: " + metadata.IPFSCID + " with SHA256: " + metadata.FileSHA256)
	}
	return returnValue, nil
}

package p2p

import (
	"fmt"
	"pinshare/internal/psfs"
	"pinshare/internal/store"
	"strings"
)

func ProcessUploads(folderPath string) {
	file, err := psfs.ListFiles(folderPath)
	var count int = 0
	if err != nil {
		return
	}
	for _, f := range file {
		ftype, err := psfs.ValidateFileType(folderPath + "/" + f)
		if err != nil {
			fmt.Println("[ERROR] func ValidateFileType() error " + string(err.Error()))
			return
		}
		if ftype {
			fmt.Println("[INFO] File type valid for file: " + f)
			fsha256, err := psfs.GetSHA256(folderPath + "/" + f)
			if err != nil {
				fmt.Println("[ERROR] func GetSha256() error " + string(err.Error()))
				return
			}

			var fresult bool
			if appconfInstance.FFSkipVT {
				fresult = true
			} else {
				result, err := psfs.GetVirusTotalVerdictByHash(fsha256) // true == safe
				if err != nil {
					fmt.Println("[ERROR] (GetVirusTotalVerdictByHash) " + string(err.Error()))
					return
				}
				fmt.Println("[INFO] File Security check passed for file: " + f + " with SHA256: " + fsha256)
				fresult = result
			}

			if appconfInstance.FFIgnoreUploadsInMetadata {
				_, exists := store.GlobalStore.GetFile(fsha256)
				if exists {
					fmt.Printf("[WARNING] File already exists in GlobalStore with SHA256: %s \n", fsha256)
					return
				}
			}

			if fresult {
				fcid := psfs.AddFileIPFS(folderPath + "/" + f)
				if fcid != "" {
					fmt.Println("[INFO] File: " + f + " ++added to IPFS with CID: " + fcid)
					fileExtension, err := psfs.GetExtension(f)
					if err != nil {
						return
					}

					metadata := store.BaseMetadata{
						FileSHA256: strings.ToLower(fsha256),
						IPFSCID:    strings.ToLower(fcid),
						FileType:   strings.ToLower(fileExtension),
					}

					errgs := store.GlobalStore.AddFile(metadata)
					if errgs != nil {
						fmt.Printf("[ERROR] failed to add file to GlobalStore: %w \n", errgs)
						return
					}
					fmt.Println("[INFO] File: " + f + " ++added to GlobalStore with CID: " + fcid)
					count = count + 1
					if appconfInstance.FFMoveUpload {
						err := psfs.MoveFile(folderPath+"/"+f, appconfInstance.CacheFolder+"/"+f)
						if err != nil {
							fmt.Println("[ERROR] Error moving file: ", err)
						}
					}
				}
			} else {
				fmt.Println("[ERROR] File Security check failed for file: " + f + " with SHA256: " + fsha256)
				if appconfInstance.FFSendFileVT {
					// TODO: uploadFile to VT here and wait for next loop
				}
			}
		} else {
			fmt.Println("[ERROR] File type invalid for file: " + f)
			if appconfInstance.FFMoveUpload {
				err := psfs.MoveFile(folderPath+"/"+f, appconfInstance.RejectFolder+"/"+f)
				if err != nil {
					fmt.Println("[ERROR] Error moving file: ", err)
				}
			}
			// move to rejected folder
			// log reason in rejected folder logfile
		}
	}
	if count >= 1 {
		store.GlobalStore.Save(appconfInstance.MetaDataFile)
	}
}

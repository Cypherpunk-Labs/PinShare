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
			if appconfInstance.FFIgnoreUploadsInMetadata {

				_, exists := store.GlobalStore.GetFile(fsha256)
				if exists {
					fmt.Printf("[WARNING] File already exists in GlobalStore with SHA256: %s \n", fsha256)
					return
				} else {

					if appconfInstance.SecurityCapability > 0 {
						fmt.Println("[INFO] File Security checking file: " + f + " with SHA256: " + fsha256)
						var result bool
						var err error
						// TODO: 				if appconfInstance.SecurityCapability [1 2 3 4]
						if appconfInstance.SecurityCapability <= 3 {
							result, err = psfs.ClamScanFileClean(folderPath + "/" + f)
							if err != nil {
								fmt.Println("[ERROR] (ClamScanFileClean) " + string(err.Error()))
								return
							}
						}

						if appconfInstance.SecurityCapability == 4 {
							if appconfInstance.FFSkipVT {
								result = true
							} else {
								result, err = psfs.GetVirusTotalWSVerdictByHash(fsha256) // true == safe
								if err != nil {
									fmt.Println("[ERROR] (GetVirusTotalVerdictByHash) " + string(err.Error()))
									return
								}
							}
						}

						// fmt.Println("[INFOSEC] File Security check passed for file: " + f + " with SHA256: " + fsha256)
						fresult = result
					}

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
				if appconfInstance.FFSendFileVT {
					// This was really to catch unknow files on VT

					// TODO: 				if appconfInstance.SecurityCapability [1 2 3 4]

					if appconfInstance.SecurityCapability == 4 {
						fmt.Println("[INFO] Submitting File to 3rd Party for Security check for file: " + f + " with SHA256: " + fsha256)
						submitresult, err := psfs.SendFileToVirusTotalWS(folderPath + "/" + f)
						if err != nil {
							fmt.Println("[ERROR] Error submitting file for security check: ", err)
						}
						if submitresult {
							fmt.Println("[INFO] Submission Passed Security check for file: " + f + " with SHA256: " + fsha256)
						} else {
							fmt.Println("[ERROR] File Security check failed for file: " + f + " with SHA256: " + fsha256)
							if appconfInstance.FFMoveUpload {
								err := psfs.MoveFile(folderPath+"/"+f, appconfInstance.RejectFolder+"/"+f)
								if err != nil {
									fmt.Println("[ERROR] Error moving file: ", err)
								}
							}
						}
					}
				} else {
					fmt.Println("[ERROR] File Security check failed for file: " + f + " with SHA256: " + fsha256)
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

package p2p

import (
	"fmt"
	"pinshare/internal/psfs"
	"pinshare/internal/store"
	"strings"
)

func ProcessUploads(folderPath string) bool {
	file, err := psfs.ListFiles(folderPath)
	var success bool = false
	var count int = 0
	if err != nil {
		return false
	}
	for _, f := range file {
		ftype, err := psfs.ValidateFileType(folderPath + "/" + f)
		if err != nil {
			fmt.Println("[ERROR] func ValidateFileType() error " + string(err.Error()))
			return false
		}
		if ftype {
			fmt.Println("[INFO] File type valid for file: " + f)
			fsha256, err := psfs.GetSHA256(folderPath + "/" + f)
			if err != nil {
				fmt.Println("[ERROR] func GetSha256() error " + string(err.Error()))
				return false
			}

			var fresult bool
			if appconfInstance.FFIgnoreUploadsInMetadata {

				_, exists := store.GlobalStore.GetFile(fsha256)
				if exists {
					fmt.Printf("[WARNING] File already exists in GlobalStore with SHA256: %s \n", fsha256)
					return false
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
								return false
							}
						}

						if appconfInstance.SecurityCapability == 4 {
							if appconfInstance.FFSkipVT {
								result = true
							} else {
								result, err = psfs.GetVirusTotalWSVerdictByHash(fsha256) // true == safe
								if err != nil {
									fmt.Println("[ERROR] (GetVirusTotalVerdictByHash) " + string(err.Error()))
									return false
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
						return false
					}

					var tagMap = make(map[string]int) // Potential for bug here to wipe all the tags since we init a blank map
					tagMap["filename:"+f] = 1         // we assume this file does not exist so do not seek to load existing tags. a check is done at line 33
					metadata := store.BaseMetadata{
						FileSHA256: strings.ToLower(fsha256),
						IPFSCID:    strings.ToLower(fcid),
						FileType:   strings.ToLower(fileExtension),
						Tags:       tagMap,
					}

					errgs := store.GlobalStore.AddFile(metadata)
					if errgs != nil {
						fmt.Printf("[ERROR] failed to add file to GlobalStore: %v\n", errgs)
						return false
					}
					fmt.Println("[INFO] File: " + f + " ++added to GlobalStore with CID: " + fcid)
					count = count + 1
					if appconfInstance.FFMoveUpload {
						err := psfs.MoveFile(folderPath+"/"+f, appconfInstance.CacheFolder+"/"+f)
						if err != nil {
							fmt.Println("[ERROR] Error moving file: ", err)
						}
					}
					success = true
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
					fmt.Println("[ERROR] File Security check, failed for file: " + f + " with SHA256: " + fsha256)
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
		success = true
	}
	return success
}

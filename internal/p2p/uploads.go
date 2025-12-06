package p2p

import (
	"fmt"
	"pinshare/internal/psfs"
	"pinshare/internal/store"
	"strings"
)

func ProcessUploads(folderPath string) {
	files, err := psfs.ListFiles(folderPath)
	if err != nil {
		return
	}

	var count int
	for _, f := range files {
		if processFile(folderPath, f) {
			count++
		}
	}

	if count >= 1 {
		store.GlobalStore.Save(appconfInstance.MetaDataFile)
	}
}

// processFile handles a single file upload. Returns true if the file was successfully added.
func processFile(folderPath, f string) bool {
	filePath := folderPath + "/" + f

	// Validate file type
	valid, err := psfs.ValidateFileType(filePath)
	if err != nil {
		fmt.Println("[ERROR] func ValidateFileType() error " + err.Error())
		return false
	}

	if !valid {
		handleInvalidFileType(folderPath, f)
		return false
	}

	fmt.Println("[INFO] File type valid for file: " + f)

	// Get file hash
	fsha256, err := psfs.GetSHA256(filePath)
	if err != nil {
		fmt.Println("[ERROR] func GetSha256() error " + err.Error())
		return false
	}

	// Check if file should be processed
	if !shouldProcessFile(fsha256) {
		return false
	}

	// Perform security scan
	scanPassed, err := performUploadSecurityScan(filePath, fsha256)
	if err != nil {
		return false
	}

	if scanPassed {
		return addFileToIPFS(folderPath, f, fsha256)
	}

	handleSecurityFailure(folderPath, f, fsha256)
	return false
}

// shouldProcessFile checks if a file should be processed based on metadata settings.
func shouldProcessFile(fsha256 string) bool {
	if !appconfInstance.FFIgnoreUploadsInMetadata {
		return true
	}

	_, exists := store.GlobalStore.GetFile(fsha256)
	if exists {
		fmt.Printf("[WARNING] File already exists in GlobalStore with SHA256: %s \n", fsha256)
		return false
	}
	return true
}

// performUploadSecurityScan scans a file for security threats.
func performUploadSecurityScan(filePath, fsha256 string) (bool, error) {
	capability := SecurityCapability(appconfInstance.SecurityCapability)

	if capability == SecurityCapabilityNone {
		return false, nil
	}

	fmt.Println("[INFO] File Security checking file: " + filePath + " with SHA256: " + fsha256)

	// Skip all security scanning if FFSkipVT is enabled
	if appconfInstance.FFSkipVT {
		fmt.Println("[INFO] Virus scanning disabled (FFSkipVT=true), skipping security check")
		return true, nil
	}

	switch {
	case capability.UsesClamAV():
		result, err := psfs.ClamScanFileClean(filePath)
		if err != nil {
			fmt.Println("[ERROR] (ClamScanFileClean) " + err.Error())
			return false, err
		}
		return result, nil

	case capability.UsesVirusTotalBrowser():
		result, err := psfs.GetVirusTotalWSVerdictByHash(fsha256)
		if err != nil {
			fmt.Println("[ERROR] (GetVirusTotalVerdictByHash) " + err.Error())
			return false, err
		}
		return result, nil

	default:
		fmt.Println("[ERROR] Unknown security capability: ", appconfInstance.SecurityCapability)
		return false, nil
	}
}

// addFileToIPFS adds a file to IPFS and the global store.
func addFileToIPFS(folderPath, f, fsha256 string) bool {
	filePath := folderPath + "/" + f

	fcid := psfs.AddFileIPFS(filePath)
	if fcid == "" {
		return false
	}

	fmt.Println("[INFO] File: " + f + " ++added to IPFS with CID: " + fcid)

	fileExtension, err := psfs.GetExtension(f)
	if err != nil {
		return false
	}

	metadata := store.BaseMetadata{
		FileSHA256: strings.ToLower(fsha256),
		IPFSCID:    strings.ToLower(fcid),
		FileType:   strings.ToLower(fileExtension),
	}

	if err := store.GlobalStore.AddFile(metadata); err != nil {
		fmt.Printf("[ERROR] failed to add file to GlobalStore: %v \n", err)
		return false
	}

	fmt.Println("[INFO] File: " + f + " ++added to GlobalStore with CID: " + fcid)

	if appconfInstance.FFMoveUpload {
		if err := psfs.MoveFile(filePath, appconfInstance.CacheFolder+"/"+f); err != nil {
			fmt.Println("[ERROR] Error moving file: ", err)
		}
	}

	return true
}

// handleSecurityFailure handles a file that failed security scanning.
func handleSecurityFailure(folderPath, f, fsha256 string) {
	filePath := folderPath + "/" + f
	capability := SecurityCapability(appconfInstance.SecurityCapability)

	// Try to submit to VirusTotal if enabled
	if appconfInstance.FFSendFileVT && capability.UsesVirusTotalBrowser() {
		fmt.Println("[INFO] Submitting File to 3rd Party for Security check for file: " + f + " with SHA256: " + fsha256)

		submitResult, err := psfs.SendFileToVirusTotalWS(filePath)
		if err != nil {
			fmt.Println("[ERROR] Error submitting file for security check: ", err)
		}

		if submitResult {
			fmt.Println("[INFO] Submission Passed Security check for file: " + f + " with SHA256: " + fsha256)
			return
		}

		fmt.Println("[ERROR] File Security check failed for file: " + f + " with SHA256: " + fsha256)
		moveToRejected(folderPath, f)
		return
	}

	fmt.Println("[ERROR] File Security check failed for file: " + f + " with SHA256: " + fsha256)
}

// handleInvalidFileType handles a file with an invalid type.
func handleInvalidFileType(folderPath, f string) {
	fmt.Println("[ERROR] File type invalid for file: " + f)
	moveToRejected(folderPath, f)
}

// moveToRejected moves a file to the rejected folder if FFMoveUpload is enabled.
func moveToRejected(folderPath, f string) {
	if !appconfInstance.FFMoveUpload {
		return
	}

	if err := psfs.MoveFile(folderPath+"/"+f, appconfInstance.RejectFolder+"/"+f); err != nil {
		fmt.Println("[ERROR] Error moving file: ", err)
	}
}

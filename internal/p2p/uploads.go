package p2p

import (
	"fmt"
	"path/filepath"
	"pinshare/internal/psfs"
	"pinshare/internal/store"
	"strings"
	"sync"
	"sync/atomic"
)

// maxConcurrentUploads limits the number of files processed in parallel
const maxConcurrentUploads = 4

// ProcessUploads processes all files in the given folder for upload to IPFS.
// Files are processed concurrently with a limit of maxConcurrentUploads simultaneous operations.
func ProcessUploads(folderPath string) {
	filenames, err := psfs.ListFiles(folderPath)
	if err != nil {
		return
	}

	var count int64
	// wg tracks completion of all file processing goroutines.
	// semaphore limits concurrent processing to maxConcurrentUploads to avoid
	// overwhelming system resources (file handles, network connections, etc.).
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, maxConcurrentUploads)

	for _, filename := range filenames {
		wg.Add(1)
		semaphore <- struct{}{} // Acquire semaphore slot (blocks if at capacity)

		go processFileWithLimit(folderPath, filename, semaphore, &wg, &count)
	}

	wg.Wait()

	if count >= 1 {
		store.GlobalStore.Save(appconfInstance.MetaDataFile)
	}
}

// processFileWithLimit wraps processFile with semaphore-based concurrency limiting.
// It releases the semaphore slot and marks the WaitGroup as done when processing completes.
func processFileWithLimit(folderPath, filename string, semaphore chan struct{}, wg *sync.WaitGroup, count *int64) {
	defer wg.Done()
	defer func() { <-semaphore }() // Release semaphore slot

	if processFile(folderPath, filename) {
		atomic.AddInt64(count, 1)
	}
}

// processFile handles a single file upload. Returns true if the file was successfully added.
func processFile(folderPath, filename string) bool {
	filePath := filepath.Join(folderPath, filename)

	// Validate file type
	valid, err := psfs.ValidateFileType(filePath)
	if err != nil {
		fmt.Printf("[ERROR] func ValidateFileType() error: %v\n", err)
		return false
	}

	if !valid {
		handleInvalidFileType(folderPath, filename)
		return false
	}

	fmt.Printf("[INFO] File type valid for file: %s\n", filename)

	// Get file hash
	fsha256, err := psfs.GetSHA256(filePath)
	if err != nil {
		fmt.Printf("[ERROR] func GetSha256() error: %v\n", err)
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
		return addFileToIPFS(folderPath, filename, fsha256)
	}

	handleSecurityFailure(folderPath, filename, fsha256)
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
	capability := appconfInstance.SecurityCapability

	if capability == SecurityCapabilityNone {
		return false, nil
	}

	fmt.Printf("[INFO] File Security checking file: %s with SHA256: %s\n", filePath, fsha256)

	// Skip all security scanning if FFSkipVT is enabled
	if appconfInstance.FFSkipVT {
		fmt.Println("[INFO] Virus scanning disabled (FFSkipVT=true), skipping security check")
		return true, nil
	}

	switch {
	case capability.UsesClamAV():
		result, err := psfs.ClamScanFileClean(filePath)
		if err != nil {
			fmt.Printf("[ERROR] (ClamScanFileClean) %v\n", err)
			return false, err
		}
		return result, nil

	case capability.UsesVirusTotalBrowser():
		result, err := psfs.GetVirusTotalWSVerdictByHash(fsha256)
		if err != nil {
			fmt.Printf("[ERROR] (GetVirusTotalVerdictByHash) %v\n", err)
			return false, err
		}
		return result, nil

	default:
		fmt.Printf("[ERROR] Unknown security capability: %d\n", appconfInstance.SecurityCapability)
		return false, nil
	}
}

// addFileToIPFS adds a file to IPFS and the global store.
func addFileToIPFS(folderPath, filename, fsha256 string) bool {
	filePath := filepath.Join(folderPath, filename)

	fcid := psfs.AddFileIPFS(filePath)
	if fcid == "" {
		return false
	}

	fmt.Printf("[INFO] File: %s ++added to IPFS with CID: %s\n", filename, fcid)

	fileExtension, err := psfs.GetExtension(filename)
	if err != nil {
		return false
	}

	metadata := store.BaseMetadata{
		FileSHA256: strings.ToLower(fsha256),
		IPFSCID:    strings.ToLower(fcid),
		FileType:   strings.ToLower(fileExtension),
	}

	if err := store.GlobalStore.AddFile(metadata); err != nil {
		fmt.Printf("[ERROR] failed to add file to GlobalStore: %v\n", err)
		return false
	}

	fmt.Printf("[INFO] File: %s ++added to GlobalStore with CID: %s\n", filename, fcid)

	if appconfInstance.FFMoveUpload {
		destPath := filepath.Join(appconfInstance.CacheFolder, filename)
		if err := psfs.MoveFile(filePath, destPath); err != nil {
			fmt.Printf("[ERROR] Error moving file: %v\n", err)
		}
	}

	return true
}

// handleSecurityFailure handles a file that failed security scanning.
func handleSecurityFailure(folderPath, filename, fsha256 string) {
	filePath := filepath.Join(folderPath, filename)
	capability := appconfInstance.SecurityCapability

	// Try to submit to VirusTotal if enabled
	if appconfInstance.FFSendFileVT && capability.UsesVirusTotalBrowser() {
		fmt.Printf("[INFO] Submitting File to 3rd Party for Security check for file: %s with SHA256: %s\n", filename, fsha256)

		submitResult, err := psfs.SendFileToVirusTotalWS(filePath)
		if err != nil {
			fmt.Printf("[ERROR] Error submitting file for security check: %v\n", err)
		}

		if submitResult {
			fmt.Printf("[INFO] Submission Passed Security check for file: %s with SHA256: %s\n", filename, fsha256)
			return
		}

		fmt.Printf("[ERROR] File Security check failed for file: %s with SHA256: %s\n", filename, fsha256)
		moveToRejected(folderPath, filename)
		return
	}

	fmt.Printf("[ERROR] File Security check failed for file: %s with SHA256: %s\n", filename, fsha256)
}

// handleInvalidFileType handles a file with an invalid type.
func handleInvalidFileType(folderPath, filename string) {
	fmt.Printf("[ERROR] File type invalid for file: %s\n", filename)
	moveToRejected(folderPath, filename)
}

// moveToRejected moves a file to the rejected folder if FFMoveUpload is enabled.
func moveToRejected(folderPath, filename string) {
	if !appconfInstance.FFMoveUpload {
		return
	}

	srcPath := filepath.Join(folderPath, filename)
	destPath := filepath.Join(appconfInstance.RejectFolder, filename)
	if err := psfs.MoveFile(srcPath, destPath); err != nil {
		fmt.Printf("[ERROR] Error moving file: %v\n", err)
	}
}

package psfs

import (
	"fmt"
	"io"
	"os"
)

// create func to list files in folder passed as parameters.
func ListFiles(folder string) ([]string, error) {
	var files []string
	entries, err := os.ReadDir(folder)
	if err != nil {
		return files, err
	}
	for _, entry := range entries {
		files = append(files, entry.Name())
	}
	return files, nil
}

func MoveFile(sourcePath, destPath string) error {
	// Attempt a direct rename first, as it's atomic and efficient.
	err := os.Rename(sourcePath, destPath)
	if err == nil {
		return nil
	}

	// If rename fails, fall back to a copy-and-delete operation.
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("could not open source file: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("could not create destination file: %w", err)
	}
	defer destFile.Close()

	if _, err = io.Copy(destFile, sourceFile); err != nil {
		// If copy fails, clean up the partially created destination file.
		os.Remove(destPath)
		return fmt.Errorf("failed to copy file contents: %w", err)
	}

	// Remove the original file after a successful copy.
	return os.Remove(sourcePath)
}

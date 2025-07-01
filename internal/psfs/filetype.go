package psfs

import (
	"fmt"
	"strings"

	"github.com/gabriel-vasile/mimetype"
)

func GetExtension(filepath string) (string, error) {
	pathSplit := strings.Split(filepath, ".")
	last := len(pathSplit) - 1
	value := strings.ToLower(pathSplit[last])
	if value != "" {
		return value, nil
	}
	return "", fmt.Errorf("no extension found")
}

func ValidateFileType(filepath string) (bool, error) {
	var validType = false

	fileExtension, err := GetExtension(filepath)
	if err != nil {
		return validType, err
	}

	if AllowedList[strings.ToLower(fileExtension)] {
		mtype, _ := mimetype.DetectFile(filepath)

		if strings.ToLower(mtype.Extension()) == "."+strings.ToLower(fileExtension) {
			validType = true
		} else {
			if mtype.Extension() == ".txt" {
				if fileExtension == "md" || fileExtension == "rtf" {
					// rtf wont trap here just md
					fmt.Println("plaintext detected")
					validType = true
				}
			} else {
				validType = false
			}
		}
	} else {
		validType = false
	}

	return validType, nil
}

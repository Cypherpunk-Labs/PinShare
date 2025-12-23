package main

import (
	"log"
	"syscall"
	"unsafe"
)

const (
	MB_YESNO        = 0x00000004
	MB_ICONQUESTION = 0x00000020
	IDYES           = 6
	IDNO            = 7
)

// showSettingsDialog launches the walk-based settings dialog.
// Returns true if settings were changed and saved, false if cancelled.
func showSettingsDialog() (changed bool, err error) {
	log.Println("Opening settings dialog...")

	saved, err := ShowSettingsDialogWalk()
	if err != nil {
		log.Printf("Settings dialog error: %v", err)
		return false, err
	}

	if saved {
		log.Println("Settings saved successfully")
	} else {
		log.Println("Settings dialog cancelled by user")
	}

	return saved, nil
}

// showConfirmDialog shows a Yes/No confirmation dialog and returns true if Yes was clicked.
func showConfirmDialog(title, message string) bool {
	titlePtr, _ := syscall.UTF16PtrFromString(title)
	messagePtr, _ := syscall.UTF16PtrFromString(message)

	ret, _, _ := procMessageBoxW.Call(
		0,
		uintptr(unsafe.Pointer(messagePtr)),
		uintptr(unsafe.Pointer(titlePtr)),
		uintptr(MB_YESNO|MB_ICONQUESTION),
	)

	return int(ret) == IDYES
}

package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"unsafe"
)

const (
	MB_YESNO        = 0x00000004
	MB_ICONQUESTION = 0x00000020
	IDYES           = 6
	IDNO            = 7
)

// showSettingsDialog launches the PowerShell settings dialog.
// Returns true if settings were changed and saved, false if cancelled.
func showSettingsDialog() (changed bool, err error) {
	// Get path to settings.ps1 (same directory as executable)
	exePath, err := os.Executable()
	if err != nil {
		return false, fmt.Errorf("failed to get executable path: %w", err)
	}

	scriptPath := filepath.Join(filepath.Dir(exePath), "resources", "settings.ps1")

	// Check if script exists
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return false, fmt.Errorf("settings script not found: %s", scriptPath)
	}

	log.Printf("Launching settings dialog from: %s", scriptPath)

	// Launch PowerShell with the settings script
	// -ExecutionPolicy Bypass: Allow running the script
	// -NoProfile: Don't load user profile (faster startup)
	// -File: Run the script file
	cmd := exec.Command("powershell.exe",
		"-ExecutionPolicy", "Bypass",
		"-NoProfile",
		"-File", scriptPath)

	// Don't attach to console (prevents black window flash)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}

	err = cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode := exitErr.ExitCode()
			switch exitCode {
			case 1:
				// Exit code 1 = user cancelled, not an error
				log.Println("Settings dialog cancelled by user")
				return false, nil
			case 2:
				// Exit code 2 = error occurred (already shown to user)
				log.Println("Settings dialog encountered an error")
				return false, nil
			default:
				return false, fmt.Errorf("settings dialog error: exit code %d", exitCode)
			}
		}
		return false, fmt.Errorf("failed to run settings dialog: %w", err)
	}

	// Exit code 0 = settings were saved successfully
	log.Println("Settings saved successfully")
	return true, nil
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

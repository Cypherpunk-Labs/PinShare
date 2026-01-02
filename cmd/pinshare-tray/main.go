package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"time"
	"unsafe"

	"github.com/getlantern/systray"
	"golang.org/x/sys/windows"
)

var (
	// appCtx is the application-wide context for graceful shutdown
	appCtx    context.Context
	appCancel context.CancelFunc
)

// SessionMarker contains user session information for the service to find user data
type SessionMarker struct {
	LocalAppData string    `json:"local_app_data"`
	Username     string    `json:"username"`
	Timestamp    time.Time `json:"timestamp"`
}

var (
	user32           = syscall.NewLazyDLL("user32.dll")
	procMessageBoxW  = user32.NewProc("MessageBoxW")

	// Package-level tray instance for access in onExit
	trayInstance *Tray
)

// ensureUserDataDirectories creates all required directories in user's LOCALAPPDATA
// and grants SYSTEM account full access so the Windows service can read/write them
func ensureUserDataDirectories() error {
	dataDir := getUserDataDirectory()

	dirs := []string{
		dataDir,
		filepath.Join(dataDir, dirIPFS),
		filepath.Join(dataDir, dirPinShare),
		filepath.Join(dataDir, dirUpload),
		filepath.Join(dataDir, dirCache),
		filepath.Join(dataDir, dirRejected),
		filepath.Join(dataDir, dirLogs),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	// Grant SYSTEM account full access to the data directory
	// This is required because the Windows service runs as SYSTEM
	if err := grantSystemAccess(dataDir); err != nil {
		log.Printf("Warning: Failed to grant SYSTEM access: %v", err)
		// Continue anyway - service might still work if permissions allow
	}

	log.Printf("User data directories ensured at: %s", dataDir)
	return nil
}

// grantSystemAccess uses icacls to grant the SYSTEM account full access to a directory
// This allows the Windows service (running as SYSTEM) to access user data
func grantSystemAccess(dir string) error {
	// Use icacls to grant SYSTEM full control with inheritance
	// /grant SYSTEM:(OI)(CI)F = Full control, Object Inherit, Container Inherit
	cmd := exec.Command("icacls", dir, "/grant", "SYSTEM:(OI)(CI)F", "/T", "/Q")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("icacls failed: %w\nOutput: %s", err, output)
	}
	log.Printf("Granted SYSTEM access to: %s", dir)
	return nil
}

// writeSessionMarker writes session info to ProgramData for the service to read
func writeSessionMarker() error {
	programData := os.Getenv(envProgramData)
	if programData == "" {
		programData = defaultProgramDataPath
	}

	// Ensure ProgramData\PinShare exists for the marker file
	markerDir := filepath.Join(programData, appName)
	if err := os.MkdirAll(markerDir, 0755); err != nil {
		return err
	}

	localAppData := os.Getenv(envLocalAppData)
	if localAppData == "" {
		userProfile := os.Getenv(envUserProfile)
		if userProfile != "" {
			localAppData = filepath.Join(userProfile, "AppData", "Local")
		}
	}

	marker := SessionMarker{
		LocalAppData: localAppData,
		Username:     os.Getenv(envUsername),
		Timestamp:    time.Now(),
	}

	data, err := json.MarshalIndent(marker, "", "  ")
	if err != nil {
		return err
	}

	markerPath := filepath.Join(markerDir, fileSession)
	if err := os.WriteFile(markerPath, data, 0644); err != nil {
		return err
	}

	log.Printf("Session marker written: %s (user: %s)", markerPath, marker.Username)
	return nil
}

func main() {
	// Ensure we're running on Windows
	if runtime.GOOS != "windows" {
		log.Fatal("This application is only supported on Windows")
	}

	systray.Run(onReady, onExit)
}

func onReady() {
	// Create application context for graceful shutdown
	appCtx, appCancel = context.WithCancel(context.Background())

	// Ensure user data directories exist (in LOCALAPPDATA)
	if err := ensureUserDataDirectories(); err != nil {
		log.Printf("Warning: Failed to create data directories: %v", err)
	}

	// Write session marker so service knows where user data is
	if err := writeSessionMarker(); err != nil {
		log.Printf("Warning: Failed to write session marker: %v", err)
	}

	// Set up the tray icon
	iconData, err := loadIcon()
	if err != nil {
		log.Printf("Warning: Failed to load icon: %v", err)
		// Use a simple default icon
		systray.SetIcon(getDefaultIcon())
	} else {
		systray.SetIcon(iconData)
	}

	systray.SetTitle(appName)
	systray.SetTooltip(appTooltip)

	// Create tray instance and store in package-level variable
	trayInstance = NewTray()

	// Build menu
	trayInstance.BuildMenu()

	// Start menu click handler with context for graceful shutdown
	trayInstance.StartMenuHandler(appCtx)

	// Ensure service is running (start if stopped)
	trayInstance.ensureServiceRunning()

	// Start status update loop
	go trayInstance.UpdateStatusLoop()
}

func onExit() {
	log.Println("PinShare tray application exiting")

	// Cancel the application context to signal shutdown to goroutines
	if appCancel != nil {
		appCancel()
	}

	// Stop the service when tray exits
	if err := stopService(); err != nil {
		log.Printf("Failed to stop service on exit: %v", err)
	} else {
		log.Println("Service stopped")
	}
}

// loadIcon loads the application icon
func loadIcon() ([]byte, error) {
	exePath, err := os.Executable()
	if err != nil {
		return nil, err
	}

	iconPath := filepath.Join(filepath.Dir(exePath), "icon.ico")
	data, err := os.ReadFile(iconPath)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// getDefaultIcon returns a valid 16x16 blue square icon in ICO format
func getDefaultIcon() []byte {
	// This is a valid ICO file with a 16x16 32-bit RGBA blue icon
	iconData := make([]byte, 0, 1150)

	// ICO Header (6 bytes)
	iconData = append(iconData,
		0x00, 0x00, // Reserved
		0x01, 0x00, // Type: 1 = ICO
		0x01, 0x00, // Count: 1 image
	)

	// ICO Directory Entry (16 bytes)
	iconData = append(iconData,
		0x10,       // Width: 16
		0x10,       // Height: 16
		0x00,       // Color palette: 0 = no palette
		0x00,       // Reserved
		0x01, 0x00, // Color planes: 1
		0x20, 0x00, // Bits per pixel: 32
		0x68, 0x04, 0x00, 0x00, // Size of image data: 1128 bytes
		0x16, 0x00, 0x00, 0x00, // Offset to image data: 22
	)

	// BMP Info Header (40 bytes)
	iconData = append(iconData,
		0x28, 0x00, 0x00, 0x00, // Header size: 40
		0x10, 0x00, 0x00, 0x00, // Width: 16
		0x20, 0x00, 0x00, 0x00, // Height: 32 (16 * 2 for XOR + AND masks)
		0x01, 0x00, // Planes: 1
		0x20, 0x00, // Bits per pixel: 32
		0x00, 0x00, 0x00, 0x00, // Compression: none
		0x00, 0x04, 0x00, 0x00, // Image size: 1024
		0x00, 0x00, 0x00, 0x00, // X pixels per meter
		0x00, 0x00, 0x00, 0x00, // Y pixels per meter
		0x00, 0x00, 0x00, 0x00, // Colors used
		0x00, 0x00, 0x00, 0x00, // Important colors
	)

	// Pixel data: 16x16 pixels, BGRA format, bottom-up
	for row := 0; row < 16; row++ {
		for col := 0; col < 16; col++ {
			y := 15 - row
			isEdge := col == 0 || col == 15 || y == 0 || y == 15
			isInner := col >= 2 && col <= 13 && y >= 2 && y <= 13

			if isEdge {
				iconData = append(iconData, 0x80, 0x40, 0x00, 0xFF) // Dark blue
			} else if isInner {
				iconData = append(iconData, 0xFF, 0x99, 0x33, 0xFF) // Bright blue
			} else {
				iconData = append(iconData, 0x00, 0x00, 0x00, 0x00) // Transparent
			}
		}
	}

	// AND mask
	for i := 0; i < 16; i++ {
		iconData = append(iconData, 0x00, 0x00, 0x00, 0x00)
	}

	return iconData
}

// openBrowser opens a URL or path using the Windows shell
func openBrowser(url string) error {
	urlPtr, err := windows.UTF16PtrFromString(url)
	if err != nil {
		return err
	}

	// ShellExecute with nil verb uses the default action (open)
	return windows.ShellExecute(0, nil, urlPtr, nil, nil, windows.SW_SHOWNORMAL)
}

// showMessage shows a Windows message box
func showMessage(title, message string) {
	log.Printf("%s: %s", title, message)
	showMessageBox(title, message, MB_OK|MB_ICONINFORMATION)
}

// showError shows a Windows error message box
func showError(title, message string) {
	log.Printf("ERROR - %s: %s", title, message)
	showMessageBox(title, message, MB_OK|MB_ICONERROR)
}

// showMessageBox displays a Windows MessageBox
func showMessageBox(title, message string, flags uint32) {
	titlePtr, _ := syscall.UTF16PtrFromString(title)
	messagePtr, _ := syscall.UTF16PtrFromString(message)
	procMessageBoxW.Call(
		0,
		uintptr(unsafe.Pointer(messagePtr)),
		uintptr(unsafe.Pointer(titlePtr)),
		uintptr(flags),
	)
}

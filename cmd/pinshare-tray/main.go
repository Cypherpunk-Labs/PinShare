package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/getlantern/systray"
)

func main() {
	// Ensure we're running on Windows
	if runtime.GOOS != "windows" {
		log.Fatal("This application is only supported on Windows")
	}

	systray.Run(onReady, onExit)
}

func onReady() {
	// Set up the tray icon
	iconData, err := loadIcon()
	if err != nil {
		log.Printf("Warning: Failed to load icon: %v", err)
		// Use a simple default icon
		systray.SetIcon(getDefaultIcon())
	} else {
		systray.SetIcon(iconData)
	}

	systray.SetTitle("PinShare")
	systray.SetTooltip("PinShare - Decentralized IPFS Pinning")

	// Create tray instance
	tray := NewTray()

	// Build menu
	tray.BuildMenu()

	// Start status update loop
	go tray.UpdateStatusLoop()
}

func onExit() {
	// Cleanup
	log.Println("PinShare tray application exiting")
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

// getDefaultIcon returns a simple default icon (1x1 pixel)
func getDefaultIcon() []byte {
	// A simple ICO file with a 16x16 icon
	return []byte{
		0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x10, 0x10, 0x00, 0x00, 0x01, 0x00,
		0x20, 0x00, 0x68, 0x04, 0x00, 0x00, 0x16, 0x00, 0x00, 0x00, 0x28, 0x00,
		0x00, 0x00, 0x10, 0x00, 0x00, 0x00, 0x20, 0x00, 0x00, 0x00, 0x01, 0x00,
		0x20, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00,
	}
}

// openBrowser opens a URL in the default browser
func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}

	return cmd.Start()
}

// showMessage shows a system notification
func showMessage(title, message string) {
	// On Windows, we can use systray tooltips or external notification tools
	// For now, just log it
	log.Printf("%s: %s", title, message)

	// Update tooltip temporarily
	systray.SetTooltip(fmt.Sprintf("PinShare - %s", message))
}

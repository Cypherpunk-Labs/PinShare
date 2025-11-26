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

// getDefaultIcon returns a valid 16x16 blue square icon in ICO format
func getDefaultIcon() []byte {
	// This is a valid ICO file with a 16x16 32-bit RGBA blue icon
	// ICO Header: 6 bytes
	// ICO Directory Entry: 16 bytes
	// BMP Info Header: 40 bytes
	// Pixel Data: 16x16x4 = 1024 bytes (BGRA format, bottom-up)
	// AND Mask: 16x2 = 64 bytes (1-bit per pixel, padded to DWORD)

	// Pre-generated valid ICO file data for a blue square icon
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
	// Create a blue "P" shape on transparent background
	for row := 0; row < 16; row++ {
		for col := 0; col < 16; col++ {
			// Flip row for bottom-up format
			y := 15 - row

			// Draw a simple "P" shape or filled square with border
			isEdge := col == 0 || col == 15 || y == 0 || y == 15
			isInner := col >= 2 && col <= 13 && y >= 2 && y <= 13

			if isEdge {
				// Dark blue border: BGRA
				iconData = append(iconData, 0x80, 0x40, 0x00, 0xFF) // Dark blue
			} else if isInner {
				// Light blue fill: BGRA
				iconData = append(iconData, 0xFF, 0x99, 0x33, 0xFF) // Bright blue
			} else {
				// Transparent
				iconData = append(iconData, 0x00, 0x00, 0x00, 0x00)
			}
		}
	}

	// AND mask: 16 rows, each row is 2 bytes (16 bits) + 2 bytes padding = 4 bytes
	// All 0s = fully opaque (when combined with 32-bit alpha)
	for i := 0; i < 16; i++ {
		iconData = append(iconData, 0x00, 0x00, 0x00, 0x00)
	}

	return iconData
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

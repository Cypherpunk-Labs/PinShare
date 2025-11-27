package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/getlantern/systray"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

// contains checks if s contains substr (case-insensitive)
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

const (
	serviceName = "PinShareService"
	uiPort      = 8888 // Default UI port
)

type Tray struct {
	// Menu items
	menuOpenUI        *systray.MenuItem
	menuStatus        *systray.MenuItem
	menuIPFSStatus    *systray.MenuItem
	menuPinShareStatus *systray.MenuItem
	menuPeersStatus   *systray.MenuItem
	menuSeparator1    *systray.MenuItem
	menuStart         *systray.MenuItem
	menuStop          *systray.MenuItem
	menuRestart       *systray.MenuItem
	menuSeparator2    *systray.MenuItem
	menuSettings      *systray.MenuItem
	menuLogs          *systray.MenuItem
	menuAbout         *systray.MenuItem
	menuSeparator3    *systray.MenuItem
	menuExit          *systray.MenuItem

	// State
	serviceRunning bool
	lastError      error
}

func NewTray() *Tray {
	return &Tray{}
}

// BuildMenu creates the tray menu
func (t *Tray) BuildMenu() {
	// Open UI
	t.menuOpenUI = systray.AddMenuItem("Open PinShare UI", "Open the PinShare web interface")

	systray.AddSeparator()

	// Status
	t.menuStatus = systray.AddMenuItem("Status: Checking...", "Service status")
	t.menuStatus.Disable()

	t.menuIPFSStatus = systray.AddMenuItem("  IPFS: Unknown", "IPFS daemon status")
	t.menuIPFSStatus.Disable()

	t.menuPinShareStatus = systray.AddMenuItem("  PinShare: Unknown", "PinShare backend status")
	t.menuPinShareStatus.Disable()

	t.menuPeersStatus = systray.AddMenuItem("  Peers: Unknown", "Connected peers")
	t.menuPeersStatus.Disable()

	systray.AddSeparator()

	// Service control
	t.menuStart = systray.AddMenuItem("Start Service", "Start the PinShare service")
	t.menuStop = systray.AddMenuItem("Stop Service", "Stop the PinShare service")
	t.menuRestart = systray.AddMenuItem("Restart Service", "Restart the PinShare service")

	systray.AddSeparator()

	// Settings and logs
	t.menuSettings = systray.AddMenuItem("Settings...", "Open settings")
	t.menuLogs = systray.AddMenuItem("View Logs...", "Open log directory")

	systray.AddSeparator()

	// About
	t.menuAbout = systray.AddMenuItem("About PinShare", "About this application")

	systray.AddSeparator()

	// Exit
	t.menuExit = systray.AddMenuItem("Exit", "Exit the PinShare tray application")

	// Handle menu clicks
	go t.handleMenuClicks()

	// Initial status check
	t.updateStatus()
}

// handleMenuClicks handles menu item clicks
func (t *Tray) handleMenuClicks() {
	for {
		select {
		case <-t.menuOpenUI.ClickedCh:
			t.handleOpenUI()

		case <-t.menuStart.ClickedCh:
			t.handleStartService()

		case <-t.menuStop.ClickedCh:
			t.handleStopService()

		case <-t.menuRestart.ClickedCh:
			t.handleRestartService()

		case <-t.menuSettings.ClickedCh:
			t.handleSettings()

		case <-t.menuLogs.ClickedCh:
			t.handleViewLogs()

		case <-t.menuAbout.ClickedCh:
			t.handleAbout()

		case <-t.menuExit.ClickedCh:
			systray.Quit()
			return
		}
	}
}

// handleOpenUI opens the PinShare UI in browser
func (t *Tray) handleOpenUI() {
	url := fmt.Sprintf("http://localhost:%d", uiPort)
	if err := openBrowser(url); err != nil {
		log.Printf("Failed to open browser: %v", err)
		showMessage("Error", "Failed to open browser")
	}
}

// handleStartService starts the service
func (t *Tray) handleStartService() {
	if err := startService(); err != nil {
		log.Printf("Failed to start service: %v", err)
		showError("PinShare", fmt.Sprintf("Failed to start service:\n\n%v\n\nNote: You may need to run as Administrator.", err))
	} else {
		showMessage("PinShare", "Service started successfully.")
		time.Sleep(1 * time.Second)
		t.updateStatus()
	}
}

// handleStopService stops the service
func (t *Tray) handleStopService() {
	if err := controlService(svc.Stop); err != nil {
		log.Printf("Failed to stop service: %v", err)
		showError("PinShare", fmt.Sprintf("Failed to stop service:\n\n%v\n\nNote: You may need to run as Administrator.", err))
	} else {
		showMessage("PinShare", "Service stopped successfully.")
		time.Sleep(1 * time.Second)
		t.updateStatus()
	}
}

// handleRestartService restarts the service
func (t *Tray) handleRestartService() {
	// Stop first
	if err := controlService(svc.Stop); err != nil {
		log.Printf("Failed to stop service: %v", err)
		showError("PinShare", fmt.Sprintf("Failed to stop service:\n\n%v\n\nNote: You may need to run as Administrator.", err))
		return
	}

	// Wait a bit
	time.Sleep(2 * time.Second)

	// Start again
	if err := startService(); err != nil {
		log.Printf("Failed to start service: %v", err)
		showError("PinShare", fmt.Sprintf("Failed to start service:\n\n%v", err))
	} else {
		showMessage("PinShare", "Service restarted successfully.")
		time.Sleep(1 * time.Second)
		t.updateStatus()
	}
}

// handleSettings opens settings (placeholder)
func (t *Tray) handleSettings() {
	showMessage("Settings", "Settings UI not yet implemented")
	// TODO: Implement settings dialog
}

// handleViewLogs opens the log directory
func (t *Tray) handleViewLogs() {
	// Get data directory
	programData := "C:\\ProgramData"
	logDir := fmt.Sprintf("%s\\PinShare\\logs", programData)

	if err := openBrowser(logDir); err != nil {
		log.Printf("Failed to open log directory: %v", err)
		showMessage("Error", "Failed to open log directory")
	}
}

// handleAbout shows about information
func (t *Tray) handleAbout() {
	showMessage("About PinShare", "PinShare - Decentralized IPFS Pinning Service\nVersion 1.0")
}

// UpdateStatusLoop periodically updates the status
func (t *Tray) UpdateStatusLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		t.updateStatus()
	}
}

// updateStatus updates the service status
func (t *Tray) updateStatus() {
	status, err := getServiceStatus()
	if err != nil {
		t.lastError = err
		t.serviceRunning = false

		// Check if it's a "service not found" error
		errStr := err.Error()
		if contains(errStr, "not found") || contains(errStr, "does not exist") ||
		   contains(errStr, "specified service") || contains(errStr, "Access is denied") ||
		   contains(errStr, "OpenService") {
			t.menuStatus.SetTitle("Status: Not Installed")
			systray.SetTooltip("PinShare - Service not installed or access denied")
		} else {
			// Show actual error for debugging
			t.menuStatus.SetTitle("Status: Error")
			shortErr := errStr
			if len(shortErr) > 50 {
				shortErr = shortErr[:50] + "..."
			}
			systray.SetTooltip(fmt.Sprintf("PinShare - %s", shortErr))
		}

		t.menuIPFSStatus.SetTitle("  IPFS: -")
		t.menuPinShareStatus.SetTitle("  PinShare: -")
		t.menuPeersStatus.SetTitle("  Peers: -")

		// Enable start (to allow install attempt), disable stop
		t.menuStart.Enable()
		t.menuStop.Disable()
		t.menuRestart.Disable()
		return
	}

	switch status {
	case svc.Running:
		t.serviceRunning = true
		t.menuStatus.SetTitle("Status: Running ✓")

		// Update icon/tooltip
		systray.SetTooltip("PinShare - Running")

		// Enable stop/restart, disable start
		t.menuStart.Disable()
		t.menuStop.Enable()
		t.menuRestart.Enable()

		// TODO: Query actual IPFS/PinShare health
		t.menuIPFSStatus.SetTitle("  IPFS: Online")
		t.menuPinShareStatus.SetTitle("  PinShare: Online")
		t.menuPeersStatus.SetTitle("  Peers: Connected")

	case svc.Stopped:
		t.serviceRunning = false
		t.menuStatus.SetTitle("Status: Stopped")
		t.menuIPFSStatus.SetTitle("  IPFS: Offline")
		t.menuPinShareStatus.SetTitle("  PinShare: Offline")
		t.menuPeersStatus.SetTitle("  Peers: None")

		// Enable start, disable stop
		t.menuStart.Enable()
		t.menuStop.Disable()
		t.menuRestart.Disable()

		systray.SetTooltip("PinShare - Stopped")

	case svc.StartPending:
		t.menuStatus.SetTitle("Status: Starting...")
		t.menuStart.Disable()
		t.menuStop.Disable()
		t.menuRestart.Disable()
		systray.SetTooltip("PinShare - Starting...")

	case svc.StopPending:
		t.menuStatus.SetTitle("Status: Stopping...")
		t.menuStart.Disable()
		t.menuStop.Disable()
		t.menuRestart.Disable()
		systray.SetTooltip("PinShare - Stopping...")

	default:
		t.menuStatus.SetTitle(fmt.Sprintf("Status: Unknown (%d)", status))
		systray.SetTooltip("PinShare - Unknown status")
	}
}

// getServiceStatus gets the current service status
func getServiceStatus() (svc.State, error) {
	manager, err := mgr.Connect()
	if err != nil {
		return svc.Stopped, fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer manager.Disconnect()

	service, err := manager.OpenService(serviceName)
	if err != nil {
		return svc.Stopped, fmt.Errorf("failed to open service: %w", err)
	}
	defer service.Close()

	status, err := service.Query()
	if err != nil {
		return svc.Stopped, fmt.Errorf("failed to query service: %w", err)
	}

	return status.State, nil
}

// startService starts the service
func startService() error {
	manager, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer manager.Disconnect()

	service, err := manager.OpenService(serviceName)
	if err != nil {
		return fmt.Errorf("failed to open service: %w", err)
	}
	defer service.Close()

	return service.Start()
}

// controlService sends a control command to the service
func controlService(cmd svc.Cmd) error {
	manager, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer manager.Disconnect()

	service, err := manager.OpenService(serviceName)
	if err != nil {
		return fmt.Errorf("failed to open service: %w", err)
	}
	defer service.Close()

	_, err = service.Control(cmd)
	return err
}

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"pinshare/internal/winservice"
	"github.com/getlantern/systray"
	"golang.org/x/sys/windows"
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
	// TODO: Re-enable when UI is ready
	// // Open UI
	// t.menuOpenUI = systray.AddMenuItem("Open PinShare UI", "Open the PinShare web interface")
	//
	// systray.AddSeparator()

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
		// TODO: Re-enable when UI is ready
		// case <-t.menuOpenUI.ClickedCh:
		// 	t.handleOpenUI()

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

// TODO: Re-enable when UI is ready
// // handleOpenUI opens the PinShare UI in browser
// func (t *Tray) handleOpenUI() {
// 	url := fmt.Sprintf("http://localhost:%d", uiPort)
// 	if err := openBrowser(url); err != nil {
// 		log.Printf("Failed to open browser: %v", err)
// 		showMessage("Error", "Failed to open browser")
// 	}
// }

// handleStartService starts the service
func (t *Tray) handleStartService() {
	if err := startService(); err != nil {
		log.Printf("Failed to start service: %v", err)
		showError("PinShare", fmt.Sprintf("Failed to start service:\n\n%v", err))
	} else {
		time.Sleep(1 * time.Second)
		t.updateStatus()
	}
}

// handleStopService stops the service
func (t *Tray) handleStopService() {
	if err := stopService(); err != nil {
		log.Printf("Failed to stop service: %v", err)
		showError("PinShare", fmt.Sprintf("Failed to stop service:\n\n%v", err))
	} else {
		time.Sleep(1 * time.Second)
		t.updateStatus()
	}
}

// handleRestartService restarts the service
func (t *Tray) handleRestartService() {
	// Stop first
	if err := stopService(); err != nil {
		log.Printf("Failed to stop service: %v", err)
		showError("PinShare", fmt.Sprintf("Failed to stop service:\n\n%v", err))
		return
	}

	// Wait a bit
	time.Sleep(2 * time.Second)

	// Start again
	if err := startService(); err != nil {
		log.Printf("Failed to start service: %v", err)
		showError("PinShare", fmt.Sprintf("Failed to start service:\n\n%v", err))
	} else {
		time.Sleep(1 * time.Second)
		t.updateStatus()
	}
}

// handleSettings opens the settings dialog
func (t *Tray) handleSettings() {
	changed, err := showSettingsDialog()
	if err != nil {
		log.Printf("Settings dialog error: %v", err)
		showError("Settings Error", fmt.Sprintf("Failed to open settings:\n\n%v", err))
		return
	}

	if changed {
		// Reload config to pick up new port settings
		reloadConfig()

		// Ask user if they want to restart the service to apply changes
		if showConfirmDialog(
			"Restart Service?",
			"Settings have been saved.\n\n"+
				"The service must be restarted for changes to take effect.\n\n"+
				"Restart the service now?") {
			t.handleRestartService()
		}
	}
}

// handleViewLogs opens the log directory
func (t *Tray) handleViewLogs() {
	// Get data directory from environment or default
	programData := os.Getenv("PROGRAMDATA")
	if programData == "" {
		programData = "C:\\ProgramData"
	}
	logDir := fmt.Sprintf("%s\\PinShare\\logs", programData)

	if err := openBrowser(logDir); err != nil {
		log.Printf("Failed to open log directory: %v", err)
		showMessage("Error", "Failed to open log directory")
	}
}

// handleAbout shows about information
func (t *Tray) handleAbout() {
	showMessage("About PinShare",
		"PinShare - Decentralized IPFS Pinning Service\n"+
			"Version 1.0\n\n"+
			"GitHub: https://github.com/Cypherpunk-Labs/PinShare")
}

// UpdateStatusLoop periodically updates the status
func (t *Tray) UpdateStatusLoop() {
	ticker := time.NewTicker(winservice.StatusCheckInterval)
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

		// Check if it's a "service not installed" error
		if status == winservice.StateNotInstalled {
			t.menuStatus.SetTitle("Status: Not Installed")
			systray.SetTooltip("PinShare - Service not installed")
		} else {
			// Show actual error for debugging
			t.menuStatus.SetTitle("Status: Error")
			systray.SetTooltip(fmt.Sprintf("PinShare - %s", truncateErrorMessage(err.Error())))
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
	case winservice.StateRunning:
		t.updateStatusRunning()
	case winservice.StateStopped:
		t.updateStatusStopped()
	case winservice.StateStartPending:
		t.updateStatusStartPending()
	case winservice.StateStopPending:
		t.updateStatusStopPending()
	case winservice.StateNotInstalled:
		t.updateStatusNotInstalled()
	default:
		t.menuStatus.SetTitle(fmt.Sprintf("Status: Unknown (%s)", status))
		systray.SetTooltip("PinShare - Unknown status")
	}
}

// updateStatusRunning updates UI for running service state
func (t *Tray) updateStatusRunning() {
	t.serviceRunning = true
	t.menuStart.Disable()
	t.menuStop.Enable()
	t.menuRestart.Enable()

	// Check actual component health via HTTP
	ipfsHealthy := checkIPFSHealth()
	pinshareHealthy := checkPinShareHealth()

	if ipfsHealthy {
		t.menuIPFSStatus.SetTitle("  IPFS: Online")
	} else {
		t.menuIPFSStatus.SetTitle("  IPFS: Starting...")
	}

	if pinshareHealthy {
		t.menuPinShareStatus.SetTitle("  PinShare: Online")
		t.menuStatus.SetTitle("Status: Running")
		systray.SetTooltip("PinShare - Running")
	} else {
		t.menuPinShareStatus.SetTitle("  PinShare: Starting...")
		t.menuStatus.SetTitle("Status: Starting...")
		systray.SetTooltip("PinShare - Components starting...")
	}

	if ipfsHealthy && pinshareHealthy {
		t.menuPeersStatus.SetTitle("  Peers: Connected")
	} else {
		t.menuPeersStatus.SetTitle("  Peers: Connecting...")
	}
}

// updateStatusStopped updates UI for stopped service state
func (t *Tray) updateStatusStopped() {
	t.serviceRunning = false
	t.menuStatus.SetTitle("Status: Stopped")
	t.menuIPFSStatus.SetTitle("  IPFS: Offline")
	t.menuPinShareStatus.SetTitle("  PinShare: Offline")
	t.menuPeersStatus.SetTitle("  Peers: None")
	t.menuStart.Enable()
	t.menuStop.Disable()
	t.menuRestart.Disable()
	systray.SetTooltip("PinShare - Stopped")
}

// updateStatusStartPending updates UI for service start pending state
func (t *Tray) updateStatusStartPending() {
	t.menuStatus.SetTitle("Status: Starting...")
	t.menuIPFSStatus.SetTitle("  IPFS: Starting...")
	t.menuPinShareStatus.SetTitle("  PinShare: Starting...")
	t.menuPeersStatus.SetTitle("  Peers: Connecting...")
	t.menuStart.Disable()
	t.menuStop.Disable()
	t.menuRestart.Disable()
	systray.SetTooltip("PinShare - Starting...")
}

// updateStatusStopPending updates UI for service stop pending state
func (t *Tray) updateStatusStopPending() {
	t.menuStatus.SetTitle("Status: Stopping...")
	t.menuStart.Disable()
	t.menuStop.Disable()
	t.menuRestart.Disable()
	systray.SetTooltip("PinShare - Stopping...")
}

// updateStatusNotInstalled updates UI for service not installed state
func (t *Tray) updateStatusNotInstalled() {
	t.serviceRunning = false
	t.menuStatus.SetTitle("Status: Not Installed")
	t.menuIPFSStatus.SetTitle("  IPFS: -")
	t.menuPinShareStatus.SetTitle("  PinShare: -")
	t.menuPeersStatus.SetTitle("  Peers: -")
	t.menuStart.Enable()
	t.menuStop.Disable()
	t.menuRestart.Disable()
	systray.SetTooltip("PinShare - Service not installed")
}

// getServiceStatus gets the current service status using Windows Service Manager API (no spawned process)
// Uses minimal permissions (SC_MANAGER_CONNECT and SERVICE_QUERY_STATUS) so no elevation is required.
func getServiceStatus() (winservice.ServiceState, error) {
	// Open service control manager with minimal permissions (connect only)
	scmHandle, err := windows.OpenSCManager(nil, nil, windows.SC_MANAGER_CONNECT)
	if err != nil {
		return winservice.StateStopped, fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer windows.CloseServiceHandle(scmHandle)

	// Open the service with query status permission only
	serviceNamePtr, err := windows.UTF16PtrFromString(winservice.ServiceName)
	if err != nil {
		return winservice.StateStopped, fmt.Errorf("invalid service name: %w", err)
	}

	svcHandle, err := windows.OpenService(scmHandle, serviceNamePtr, windows.SERVICE_QUERY_STATUS)
	if err != nil {
		// Service doesn't exist (ERROR_SERVICE_DOES_NOT_EXIST = 1060)
		return winservice.StateNotInstalled, fmt.Errorf("service not installed")
	}
	defer windows.CloseServiceHandle(svcHandle)

	// Query the service status
	var status windows.SERVICE_STATUS
	err = windows.QueryServiceStatus(svcHandle, &status)
	if err != nil {
		return winservice.StateStopped, fmt.Errorf("failed to query service: %w", err)
	}

	// Map Windows service state to our ServiceState
	switch status.CurrentState {
	case windows.SERVICE_RUNNING:
		return winservice.StateRunning, nil
	case windows.SERVICE_STOPPED:
		return winservice.StateStopped, nil
	case windows.SERVICE_START_PENDING:
		return winservice.StateStartPending, nil
	case windows.SERVICE_STOP_PENDING:
		return winservice.StateStopPending, nil
	case windows.SERVICE_PAUSED, windows.SERVICE_PAUSE_PENDING, windows.SERVICE_CONTINUE_PENDING:
		return winservice.StateStopped, nil
	default:
		return winservice.StateStopped, fmt.Errorf("unknown service state: %d", status.CurrentState)
	}
}

// checkIPFSHealth checks if IPFS daemon is responding
func checkIPFSHealth() bool {
	config := getConfig()
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	// IPFS version endpoint requires POST
	url := fmt.Sprintf("http://localhost:%d/api/v0/version", config.IPFSAPIPort)
	resp, err := client.Post(url, "application/json", nil)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// checkPinShareHealth checks if PinShare API is responding
func checkPinShareHealth() bool {
	config := getConfig()
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	url := fmt.Sprintf("http://localhost:%d/api/health", config.PinShareAPIPort)
	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// startService starts the service using Windows API (no UAC required if DACL is set)
func startService() error {
	log.Printf("Starting service %s...", winservice.ServiceName)

	// Open service control manager with minimal permissions
	scmHandle, err := windows.OpenSCManager(nil, nil, windows.SC_MANAGER_CONNECT)
	if err != nil {
		return fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer windows.CloseServiceHandle(scmHandle)

	// Open service with start permission
	serviceNamePtr, _ := windows.UTF16PtrFromString(winservice.ServiceName)
	svcHandle, err := windows.OpenService(scmHandle, serviceNamePtr, windows.SERVICE_START|windows.SERVICE_QUERY_STATUS)
	if err != nil {
		return fmt.Errorf("failed to open service: %w", err)
	}
	defer windows.CloseServiceHandle(svcHandle)

	// Check if already running
	var status windows.SERVICE_STATUS
	if err := windows.QueryServiceStatus(svcHandle, &status); err == nil {
		if status.CurrentState == windows.SERVICE_RUNNING {
			log.Printf("Service already running")
			return nil
		}
	}

	// Start the service
	err = windows.StartService(svcHandle, 0, nil)
	if err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}

	log.Printf("Service start initiated")
	return nil
}

// stopService stops the service using Windows API (no UAC required if DACL is set)
func stopService() error {
	log.Printf("Stopping service %s...", winservice.ServiceName)

	scmHandle, err := windows.OpenSCManager(nil, nil, windows.SC_MANAGER_CONNECT)
	if err != nil {
		return fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer windows.CloseServiceHandle(scmHandle)

	serviceNamePtr, _ := windows.UTF16PtrFromString(winservice.ServiceName)
	svcHandle, err := windows.OpenService(scmHandle, serviceNamePtr, windows.SERVICE_STOP|windows.SERVICE_QUERY_STATUS)
	if err != nil {
		return fmt.Errorf("failed to open service: %w", err)
	}
	defer windows.CloseServiceHandle(svcHandle)

	// Check if already stopped
	var status windows.SERVICE_STATUS
	if err := windows.QueryServiceStatus(svcHandle, &status); err == nil {
		if status.CurrentState == windows.SERVICE_STOPPED {
			log.Printf("Service already stopped")
			return nil
		}
	}

	// Stop the service
	err = windows.ControlService(svcHandle, windows.SERVICE_CONTROL_STOP, &status)
	if err != nil {
		return fmt.Errorf("failed to stop service: %w", err)
	}

	log.Printf("Service stop initiated")
	return nil
}

// ensureServiceRunning starts the service if it's not already running.
// Called when the tray application starts.
func (t *Tray) ensureServiceRunning() {
	status, err := getServiceStatus()
	if err != nil {
		if status == winservice.StateNotInstalled {
			log.Printf("Service not installed, cannot auto-start")
			return
		}
		log.Printf("Failed to get service status: %v", err)
		return
	}

	switch status {
	case winservice.StateRunning:
		log.Printf("Service already running")
	case winservice.StateStopped:
		log.Printf("Service stopped, starting it...")
		if err := startService(); err != nil {
			log.Printf("Failed to start service: %v", err)
			showError("PinShare", fmt.Sprintf("Failed to start service:\n\n%v", err))
		} else {
			// Wait a moment and update status
			time.Sleep(1 * time.Second)
			t.updateStatus()
		}
	case winservice.StateStartPending:
		log.Printf("Service is starting...")
	default:
		log.Printf("Service in state: %s", status)
	}
}

// truncateErrorMessage truncates error messages to a maximum length for display
func truncateErrorMessage(msg string) string {
	if len(msg) > winservice.MaxErrorMessageLength {
		return msg[:winservice.MaxErrorMessageLength] + "..."
	}
	return msg
}

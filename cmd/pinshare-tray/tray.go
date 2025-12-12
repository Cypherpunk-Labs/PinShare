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

const (
	// healthCheckTimeout is the HTTP timeout for health check requests
	healthCheckTimeout = 2 * time.Second

	// serviceActionDelay is the delay after starting/stopping service before checking status
	serviceActionDelay = 1 * time.Second

	// serviceRestartDelay is the delay between stop and start during restart
	serviceRestartDelay = 2 * time.Second
)
// ServiceState represents the state of the Windows service
type ServiceState string

const (
	StateRunning      ServiceState = "RUNNING"
	StateStopped      ServiceState = "STOPPED"
	StateStartPending ServiceState = "START_PENDING"
	StateStopPending  ServiceState = "STOP_PENDING"
	StateNotInstalled ServiceState = "NOT_INSTALLED"
)

type Tray struct {
	// Menu items
	menuOpenUI         *systray.MenuItem
	menuStatus         *systray.MenuItem
	menuIPFSStatus     *systray.MenuItem
	menuPinShareStatus *systray.MenuItem
	menuPeersStatus    *systray.MenuItem
	menuSeparator1     *systray.MenuItem
	menuStart          *systray.MenuItem
	menuStop           *systray.MenuItem
	menuRestart        *systray.MenuItem
	menuSeparator2     *systray.MenuItem
	menuSettings       *systray.MenuItem
	menuLogs           *systray.MenuItem
	menuAbout          *systray.MenuItem
	menuSeparator3     *systray.MenuItem
	menuExit           *systray.MenuItem

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
		time.Sleep(serviceActionDelay)
		t.updateStatus()
	}
}

// handleStopService stops the service
func (t *Tray) handleStopService() {
	if err := stopService(); err != nil {
		log.Printf("Failed to stop service: %v", err)
		showError("PinShare", fmt.Sprintf("Failed to stop service:\n\n%v", err))
	} else {
		time.Sleep(serviceActionDelay)
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
	time.Sleep(serviceRestartDelay)

	// Start again
	if err := startService(); err != nil {
		log.Printf("Failed to start service: %v", err)
		showError("PinShare", fmt.Sprintf("Failed to start service:\n\n%v", err))
	} else {
		time.Sleep(serviceActionDelay)
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
			"https://github.com/Cypherpunk-Labs/PinShare")
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
		if status == StateNotInstalled {
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
	case StateRunning:
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

	case StateStopped:
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

	case StateStartPending:
		t.menuStatus.SetTitle("Status: Starting...")
		t.menuIPFSStatus.SetTitle("  IPFS: Starting...")
		t.menuPinShareStatus.SetTitle("  PinShare: Starting...")
		t.menuPeersStatus.SetTitle("  Peers: Connecting...")
		t.menuStart.Disable()
		t.menuStop.Disable()
		t.menuRestart.Disable()
		systray.SetTooltip("PinShare - Starting...")

	case StateStopPending:
		t.menuStatus.SetTitle("Status: Stopping...")
		t.menuStart.Disable()
		t.menuStop.Disable()
		t.menuRestart.Disable()
		systray.SetTooltip("PinShare - Stopping...")

	case StateNotInstalled:
		t.serviceRunning = false
		t.menuStatus.SetTitle("Status: Not Installed")
		t.menuIPFSStatus.SetTitle("  IPFS: -")
		t.menuPinShareStatus.SetTitle("  PinShare: -")
		t.menuPeersStatus.SetTitle("  Peers: -")
		t.menuStart.Enable()
		t.menuStop.Disable()
		t.menuRestart.Disable()
		systray.SetTooltip("PinShare - Service not installed")

	default:
		t.menuStatus.SetTitle(fmt.Sprintf("Status: Unknown (%s)", status))
		systray.SetTooltip("PinShare - Unknown status")
	}
}

// getServiceStatus gets the current service status using Windows Service Manager API (no spawned process)
// Uses minimal permissions (SC_MANAGER_CONNECT and SERVICE_QUERY_STATUS) so no elevation is required.
func getServiceStatus() (ServiceState, error) {
	// Open service control manager with minimal permissions (connect only)
	scmHandle, err := windows.OpenSCManager(nil, nil, windows.SC_MANAGER_CONNECT)
	if err != nil {
		return StateStopped, fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer windows.CloseServiceHandle(scmHandle)

	// Open the service with query status permission only
	serviceNamePtr, err := windows.UTF16PtrFromString(winservice.ServiceName)
	if err != nil {
		return StateStopped, fmt.Errorf("invalid service name: %w", err)
	}

	svcHandle, err := windows.OpenService(scmHandle, serviceNamePtr, windows.SERVICE_QUERY_STATUS)
	if err != nil {
		// Service doesn't exist (ERROR_SERVICE_DOES_NOT_EXIST = 1060)
		return StateNotInstalled, fmt.Errorf("service not installed")
	}
	defer windows.CloseServiceHandle(svcHandle)

	// Query the service status
	var status windows.SERVICE_STATUS
	err = windows.QueryServiceStatus(svcHandle, &status)
	if err != nil {
		return StateStopped, fmt.Errorf("failed to query service: %w", err)
	}

	// Map Windows service state to our ServiceState
	switch status.CurrentState {
	case windows.SERVICE_RUNNING:
		return StateRunning, nil
	case windows.SERVICE_STOPPED:
		return StateStopped, nil
	case windows.SERVICE_START_PENDING:
		return StateStartPending, nil
	case windows.SERVICE_STOP_PENDING:
		return StateStopPending, nil
	case windows.SERVICE_PAUSED, windows.SERVICE_PAUSE_PENDING, windows.SERVICE_CONTINUE_PENDING:
		return StateStopped, nil
	default:
		return StateStopped, fmt.Errorf("unknown service state: %d", status.CurrentState)
	}
}

// checkIPFSHealth checks if IPFS daemon is responding
func checkIPFSHealth() bool {
	config := getConfig()
	client := &http.Client{
		Timeout: healthCheckTimeout,
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
		Timeout: healthCheckTimeout,
	}

	url := fmt.Sprintf("http://localhost:%d/api/health", config.PinShareAPIPort)
	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// runElevated runs a command with UAC elevation using ShellExecute
func runElevated(executable, args string) error {
	verbPtr, _ := windows.UTF16PtrFromString("runas")
	exePtr, _ := windows.UTF16PtrFromString(executable)
	argsPtr, _ := windows.UTF16PtrFromString(args)

	// ShellExecute with "runas" verb triggers UAC prompt
	err := windows.ShellExecute(0, verbPtr, exePtr, argsPtr, nil, windows.SW_HIDE)
	if err != nil {
		return err
	}
	return nil
}

// startService starts the service using sc.exe with UAC elevation
func startService() error {
	log.Printf("Starting service %s with elevation...", winservice.ServiceName)

	err := runElevated("sc.exe", fmt.Sprintf("start %s", winservice.ServiceName))
	if err != nil {
		log.Printf("Failed to start service: %v", err)
		return fmt.Errorf("failed to start service: %w", err)
	}

	log.Printf("Service start command initiated")
	return nil
}

// stopService stops the service using sc.exe with UAC elevation
func stopService() error {
	log.Printf("Stopping service %s with elevation...", winservice.ServiceName)

	err := runElevated("sc.exe", fmt.Sprintf("stop %s", winservice.ServiceName))
	if err != nil {
		log.Printf("Failed to stop service: %v", err)
		return fmt.Errorf("failed to stop service: %w", err)
	}

	log.Printf("Service stop command initiated")
	return nil
}

// truncateErrorMessage truncates error messages to a maximum length for display
func truncateErrorMessage(msg string) string {
	if len(msg) > winservice.MaxErrorMessageLength {
		return msg[:winservice.MaxErrorMessageLength] + "..."
	}
	return msg
}

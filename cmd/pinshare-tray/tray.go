package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"
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

	// Initial status check
	t.updateStatus()
}

// StartMenuHandler starts the menu click handler goroutine.
// Call this after BuildMenu with the context from your main function.
func (t *Tray) StartMenuHandler(ctx context.Context) {
	go t.handleMenuClicks(ctx)
}

// handleMenuClicks handles menu item clicks.
// It accepts a context for graceful shutdown.
func (t *Tray) handleMenuClicks(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Menu handler shutting down...")
			return

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
			if t.handleExit() {
				return
			}
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
		showError(appName, fmt.Sprintf("Failed to start service:\n\n%v", err))
	} else {
		time.Sleep(serviceActionDelay)
		t.updateStatus()
	}
}

// handleStopService stops the service
func (t *Tray) handleStopService() {
	if err := stopService(); err != nil {
		log.Printf("Failed to stop service: %v", err)
		showError(appName, fmt.Sprintf("Failed to stop service:\n\n%v", err))
	} else {
		time.Sleep(serviceActionDelay)
		t.updateStatus()
	}
}

// handleRestartService restarts the service by stopping and starting it.
func (t *Tray) handleRestartService() {
	// Stop first using existing handler
	t.handleStopService()

	// Wait before starting
	time.Sleep(winservice.ServiceRestartDelay)

	// Start again using existing handler
	t.handleStartService()
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
	dataDir := getUserDataDirectory()
	logDir := filepath.Join(dataDir, "logs")

	if err := openBrowser(logDir); err != nil {
		log.Printf("Failed to open log directory: %v", err)
		showMessage("Error", "Failed to open log directory")
	}
}

// handleAbout shows about information
func (t *Tray) handleAbout() {
	showMessage("About PinShare",
		"PinShare - Decentralized IPFS Pinning Service\n"+
			"Version "+winservice.Version+"\n\n"+
			"https://github.com/Cypherpunk-Labs/PinShare")
}

// handleExit handles the Exit menu item with a confirmation dialog.
// Returns true if the application should exit, false to stay in tray.
func (t *Tray) handleExit() bool {
	result := showYesNoCancelDialog(
		"Exit PinShare",
		"Do you want to stop the PinShare service before exiting?\n\n"+
			"Yes - Stop service and exit\n"+
			"No - Exit (service continues running)\n"+
			"Cancel - Stay in tray")

	switch result {
	case IDYES:
		log.Println("User chose to stop service and exit")
		if err := stopService(); err != nil {
			log.Printf("Failed to stop service: %v", err)
			showError(appName, fmt.Sprintf("Failed to stop service:\n\n%v", err))
		}
		systray.Quit()
		return true
	case IDNO:
		log.Println("User chose to exit without stopping service")
		systray.Quit()
		return true
	default:
		log.Println("User cancelled exit")
		return false
	}
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

// startService starts the service, trying direct API first, then falling back to UAC elevation
func startService() error {
	log.Printf("Starting service %s...", winservice.ServiceName)

	// Try direct API first (works if DACL was set during installation)
	err := startServiceDirect()
	if err == nil {
		return nil
	}

	// Check if it's an access denied error
	if isAccessDenied(err) {
		log.Printf("Direct API access denied, trying with elevation...")
		return startServiceElevated()
	}

	return err
}

// startServiceDirect tries to start the service using Windows API directly
func startServiceDirect() error {
	scmHandle, err := windows.OpenSCManager(nil, nil, windows.SC_MANAGER_CONNECT)
	if err != nil {
		return fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer windows.CloseServiceHandle(scmHandle)

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

	err = windows.StartService(svcHandle, 0, nil)
	if err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}

	log.Printf("Service start initiated")
	return nil
}

// startServiceElevated starts the service using sc.exe with UAC elevation
func startServiceElevated() error {
	log.Printf("Starting service %s with elevation...", winservice.ServiceName)

	err := runElevated("sc.exe", fmt.Sprintf("start %s", winservice.ServiceName))
	if err != nil {
		return err
	}

	// ShellExecute is async, so we need to poll until the service is actually running
	return waitForServiceState(windows.SERVICE_RUNNING, winservice.ServiceStartTimeout)
}

// stopService stops the service, trying direct API first, then falling back to UAC elevation
func stopService() error {
	log.Printf("Stopping service %s...", winservice.ServiceName)

	// Try direct API first (works if DACL was set during installation)
	err := stopServiceDirect()
	if err == nil {
		return nil
	}

	// Check if it's an access denied error
	if isAccessDenied(err) {
		log.Printf("Direct API access denied, trying with elevation...")
		return stopServiceElevated()
	}

	return err
}

// stopServiceDirect tries to stop the service using Windows API directly
func stopServiceDirect() error {
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

	err = windows.ControlService(svcHandle, windows.SERVICE_CONTROL_STOP, &status)
	if err != nil {
		return fmt.Errorf("failed to stop service: %w", err)
	}

	log.Printf("Service stop initiated, waiting for stop to complete...")

	// Wait for the service to fully stop
	return waitForServiceState(windows.SERVICE_STOPPED, winservice.ServiceStopTimeout)
}

// stopServiceElevated stops the service using sc.exe with UAC elevation
func stopServiceElevated() error {
	log.Printf("Stopping service %s with elevation...", winservice.ServiceName)

	err := runElevated("sc.exe", fmt.Sprintf("stop %s", winservice.ServiceName))
	if err != nil {
		return err
	}

	// ShellExecute is async, so we need to poll until the service is actually stopped
	return waitForServiceState(windows.SERVICE_STOPPED, winservice.ServiceStopTimeout)
}

// runElevated runs a command with UAC elevation using ShellExecute
func runElevated(executable, args string) error {
	verbPtr, _ := windows.UTF16PtrFromString("runas")
	exePtr, _ := windows.UTF16PtrFromString(executable)
	argsPtr, _ := windows.UTF16PtrFromString(args)

	err := windows.ShellExecute(0, verbPtr, exePtr, argsPtr, nil, windows.SW_HIDE)
	if err != nil {
		return fmt.Errorf("failed to execute with elevation: %w", err)
	}
	return nil
}

// waitForServiceState polls until the service reaches the desired state or times out
func waitForServiceState(desiredState uint32, timeout time.Duration) error {
	scmHandle, err := windows.OpenSCManager(nil, nil, windows.SC_MANAGER_CONNECT)
	if err != nil {
		return fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer windows.CloseServiceHandle(scmHandle)

	serviceNamePtr, _ := windows.UTF16PtrFromString(winservice.ServiceName)
	svcHandle, err := windows.OpenService(scmHandle, serviceNamePtr, windows.SERVICE_QUERY_STATUS)
	if err != nil {
		return fmt.Errorf("failed to open service: %w", err)
	}
	defer windows.CloseServiceHandle(svcHandle)

	deadline := time.Now().Add(timeout)
	pollInterval := winservice.ServicePollInterval

	for time.Now().Before(deadline) {
		var status windows.SERVICE_STATUS
		if err := windows.QueryServiceStatus(svcHandle, &status); err == nil {
			if status.CurrentState == desiredState {
				log.Printf("Service reached desired state")
				return nil
			}
		}
		time.Sleep(pollInterval)
	}

	return fmt.Errorf("timeout waiting for service to reach desired state")
}

// isAccessDenied checks if an error is an "access denied" error
func isAccessDenied(err error) bool {
	if err == nil {
		return false
	}
	// Check for Windows ERROR_ACCESS_DENIED (5)
	if err == windows.ERROR_ACCESS_DENIED {
		return true
	}
	// Also check error message for wrapped errors
	errStr := err.Error()
	return strings.Contains(errStr, "Access is denied") ||
		strings.Contains(errStr, "access denied") ||
		strings.Contains(errStr, "ERROR_ACCESS_DENIED")
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
			showError(appName, fmt.Sprintf("Failed to start service:\n\n%v", err))
		} else {
			// Wait a moment and update status
			time.Sleep(winservice.HealthCheckPoll)
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

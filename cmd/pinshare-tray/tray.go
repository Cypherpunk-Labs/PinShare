package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/getlantern/systray"
)

// contains checks if s contains substr (case-insensitive)
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

const (
	serviceName = "PinShareService"
	uiPort      = 8888 // Default UI port
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
		showError("PinShare", fmt.Sprintf("Failed to start service:\n\n%v", err))
	} else {
		showMessage("PinShare", "Service started successfully.")
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
		showMessage("PinShare", "Service stopped successfully.")
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

		// Check if it's a "service not installed" error
		errStr := err.Error()
		if contains(errStr, "not installed") {
			t.menuStatus.SetTitle("Status: Not Installed")
			systray.SetTooltip("PinShare - Service not installed")
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

// getServiceStatus gets the current service status using sc query (no admin required)
func getServiceStatus() (ServiceState, error) {
	cmd := exec.Command("sc", "query", serviceName)
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	if err != nil {
		// Check for error 1060: service doesn't exist
		if strings.Contains(outputStr, "1060") ||
			strings.Contains(outputStr, "does not exist") ||
			strings.Contains(outputStr, "FAILED 1060") {
			return StateNotInstalled, fmt.Errorf("service not installed")
		}
		return StateStopped, fmt.Errorf("sc query failed: %w", err)
	}

	// Parse state from sc query output
	if strings.Contains(outputStr, "RUNNING") {
		return StateRunning, nil
	} else if strings.Contains(outputStr, "STOPPED") {
		return StateStopped, nil
	} else if strings.Contains(outputStr, "START_PENDING") {
		return StateStartPending, nil
	} else if strings.Contains(outputStr, "STOP_PENDING") {
		return StateStopPending, nil
	}

	return StateStopped, fmt.Errorf("unknown service state")
}

// checkIPFSHealth checks if IPFS daemon is responding
func checkIPFSHealth() bool {
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	// IPFS version endpoint requires POST
	resp, err := client.Post("http://localhost:5001/api/v0/version",
		"application/json", nil)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// checkPinShareHealth checks if PinShare API is responding
func checkPinShareHealth() bool {
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	resp, err := client.Get("http://localhost:9090/api/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// startService starts the service using PowerShell with UAC elevation
func startService() error {
	log.Printf("Starting service %s with elevation...", serviceName)

	// Use PowerShell Start-Process with -Verb RunAs for UAC elevation
	// -WindowStyle Hidden prevents console window from appearing
	psCmd := fmt.Sprintf(
		"Start-Process -FilePath 'sc' -ArgumentList 'start %s' "+
			"-Verb RunAs -Wait -WindowStyle Hidden",
		serviceName)

	cmd := exec.Command("powershell", "-Command", psCmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Failed to start service: %v, output: %s", err, string(output))
		return fmt.Errorf("failed to start service: %w", err)
	}

	log.Printf("Service start command completed")
	return nil
}

// stopService stops the service using PowerShell with UAC elevation
func stopService() error {
	log.Printf("Stopping service %s with elevation...", serviceName)

	psCmd := fmt.Sprintf(
		"Start-Process -FilePath 'sc' -ArgumentList 'stop %s' "+
			"-Verb RunAs -Wait -WindowStyle Hidden",
		serviceName)

	cmd := exec.Command("powershell", "-Command", psCmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Failed to stop service: %v, output: %s", err, string(output))
		return fmt.Errorf("failed to stop service: %w", err)
	}

	log.Printf("Service stop command completed")
	return nil
}

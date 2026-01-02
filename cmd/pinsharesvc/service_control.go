package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"pinshare/internal/winservice"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

// getCurrentUserSID returns the SID string for the current user.
// This is used to grant service control permissions to the installing user.
func getCurrentUserSID() (string, error) {
	token := windows.GetCurrentProcessToken()
	tokenUser, err := token.GetTokenUser()
	if err != nil {
		return "", fmt.Errorf("failed to get token user: %w", err)
	}
	return tokenUser.User.Sid.String(), nil
}

// setServiceDACL modifies the service security descriptor to allow the specified
// user SID to start/stop the service without administrator privileges.
// This enables the tray application to control the service without UAC prompts.
func setServiceDACL(serviceName, userSID string) error {
	// Get current security descriptor using sc.exe sdshow
	getCmd := exec.Command("sc.exe", "sdshow", serviceName)
	output, err := getCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get service security descriptor: %w", err)
	}

	currentSD := strings.TrimSpace(string(output))

	// Build ACE (Access Control Entry) for user:
	// A = Allow
	// RPWPDTLO = SERVICE_START | SERVICE_STOP | SERVICE_PAUSE_CONTINUE |
	//            SERVICE_INTERROGATE | SERVICE_QUERY_STATUS | SERVICE_QUERY_CONFIG
	// Format: (A;;RPWPDTLO;;;user-sid)
	userACE := fmt.Sprintf("(A;;RPWPDTLO;;;%s)", userSID)

	// Insert user ACE into DACL after "D:"
	// Existing format typically: D:(A;;...)(A;;...)S:(AU;...)
	if !strings.Contains(currentSD, "D:") {
		return fmt.Errorf("unexpected security descriptor format: missing DACL")
	}

	newSD := strings.Replace(currentSD, "D:", "D:"+userACE, 1)

	// Set the new security descriptor using sc.exe sdset
	setCmd := exec.Command("sc.exe", "sdset", serviceName, newSD)
	if output, err := setCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to set service security descriptor: %w\nOutput: %s", err, output)
	}

	return nil
}

// installService installs PinShare as a Windows service
// If autoStart is true, the service will start automatically on boot.
// If false (default), the service starts manually (tray app controls it).
func installService(autoStart bool) error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Connect to service manager
	manager, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer manager.Disconnect()

	// Check if service already exists
	service, err := manager.OpenService(winservice.ServiceName)
	if err == nil {
		// Service already exists - update the start type if needed
		fmt.Printf("Service %s already exists, updating configuration...\n", winservice.ServiceName)

		// Determine desired start type
		var desiredStartType uint32 = mgr.StartManual
		if autoStart {
			desiredStartType = mgr.StartAutomatic
		}

		// Get current config to update
		currentConfig, err := service.Config()
		if err != nil {
			service.Close()
			return fmt.Errorf("failed to get service config: %w", err)
		}

		// Update start type if different
		if currentConfig.StartType != desiredStartType {
			currentConfig.StartType = desiredStartType
			if err := service.UpdateConfig(currentConfig); err != nil {
				service.Close()
				return fmt.Errorf("failed to update service config: %w", err)
			}
			startTypeName := "Manual"
			if autoStart {
				startTypeName = "Automatic"
			}
			fmt.Printf("Service start type updated to: %s\n", startTypeName)
		}

		service.Close()
		return nil
	}

	// Determine start type based on autoStart flag
	var startType uint32 = mgr.StartManual
	if autoStart {
		startType = mgr.StartAutomatic
	}

	// Create Windows service configuration
	winSvcConfig := mgr.Config{
		DisplayName:  winservice.ServiceDisplayName,
		Description:  winservice.ServiceDescription,
		StartType:    startType,
		ErrorControl: mgr.ErrorNormal,
	}

	// Create service
	service, err = manager.CreateService(winservice.ServiceName, exePath, winSvcConfig)
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}
	defer service.Close()

	// Set recovery options
	recoveryActions := []mgr.RecoveryAction{
		{
			Type:  mgr.ServiceRestart,
			Delay: winservice.RecoveryDelayFirst,
		},
		{
			Type:  mgr.ServiceRestart,
			Delay: winservice.RecoveryDelaySecond,
		},
		{
			Type:  mgr.ServiceRestart,
			Delay: winservice.RecoveryDelayThird,
		},
	}

	if err := service.SetRecoveryActions(recoveryActions, winservice.RecoveryResetPeriod); err != nil {
		// Non-fatal, just log
		fmt.Printf("Warning: Failed to set recovery actions: %v\n", err)
	}

	// Get current user SID and set DACL to allow user control without UAC
	userSID, err := getCurrentUserSID()
	if err != nil {
		fmt.Printf("Warning: Failed to get user SID: %v\n", err)
		fmt.Println("Service control will require administrator privileges")
	} else {
		if err := setServiceDACL(winservice.ServiceName, userSID); err != nil {
			fmt.Printf("Warning: Failed to set service DACL: %v\n", err)
			fmt.Println("Service control will require administrator privileges")
		} else {
			fmt.Println("Service permissions configured for user control (no UAC required)")
		}
	}

	// Install event log source
	if err := installEventLogSource(); err != nil {
		fmt.Printf("Warning: Failed to install event log source: %v\n", err)
	}

	// Initialize PinShare application configuration
	pinShareConfig, err := getDefaultConfig()
	if err != nil {
		return fmt.Errorf("failed to get default config: %w", err)
	}

	// Get install directory from executable path
	pinShareConfig.InstallDirectory = filepath.Dir(exePath)

	// Ensure directories exist
	if err := pinShareConfig.EnsureDirectories(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Save configuration to JSON file
	if err := pinShareConfig.SaveToFile(); err != nil {
		return fmt.Errorf("failed to save config file: %w", err)
	}

	fmt.Printf("Service %s installed successfully\n", winservice.ServiceName)
	fmt.Printf("Installation directory: %s\n", pinShareConfig.InstallDirectory)
	fmt.Printf("Data directory: %s\n", pinShareConfig.DataDirectory)
	fmt.Printf("\nTo start the service, run: %s start\n", exePath)

	return nil
}

// uninstallService uninstalls the PinShare Windows service
func uninstallService() error {
	// Connect to service manager
	manager, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer manager.Disconnect()

	// Open service
	service, err := manager.OpenService(winservice.ServiceName)
	if err != nil {
		return fmt.Errorf("service %s not found: %w", winservice.ServiceName, err)
	}
	defer service.Close()

	// Stop service if running
	status, err := service.Query()
	if err != nil {
		return fmt.Errorf("failed to query service status: %w", err)
	}

	if status.State != svc.Stopped {
		fmt.Println("Stopping service...")
		status, err = service.Control(svc.Stop)
		if err != nil {
			return fmt.Errorf("failed to stop service: %w", err)
		}

		// Wait for service to stop
		timeout := time.Now().Add(winservice.ServiceStopTimeout)
		for status.State != svc.Stopped {
			if time.Now().After(timeout) {
				return fmt.Errorf("timeout waiting for service to stop")
			}
			time.Sleep(winservice.ServicePollInterval)
			status, err = service.Query()
			if err != nil {
				return fmt.Errorf("failed to query service status: %w", err)
			}
		}
		fmt.Println("Service stopped")
	}

	// Delete service
	if err := service.Delete(); err != nil {
		return fmt.Errorf("failed to delete service: %w", err)
	}

	// Remove event log source
	if err := removeEventLogSource(); err != nil {
		fmt.Printf("Warning: Failed to remove event log source: %v\n", err)
	}

	fmt.Printf("Service %s uninstalled successfully\n", winservice.ServiceName)
	fmt.Println("\nNote: Data directory was not removed. To remove it manually, delete:")

	config, err := LoadConfig()
	if err == nil {
		fmt.Printf("  %s\n", config.DataDirectory)
	}

	return nil
}

// startService starts the PinShare service
func startService() error {
	// Connect to service manager
	manager, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer manager.Disconnect()

	// Open service
	service, err := manager.OpenService(winservice.ServiceName)
	if err != nil {
		return fmt.Errorf("service %s not found: %w", winservice.ServiceName, err)
	}
	defer service.Close()

	// Check current state first
	status, err := service.Query()
	if err != nil {
		return fmt.Errorf("failed to query service status: %w", err)
	}

	// If already running, nothing to do
	if status.State == svc.Running {
		fmt.Printf("Service %s is already running\n", winservice.ServiceName)
		return nil
	}

	// Start service
	if err := service.Start(); err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}

	// Wait for service to be running
	fmt.Printf("Starting service %s...\n", winservice.ServiceName)
	timeout := time.Now().Add(winservice.ServiceStartTimeout)
	for {
		status, err = service.Query()
		if err != nil {
			return fmt.Errorf("failed to query service status: %w", err)
		}

		if status.State == svc.Running {
			break
		}

		if status.State == svc.Stopped {
			return fmt.Errorf("service failed to start (stopped)")
		}

		if time.Now().After(timeout) {
			return fmt.Errorf("timeout waiting for service to start")
		}

		time.Sleep(winservice.ServicePollInterval)
	}

	fmt.Printf("Service %s started successfully\n", winservice.ServiceName)

	// Load config to show API URL
	config, err := LoadConfig()
	if err == nil {
		fmt.Printf("\nPinShare API available at: http://localhost:%d\n", config.PinShareAPIPort)
	}

	return nil
}

// stopService stops the PinShare service
func stopService() error {
	// Connect to service manager
	manager, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer manager.Disconnect()

	// Open service
	service, err := manager.OpenService(winservice.ServiceName)
	if err != nil {
		return fmt.Errorf("service %s not found: %w", winservice.ServiceName, err)
	}
	defer service.Close()

	// Stop service
	status, err := service.Control(svc.Stop)
	if err != nil {
		return fmt.Errorf("failed to stop service: %w", err)
	}

	// Wait for service to stop
	timeout := time.Now().Add(winservice.ServiceStopTimeout)
	for status.State != svc.Stopped {
		if time.Now().After(timeout) {
			return fmt.Errorf("timeout waiting for service to stop")
		}
		time.Sleep(winservice.ServicePollInterval)
		status, err = service.Query()
		if err != nil {
			return fmt.Errorf("failed to query service status: %w", err)
		}
	}

	fmt.Printf("Service %s stopped successfully\n", winservice.ServiceName)
	return nil
}

// restartService restarts the PinShare service
func restartService() error {
	fmt.Println("Stopping service...")
	if err := stopService(); err != nil {
		return err
	}

	time.Sleep(winservice.ServiceRestartDelay)

	fmt.Println("Starting service...")
	return startService()
}

// installEventLogSource installs the event log source
func installEventLogSource() error {
	// Custom event log source registration requires registry modification under
	// HKLM\SYSTEM\CurrentControlSet\Services\EventLog\Application\<ServiceName>
	// which needs admin privileges. The Windows event log will work without this
	// custom source registration - events will be logged under the generic
	// "Application" source. Skipping for now to avoid registry dependencies.
	fmt.Println("Note: Custom event log source registration skipped (using generic Application source)")
	return nil
}

// removeEventLogSource removes the event log source
func removeEventLogSource() error {
	// Corresponding cleanup for installEventLogSource - since we don't register
	// a custom source, there's nothing to remove.
	return nil
}

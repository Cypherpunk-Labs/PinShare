package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Cypherpunk-Labs/PinShare/internal/winservice"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

// installService installs PinShare as a Windows service
func installService() error {
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
		service.Close()
		// Service already exists - this is fine for reinstall/upgrade scenarios where
		// the MSI installer runs the install command but the service is already registered.
		// We skip re-registration to preserve the existing service configuration and avoid
		// errors from attempting to create a duplicate service entry.
		fmt.Printf("Service %s already exists, skipping installation\n", winservice.ServiceName)
		return nil
	}

	// Create Windows service configuration
	winSvcConfig := mgr.Config{
		DisplayName:  winservice.ServiceDisplayName,
		Description:  winservice.ServiceDescription,
		StartType:    mgr.StartAutomatic,
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
			Delay: 5 * time.Second,
		},
		{
			Type:  mgr.ServiceRestart,
			Delay: 10 * time.Second,
		},
		{
			Type:  mgr.ServiceRestart,
			Delay: 30 * time.Second,
		},
	}

	if err := service.SetRecoveryActions(recoveryActions, 60); err != nil {
		// Non-fatal, just log
		fmt.Printf("Warning: Failed to set recovery actions: %v\n", err)
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

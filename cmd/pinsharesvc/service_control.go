package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

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
	service, err := manager.OpenService(serviceName)
	if err == nil {
		service.Close()
		return fmt.Errorf("service %s already exists", serviceName)
	}

	// Create service configuration
	config := mgr.Config{
		DisplayName:  "PinShare Service",
		Description:  "PinShare - Decentralized IPFS pinning service with libp2p",
		StartType:    mgr.StartAutomatic,
		ErrorControl: mgr.ErrorNormal,
	}

	// Create service
	service, err = manager.CreateService(serviceName, exePath, config)
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

	// Initialize configuration
	config_, err := getDefaultConfig()
	if err != nil {
		return fmt.Errorf("failed to get default config: %w", err)
	}

	// Get install directory from executable path
	config_.InstallDirectory = filepath.Dir(exePath)

	// Ensure directories exist
	if err := config_.EnsureDirectories(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Save configuration
	if err := config_.SaveToRegistry(); err != nil {
		fmt.Printf("Warning: Failed to save to registry: %v\n", err)
	}

	if err := config_.SaveToFile(); err != nil {
		fmt.Printf("Warning: Failed to save to file: %v\n", err)
	}

	fmt.Printf("Service %s installed successfully\n", serviceName)
	fmt.Printf("Installation directory: %s\n", config_.InstallDirectory)
	fmt.Printf("Data directory: %s\n", config_.DataDirectory)
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
	service, err := manager.OpenService(serviceName)
	if err != nil {
		return fmt.Errorf("service %s not found: %w", serviceName, err)
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
		timeout := time.Now().Add(30 * time.Second)
		for status.State != svc.Stopped {
			if time.Now().After(timeout) {
				return fmt.Errorf("timeout waiting for service to stop")
			}
			time.Sleep(300 * time.Millisecond)
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

	fmt.Printf("Service %s uninstalled successfully\n", serviceName)
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
	service, err := manager.OpenService(serviceName)
	if err != nil {
		return fmt.Errorf("service %s not found: %w", serviceName, err)
	}
	defer service.Close()

	// Start service
	if err := service.Start(); err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}

	fmt.Printf("Service %s started successfully\n", serviceName)

	// Load config to show UI URL
	config, err := LoadConfig()
	if err == nil {
		fmt.Printf("\nPinShare UI available at: http://localhost:%d\n", config.UIPort)
		fmt.Printf("PinShare API available at: http://localhost:%d\n", config.PinShareAPIPort)
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
	service, err := manager.OpenService(serviceName)
	if err != nil {
		return fmt.Errorf("service %s not found: %w", serviceName, err)
	}
	defer service.Close()

	// Stop service
	status, err := service.Control(svc.Stop)
	if err != nil {
		return fmt.Errorf("failed to stop service: %w", err)
	}

	// Wait for service to stop
	timeout := time.Now().Add(30 * time.Second)
	for status.State != svc.Stopped {
		if time.Now().After(timeout) {
			return fmt.Errorf("timeout waiting for service to stop")
		}
		time.Sleep(300 * time.Millisecond)
		status, err = service.Query()
		if err != nil {
			return fmt.Errorf("failed to query service status: %w", err)
		}
	}

	fmt.Printf("Service %s stopped successfully\n", serviceName)
	return nil
}

// restartService restarts the PinShare service
func restartService() error {
	fmt.Println("Stopping service...")
	if err := stopService(); err != nil {
		return err
	}

	time.Sleep(2 * time.Second)

	fmt.Println("Starting service...")
	return startService()
}

// installEventLogSource installs the event log source
func installEventLogSource() error {
	// This requires registry modification which needs admin privileges
	// The event log will work without this, just won't have a custom source
	// For now, we'll skip this and use the generic event log
	return nil
}

// removeEventLogSource removes the event log source
func removeEventLogSource() error {
	// Corresponding cleanup for installEventLogSource
	return nil
}

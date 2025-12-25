package main

import (
	"fmt"
	"log"
	"os"

	"pinshare/internal/winservice"

	"golang.org/x/sys/windows/svc"
)

func main() {
	// Check if running as Windows service
	isWindowsService, err := svc.IsWindowsService()
	if err != nil {
		log.Fatalf("Failed to determine if running as service: %v", err)
	}

	if isWindowsService {
		// Run as Windows service
		runService()
		return
	}

	// Command-line interface for service management
	if len(os.Args) < 2 {
		usage()
		return
	}

	cmd := os.Args[1]
	switch cmd {
	case "install":
		// Check for --auto-start flag
		autoStart := false
		for _, arg := range os.Args[2:] {
			if arg == "--auto-start" {
				autoStart = true
				break
			}
		}
		err = installService(autoStart)
	case "uninstall":
		err = uninstallService()
	case "start":
		err = startService()
	case "stop":
		err = stopService()
	case "restart":
		err = restartService()
	case "debug":
		// Run in console mode for debugging
		err = runDebugMode()
	default:
		usage()
		return
	}

	if err != nil {
		log.Fatalf("Error executing %s: %v", cmd, err)
	}
	fmt.Printf("Successfully executed %s\n", cmd)
}

func usage() {
	fmt.Fprintf(os.Stderr, `Usage: %s <command> [options]

Commands:
  install [--auto-start]  Install PinShare as a Windows service
                          --auto-start: Start automatically on boot
  uninstall               Uninstall PinShare Windows service
  start                   Start PinShare service
  stop                    Stop PinShare service
  restart                 Restart PinShare service
  debug                   Run in console mode (for debugging)

`, os.Args[0])
}

func runService() {
	err := svc.Run(winservice.ServiceName, &pinshareService{})
	if err != nil {
		log.Fatalf("Service failed: %v", err)
	}
}

func runDebugMode() error {
	fmt.Println("Running PinShare in debug mode...")
	fmt.Println("Press Ctrl+C to stop")

	service := new(pinshareService)
	return service.runInteractive()
}

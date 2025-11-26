package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
)

type pinshareService struct {
	config          *ServiceConfig
	processManager  *ProcessManager
	uiServer        *UIServer
	healthChecker   *HealthChecker
	eventLog        debug.Log
	ctx             context.Context
	cancel          context.CancelFunc
	wg              sync.WaitGroup
}

// Execute implements the svc.Handler interface
func (s *pinshareService) Execute(args []string, changeReq <-chan svc.ChangeRequest, statusChan chan<- svc.Status) (bool, uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown | svc.AcceptPauseAndContinue

	// Signal service is starting
	statusChan <- svc.Status{State: svc.StartPending}

	// Initialize service
	if err := s.initialize(); err != nil {
		s.logError("Failed to initialize service", err)
		return true, 1
	}

	// Signal service is running
	statusChan <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
	s.logInfo("PinShare service started successfully")

loop:
	for {
		select {
		case c := <-changeReq:
			switch c.Cmd {
			case svc.Interrogate:
				statusChan <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				s.logInfo("Service stop requested")
				statusChan <- svc.Status{State: svc.StopPending}
				s.shutdown()
				break loop
			case svc.Pause:
				statusChan <- svc.Status{State: svc.Paused, Accepts: cmdsAccepted}
				s.logInfo("Service paused")
			case svc.Continue:
				statusChan <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
				s.logInfo("Service continued")
			default:
				s.logError(fmt.Sprintf("Unexpected control request #%d", c), nil)
			}
		case <-s.ctx.Done():
			s.logInfo("Service context cancelled")
			break loop
		}
	}

	// Wait for shutdown to complete
	s.wg.Wait()
	statusChan <- svc.Status{State: svc.Stopped}
	return false, 0
}

// initialize sets up all service components
func (s *pinshareService) initialize() error {
	s.ctx, s.cancel = context.WithCancel(context.Background())

	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	s.config = config

	// Initialize event log
	s.eventLog, err = openEventLog(serviceName)
	if err != nil {
		return fmt.Errorf("failed to open event log: %w", err)
	}

	s.logInfo(fmt.Sprintf("Loaded configuration: DataDir=%s, IPFSPort=%d, APIPort=%d",
		config.DataDirectory, config.IPFSAPIPort, config.PinShareAPIPort))

	// Ensure directories exist
	if err := s.config.EnsureDirectories(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Initialize process manager
	s.processManager = NewProcessManager(s.config, s.eventLog)

	// Initialize health checker before starting processes (needed for health checks during startup)
	s.logInfo("Initializing health checker...")
	s.healthChecker = NewHealthChecker(s.config, s.processManager, s.eventLog)

	// Start IPFS daemon
	s.logInfo("Starting IPFS daemon...")
	if err := s.processManager.StartIPFS(s.ctx); err != nil {
		return fmt.Errorf("failed to start IPFS: %w", err)
	}

	// Wait for IPFS to be ready
	s.logInfo("Waiting for IPFS to become ready...")
	if err := s.waitForIPFS(); err != nil {
		return fmt.Errorf("IPFS failed to start: %w", err)
	}
	s.logInfo("IPFS daemon is ready")

	// Start PinShare backend
	s.logInfo("Starting PinShare backend...")
	if err := s.processManager.StartPinShare(s.ctx); err != nil {
		return fmt.Errorf("failed to start PinShare: %w", err)
	}

	// Wait for PinShare to be ready
	s.logInfo("Waiting for PinShare API to become ready...")
	if err := s.waitForPinShare(); err != nil {
		return fmt.Errorf("PinShare failed to start: %w", err)
	}
	s.logInfo("PinShare backend is ready")

	// Start embedded UI server
	s.logInfo("Starting embedded UI server...")
	s.uiServer = NewUIServer(s.config, s.eventLog)
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err := s.uiServer.Start(s.ctx); err != nil {
			s.logError("UI server error", err)
		}
	}()
	s.logInfo(fmt.Sprintf("UI server started on http://localhost:%d", s.config.UIPort))

	// Start health checker background monitoring
	s.logInfo("Starting health checker background monitoring...")
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.healthChecker.Run(s.ctx)
	}()
	s.logInfo("Health checker monitoring started")

	return nil
}

// waitForIPFS waits for IPFS daemon to be ready
func (s *pinshareService) waitForIPFS() error {
	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout waiting for IPFS to start")
		case <-ticker.C:
			if s.healthChecker.CheckIPFSHealth() {
				return nil
			}
		}
	}
}

// waitForPinShare waits for PinShare API to be ready
func (s *pinshareService) waitForPinShare() error {
	// PinShare needs time to initialize libp2p, DHT, and connect to peers
	timeout := time.After(60 * time.Second)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout waiting for PinShare to start")
		case <-ticker.C:
			if s.healthChecker.CheckPinShareHealth() {
				return nil
			}
		}
	}
}

// shutdown gracefully stops all service components
func (s *pinshareService) shutdown() {
	s.logInfo("Shutting down PinShare service...")

	// Cancel context to signal all goroutines
	s.cancel()

	// Stop UI server
	if s.uiServer != nil {
		s.logInfo("Stopping UI server...")
		s.uiServer.Stop()
	}

	// Stop PinShare backend
	s.logInfo("Stopping PinShare backend...")
	if err := s.processManager.StopPinShare(); err != nil {
		s.logError("Error stopping PinShare", err)
	}

	// Stop IPFS daemon
	s.logInfo("Stopping IPFS daemon...")
	if err := s.processManager.StopIPFS(); err != nil {
		s.logError("Error stopping IPFS", err)
	}

	// Close event log
	if s.eventLog != nil {
		s.eventLog.Close()
	}

	s.logInfo("PinShare service stopped")
}

// runInteractive runs the service in interactive/debug mode
func (s *pinshareService) runInteractive() error {
	// Initialize service
	if err := s.initialize(); err != nil {
		return fmt.Errorf("failed to initialize: %w", err)
	}

	fmt.Println("Service started successfully!")
	fmt.Printf("UI available at: http://localhost:%d\n", s.config.UIPort)
	fmt.Printf("API available at: http://localhost:%d\n", s.config.PinShareAPIPort)
	fmt.Println("\nPress Ctrl+C to stop...")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nShutting down...")
	s.shutdown()
	s.wg.Wait()
	fmt.Println("Stopped successfully")

	return nil
}

// Logging helpers
func (s *pinshareService) logInfo(msg string) {
	if s.eventLog != nil {
		s.eventLog.Info(1, msg)
	} else {
		log.Println("INFO:", msg)
	}
}

func (s *pinshareService) logError(msg string, err error) {
	errMsg := msg
	if err != nil {
		errMsg = fmt.Sprintf("%s: %v", msg, err)
	}
	if s.eventLog != nil {
		s.eventLog.Error(1, errMsg)
	} else {
		log.Println("ERROR:", errMsg)
	}
}

// openEventLog opens the Windows event log
func openEventLog(serviceName string) (debug.Log, error) {
	elog := debug.New(serviceName)
	return elog, nil
}

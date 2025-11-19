package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/sys/windows/svc/debug"
)

type HealthChecker struct {
	config         *ServiceConfig
	processManager *ProcessManager
	eventLog       debug.Log
	ipfsHealthy    bool
	pinshareHealthy bool
	checkInterval  time.Duration
	restartDelay   time.Duration
	maxRestarts    int
	ipfsRestarts   int
	pinshareRestarts int
}

func NewHealthChecker(config *ServiceConfig, pm *ProcessManager, eventLog debug.Log) *HealthChecker {
	return &HealthChecker{
		config:         config,
		processManager: pm,
		eventLog:       eventLog,
		checkInterval:  30 * time.Second,
		restartDelay:   5 * time.Second,
		maxRestarts:    3,
	}
}

// Run starts the health checking loop
func (hc *HealthChecker) Run(ctx context.Context) {
	ticker := time.NewTicker(hc.checkInterval)
	defer ticker.Stop()

	hc.logInfo("Health checker started")

	for {
		select {
		case <-ctx.Done():
			hc.logInfo("Health checker stopped")
			return
		case <-ticker.C:
			hc.performHealthChecks(ctx)
		}
	}
}

// performHealthChecks checks health of all components
func (hc *HealthChecker) performHealthChecks(ctx context.Context) {
	// Check IPFS
	ipfsHealthy := hc.CheckIPFSHealth()
	if !ipfsHealthy && hc.ipfsHealthy {
		// IPFS just became unhealthy
		hc.logError("IPFS health check failed", nil)
		hc.handleIPFSFailure(ctx)
	} else if ipfsHealthy && !hc.ipfsHealthy {
		// IPFS recovered
		hc.logInfo("IPFS health check passed (recovered)")
		hc.ipfsRestarts = 0 // Reset restart counter
	}
	hc.ipfsHealthy = ipfsHealthy

	// Check PinShare (only if IPFS is healthy)
	if ipfsHealthy {
		pinshareHealthy := hc.CheckPinShareHealth()
		if !pinshareHealthy && hc.pinshareHealthy {
			// PinShare just became unhealthy
			hc.logError("PinShare health check failed", nil)
			hc.handlePinShareFailure(ctx)
		} else if pinshareHealthy && !hc.pinshareHealthy {
			// PinShare recovered
			hc.logInfo("PinShare health check passed (recovered)")
			hc.pinshareRestarts = 0 // Reset restart counter
		}
		hc.pinshareHealthy = pinshareHealthy
	}
}

// CheckIPFSHealth checks if IPFS is healthy
func (hc *HealthChecker) CheckIPFSHealth() bool {
	url := fmt.Sprintf("http://localhost:%d/api/v0/version", hc.config.IPFSAPIPort)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Post(url, "application/json", nil)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// CheckPinShareHealth checks if PinShare is healthy
func (hc *HealthChecker) CheckPinShareHealth() bool {
	url := fmt.Sprintf("http://localhost:%d/api/health", hc.config.PinShareAPIPort)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// handleIPFSFailure handles IPFS failure
func (hc *HealthChecker) handleIPFSFailure(ctx context.Context) {
	if hc.ipfsRestarts >= hc.maxRestarts {
		hc.logError(fmt.Sprintf("IPFS has failed %d times, not restarting", hc.maxRestarts), nil)
		return
	}

	hc.ipfsRestarts++
	hc.logInfo(fmt.Sprintf("Attempting to restart IPFS (attempt %d/%d)...", hc.ipfsRestarts, hc.maxRestarts))

	// Wait before restarting
	time.Sleep(hc.restartDelay)

	if err := hc.processManager.RestartIPFS(ctx); err != nil {
		hc.logError("Failed to restart IPFS", err)
	} else {
		hc.logInfo("IPFS restart initiated")
	}
}

// handlePinShareFailure handles PinShare failure
func (hc *HealthChecker) handlePinShareFailure(ctx context.Context) {
	if hc.pinshareRestarts >= hc.maxRestarts {
		hc.logError(fmt.Sprintf("PinShare has failed %d times, not restarting", hc.maxRestarts), nil)
		return
	}

	hc.pinshareRestarts++
	hc.logInfo(fmt.Sprintf("Attempting to restart PinShare (attempt %d/%d)...", hc.pinshareRestarts, hc.maxRestarts))

	// Wait before restarting
	time.Sleep(hc.restartDelay)

	if err := hc.processManager.RestartPinShare(ctx); err != nil {
		hc.logError("Failed to restart PinShare", err)
	} else {
		hc.logInfo("PinShare restart initiated")
	}
}

// Logging helpers
func (hc *HealthChecker) logInfo(msg string) {
	if hc.eventLog != nil {
		hc.eventLog.Info(1, msg)
	}
}

func (hc *HealthChecker) logError(msg string, err error) {
	errMsg := msg
	if err != nil {
		errMsg = fmt.Sprintf("%s: %v", msg, err)
	}
	if hc.eventLog != nil {
		hc.eventLog.Error(1, errMsg)
	}
}

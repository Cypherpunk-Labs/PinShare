// Package winservice provides shared constants and utilities for Windows service management.
package winservice

import "time"

// Service identification
const (
	// ServiceName is the Windows service name used for registration and control.
	// This must be consistent across all components (service, tray, etc.).
	ServiceName = "PinShareService"

	// ServiceDisplayName is the human-readable name shown in Windows Services.
	ServiceDisplayName = "PinShare Service"

	// ServiceDescription is the description shown in Windows Services.
	ServiceDescription = "PinShare - Decentralized IPFS pinning service with libp2p"
)

// Default ports
const (
	DefaultIPFSAPIPort     = 5001
	DefaultIPFSGatewayPort = 8080
	DefaultIPFSSwarmPort   = 4001
	DefaultPinShareAPIPort = 9090
	DefaultPinShareP2PPort = 50001
	DefaultUIPort          = 8888
)

// Timing constants
const (
	StatusCheckInterval    = 10 * time.Second
	HealthCheckInterval    = 30 * time.Second
	ServiceStartTimeout    = 30 * time.Second
	ServiceStopTimeout     = 30 * time.Second
	ServicePollInterval    = 500 * time.Millisecond
	ServiceRestartDelay    = 2 * time.Second
	ProcessShutdownTimeout = 10 * time.Second
)

// Error message limits
const (
	MaxErrorMessageLength = 50
)

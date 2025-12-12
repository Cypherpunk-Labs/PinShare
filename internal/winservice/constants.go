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

// Default port configuration - shared across all components
const (
	DefaultIPFSAPIPort     = 5001
	DefaultIPFSGatewayPort = 8080
	DefaultIPFSSwarmPort   = 4001
	DefaultPinShareAPIPort = 9090
	DefaultPinShareP2PPort = 50001
	DefaultUIPort          = 8888
)

// Service control timeouts
const (
	StatusCheckInterval    = 10 * time.Second
	HealthCheckInterval    = 30 * time.Second
	ServiceStartTimeout    = 60 * time.Second
	ServiceStopTimeout     = 30 * time.Second
	ServicePollInterval    = 300 * time.Millisecond
	ServiceRestartDelay    = 2 * time.Second
	ProcessShutdownTimeout = 10 * time.Second
)

// Recovery action delays for Windows service manager
const (
	RecoveryDelayFirst  = 5 * time.Second
	RecoveryDelaySecond = 10 * time.Second
	RecoveryDelayThird  = 30 * time.Second
	RecoveryResetPeriod = 60 // seconds
)

// Error message limits
const (
	MaxErrorMessageLength = 50
)

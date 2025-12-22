package winservice

import "time"

// Service identification
const (
	ServiceName        = "PinShareService"
	ServiceDisplayName = "PinShare Service"
	ServiceDescription = "PinShare decentralized IPFS pinning service"
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

	// Startup wait timeouts
	IPFSStartTimeout     = 30 * time.Second
	PinShareStartTimeout = 60 * time.Second
	HealthCheckPoll      = 1 * time.Second

	// Service recovery delays
	RecoveryDelayFirst  = 5 * time.Second
	RecoveryDelaySecond = 10 * time.Second
	RecoveryDelayThird  = 30 * time.Second
	RecoveryResetPeriod = 60 // seconds
)

// Error message limits
const (
	MaxErrorMessageLength = 50
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

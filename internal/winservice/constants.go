// Package winservice provides shared constants and utilities for Windows service management.
package winservice

import "time"

// Application identity - shared across all components
const (
	// AppName is the application name used for directory paths and display.
	AppName = "PinShare"

	// ServiceName is the Windows service name used for registration and control.
	// This must be consistent across all components (service, tray, etc.).
	ServiceName = "PinShareService"

	// ServiceDisplayName is the human-readable name shown in Windows Services.
	ServiceDisplayName = "PinShare Service"

	// ServiceDescription is the description shown in Windows Services.
	ServiceDescription = "PinShare - Decentralized IPFS pinning service with libp2p"

	// Version is the current version of PinShare.
	// This should be updated for each release.
	Version = "0.1.3"
)

// Directory names within the data directory
const (
	DirIPFS     = "ipfs"
	DirPinShare = "pinshare"
	DirUpload   = "upload"
	DirCache    = "cache"
	DirRejected = "rejected"
	DirLogs     = "logs"
)

// File names
const (
	FileConfig     = "config.json"
	FileSession    = "session.json"
	FileServiceLog = "service.log"
)

// Environment variable names
const (
	EnvLocalAppData = "LOCALAPPDATA"
	EnvUserProfile  = "USERPROFILE"
	EnvProgramData  = "PROGRAMDATA"
	EnvProgramFiles = "PROGRAMFILES"
	EnvUsername     = "USERNAME"
)

// Default paths when environment variables are not available
const (
	DefaultLocalAppDataPath = `C:\Users\Default\AppData\Local`
	DefaultProgramDataPath  = `C:\ProgramData`
	DefaultProgramFilesPath = `C:\Program Files`
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

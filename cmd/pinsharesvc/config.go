package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"pinshare/internal/winservice"
)

// Default organization settings (service-specific)
const (
	defaultOrgName   = "MyOrganization"
	defaultGroupName = "MyGroup"
	defaultLogLevel  = "info"
)

// SessionMarker contains user session information written by the tray app
// This allows the SYSTEM service to find the current user's data directory
type SessionMarker struct {
	LocalAppData string    `json:"local_app_data"`
	Username     string    `json:"username"`
	Timestamp    time.Time `json:"timestamp"`
}

// getSessionMarkerPath returns the path to the session marker file
func getSessionMarkerPath() string {
	programData := os.Getenv(winservice.EnvProgramData)
	if programData == "" {
		programData = winservice.DefaultProgramDataPath
	}
	return filepath.Join(programData, winservice.ServiceDisplayName, winservice.FileSession)
}

// loadSessionMarker reads the session marker written by the tray app
func loadSessionMarker() (*SessionMarker, error) {
	markerPath := getSessionMarkerPath()
	data, err := os.ReadFile(markerPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read session marker: %w", err)
	}

	marker := &SessionMarker{}
	if err := json.Unmarshal(data, marker); err != nil {
		return nil, fmt.Errorf("failed to parse session marker: %w", err)
	}

	return marker, nil
}

// getUserDataDirectory returns the user's data directory from the session marker
// Falls back to LOCALAPPDATA env var if running in user context (e.g., debug mode)
func getUserDataDirectory() (string, error) {
	// First try to read session marker (for SYSTEM service context)
	marker, err := loadSessionMarker()
	if err == nil && marker.LocalAppData != "" {
		return filepath.Join(marker.LocalAppData, winservice.AppName), nil
	}

	// Fall back to LOCALAPPDATA (for user context, e.g., debug mode)
	localAppData := os.Getenv(winservice.EnvLocalAppData)
	if localAppData != "" {
		return filepath.Join(localAppData, winservice.AppName), nil
	}

	// Last resort: try to construct from USERPROFILE
	userProfile := os.Getenv(winservice.EnvUserProfile)
	if userProfile != "" {
		return filepath.Join(userProfile, "AppData", "Local", winservice.AppName), nil
	}

	return "", fmt.Errorf("cannot determine user data directory: no session marker and %s not set", winservice.EnvLocalAppData)
}

// EncryptionKeyLength is the length in bytes for generated encryption keys
const EncryptionKeyLength = 32

type ServiceConfig struct {
	// Installation paths
	InstallDirectory string `json:"install_directory"`
	DataDirectory    string `json:"data_directory"`

	// Binary paths
	IPFSBinary     string `json:"ipfs_binary"`
	PinShareBinary string `json:"pinshare_binary"`

	// Ports
	IPFSAPIPort     int `json:"ipfs_api_port"`
	IPFSGatewayPort int `json:"ipfs_gateway_port"`
	IPFSSwarmPort   int `json:"ipfs_swarm_port"`
	PinShareAPIPort int `json:"pinshare_api_port"`
	PinShareP2PPort int `json:"pinshare_p2p_port"`
	UIPort          int `json:"ui_port"` // Reserved for future web UI integration

	// PinShare configuration
	OrgName   string `json:"org_name"`
	GroupName string `json:"group_name"`

	// Feature flags
	SkipVirusTotal bool `json:"skip_virus_total"`
	EnableCache    bool `json:"enable_cache"`
	ArchiveNode    bool `json:"archive_node"`

	// Security
	VirusTotalToken string `json:"virus_total_token,omitempty"`
	EncryptionKey   string `json:"encryption_key"`

	// Logging
	LogLevel    string `json:"log_level"`
	LogFilePath string `json:"log_file_path"`
}

// LoadConfig loads configuration from JSON file
func LoadConfig() (*ServiceConfig, error) {
	config, err := loadFromFile()
	if err != nil {
		// Use defaults if config file doesn't exist or can't be read
		return getDefaultConfig()
	}
	return config, nil
}

// loadFromFile loads configuration from JSON file in user's LOCALAPPDATA
func loadFromFile() (*ServiceConfig, error) {
	dataDir, err := getUserDataDirectory()
	if err != nil {
		return nil, fmt.Errorf("failed to determine data directory: %w", err)
	}

	configPath := filepath.Join(dataDir, winservice.FileConfig)

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := &ServiceConfig{}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	config.applyDefaults()
	return config, nil
}

// getDefaultConfig returns a configuration with default values
func getDefaultConfig() (*ServiceConfig, error) {
	programFiles := os.Getenv(winservice.EnvProgramFiles)
	if programFiles == "" {
		programFiles = winservice.DefaultProgramFilesPath
	}

	installDir := filepath.Join(programFiles, winservice.AppName)

	// Get user data directory from session marker or LOCALAPPDATA
	dataDir, err := getUserDataDirectory()
	if err != nil {
		return nil, fmt.Errorf("failed to determine data directory: %w", err)
	}

	config := &ServiceConfig{
		InstallDirectory: installDir,
		DataDirectory:    dataDir,
		IPFSBinary:       filepath.Join(installDir, "ipfs.exe"),
		PinShareBinary:   filepath.Join(installDir, "pinshare.exe"),

		IPFSAPIPort:     winservice.DefaultIPFSAPIPort,
		IPFSGatewayPort: winservice.DefaultIPFSGatewayPort,
		IPFSSwarmPort:   winservice.DefaultIPFSSwarmPort,
		PinShareAPIPort: winservice.DefaultPinShareAPIPort,
		PinShareP2PPort: winservice.DefaultPinShareP2PPort,
		UIPort:          winservice.DefaultUIPort,

		OrgName:   defaultOrgName,
		GroupName: defaultGroupName,

		SkipVirusTotal: false, // Default to enabled; note: without VT_TOKEN, scanning is auto-skipped in service context
		EnableCache:    true,
		ArchiveNode:    false,

		EncryptionKey: generateEncryptionKey(),

		LogLevel:    defaultLogLevel,
		LogFilePath: filepath.Join(dataDir, winservice.DirLogs, winservice.FileServiceLog),
	}

	return config, nil
}

// applyDefaults fills in missing configuration values with defaults
func (c *ServiceConfig) applyDefaults() {
	if c.IPFSAPIPort == 0 {
		c.IPFSAPIPort = winservice.DefaultIPFSAPIPort
	}
	if c.IPFSGatewayPort == 0 {
		c.IPFSGatewayPort = winservice.DefaultIPFSGatewayPort
	}
	if c.IPFSSwarmPort == 0 {
		c.IPFSSwarmPort = winservice.DefaultIPFSSwarmPort
	}
	if c.PinShareAPIPort == 0 {
		c.PinShareAPIPort = winservice.DefaultPinShareAPIPort
	}
	if c.PinShareP2PPort == 0 {
		c.PinShareP2PPort = winservice.DefaultPinShareP2PPort
	}
	if c.UIPort == 0 {
		c.UIPort = winservice.DefaultUIPort
	}
	if c.LogLevel == "" {
		c.LogLevel = defaultLogLevel
	}
	if c.OrgName == "" {
		c.OrgName = defaultOrgName
	}
	if c.GroupName == "" {
		c.GroupName = defaultGroupName
	}
	if c.EncryptionKey == "" {
		c.EncryptionKey = generateEncryptionKey()
	}

	// Set default paths if not specified
	if c.DataDirectory == "" {
		// getUserDataDirectory already has comprehensive fallback logic,
		// so if it fails, there's no reasonable default we can use
		if dataDir, err := getUserDataDirectory(); err == nil {
			c.DataDirectory = dataDir
		}
	}

	if c.InstallDirectory == "" {
		programFiles := os.Getenv(winservice.EnvProgramFiles)
		if programFiles == "" {
			programFiles = winservice.DefaultProgramFilesPath
		}
		c.InstallDirectory = filepath.Join(programFiles, winservice.AppName)
	}

	if c.IPFSBinary == "" {
		c.IPFSBinary = filepath.Join(c.InstallDirectory, "ipfs.exe")
	}

	if c.PinShareBinary == "" {
		c.PinShareBinary = filepath.Join(c.InstallDirectory, "pinshare.exe")
	}

	if c.LogFilePath == "" {
		c.LogFilePath = filepath.Join(c.DataDirectory, winservice.DirLogs, winservice.FileServiceLog)
	}
}

// EnsureDirectories creates all required directories
func (c *ServiceConfig) EnsureDirectories() error {
	dirs := []string{
		c.DataDirectory,
		filepath.Join(c.DataDirectory, winservice.DirIPFS),
		filepath.Join(c.DataDirectory, winservice.DirPinShare),
		filepath.Join(c.DataDirectory, winservice.DirUpload),
		filepath.Join(c.DataDirectory, winservice.DirCache),
		filepath.Join(c.DataDirectory, winservice.DirRejected),
		filepath.Join(c.DataDirectory, winservice.DirLogs),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// GetIPFSDataPath returns the IPFS data directory path
func (c *ServiceConfig) GetIPFSDataPath() string {
	return filepath.Join(c.DataDirectory, winservice.DirIPFS)
}

// GetPinShareDataPath returns the PinShare data directory
func (c *ServiceConfig) GetPinShareDataPath() string {
	return filepath.Join(c.DataDirectory, winservice.DirPinShare)
}

// SaveToFile saves the configuration to a JSON file
func (c *ServiceConfig) SaveToFile() error {
	configPath := filepath.Join(c.DataDirectory, winservice.FileConfig)

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// generateEncryptionKey generates a cryptographically secure random encryption key
func generateEncryptionKey() string {
	bytes := make([]byte, EncryptionKeyLength)
	if _, err := rand.Read(bytes); err != nil {
		// If random generation fails, panic as this is a critical security requirement
		panic(fmt.Sprintf("failed to generate encryption key: %v", err))
	}
	return hex.EncodeToString(bytes)
}

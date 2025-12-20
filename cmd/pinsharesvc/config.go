package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"pinshare/internal/winservice"
)

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
	if err == nil {
		return config, nil
	}

	// Use defaults if config file doesn't exist
	return getDefaultConfig()
}

// loadFromFile loads configuration from JSON file
func loadFromFile() (*ServiceConfig, error) {
	programData := os.Getenv("PROGRAMDATA")
	if programData == "" {
		programData = `C:\ProgramData`
	}

	configPath := filepath.Join(programData, "PinShare", "config.json")

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
	programData := os.Getenv("PROGRAMDATA")
	if programData == "" {
		programData = `C:\ProgramData`
	}

	programFiles := os.Getenv("PROGRAMFILES")
	if programFiles == "" {
		programFiles = `C:\Program Files`
	}

	installDir := filepath.Join(programFiles, "PinShare")
	dataDir := filepath.Join(programData, "PinShare")

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

		OrgName:   "MyOrganization",
		GroupName: "MyGroup",

		SkipVirusTotal: false, // Default to enabled; note: without VT_TOKEN, scanning is auto-skipped in service context
		EnableCache:    true,
		ArchiveNode:    false,

		EncryptionKey: generateEncryptionKey(),

		LogLevel:    "info",
		LogFilePath: filepath.Join(dataDir, "logs", "service.log"),
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
		c.LogLevel = "info"
	}
	if c.OrgName == "" {
		c.OrgName = "MyOrganization"
	}
	if c.GroupName == "" {
		c.GroupName = "MyGroup"
	}
	if c.EncryptionKey == "" {
		c.EncryptionKey = generateEncryptionKey()
	}

	// Set default paths if not specified
	if c.DataDirectory == "" {
		programData := os.Getenv("PROGRAMDATA")
		if programData == "" {
			programData = `C:\ProgramData`
		}
		c.DataDirectory = filepath.Join(programData, "PinShare")
	}

	if c.InstallDirectory == "" {
		programFiles := os.Getenv("PROGRAMFILES")
		if programFiles == "" {
			programFiles = `C:\Program Files`
		}
		c.InstallDirectory = filepath.Join(programFiles, "PinShare")
	}

	if c.IPFSBinary == "" {
		c.IPFSBinary = filepath.Join(c.InstallDirectory, "ipfs.exe")
	}

	if c.PinShareBinary == "" {
		c.PinShareBinary = filepath.Join(c.InstallDirectory, "pinshare.exe")
	}

	if c.LogFilePath == "" {
		c.LogFilePath = filepath.Join(c.DataDirectory, "logs", "service.log")
	}
}

// EnsureDirectories creates all required directories
func (c *ServiceConfig) EnsureDirectories() error {
	dirs := []string{
		c.DataDirectory,
		filepath.Join(c.DataDirectory, "ipfs"),
		filepath.Join(c.DataDirectory, "pinshare"),
		filepath.Join(c.DataDirectory, "upload"),
		filepath.Join(c.DataDirectory, "cache"),
		filepath.Join(c.DataDirectory, "rejected"),
		filepath.Join(c.DataDirectory, "logs"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// GetIPFSRepoPath returns the IPFS repository path
func (c *ServiceConfig) GetIPFSRepoPath() string {
	return filepath.Join(c.DataDirectory, "ipfs")
}

// GetPinShareDataPath returns the PinShare data directory
func (c *ServiceConfig) GetPinShareDataPath() string {
	return filepath.Join(c.DataDirectory, "pinshare")
}

// SaveToFile saves the configuration to a JSON file
func (c *ServiceConfig) SaveToFile() error {
	configPath := filepath.Join(c.DataDirectory, "config.json")

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// generateEncryptionKey generates a cryptographically secure random 32-byte encryption key
func generateEncryptionKey() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		// If random generation fails, panic as this is a critical security requirement
		panic(fmt.Sprintf("failed to generate encryption key: %v", err))
	}
	return hex.EncodeToString(bytes)
}

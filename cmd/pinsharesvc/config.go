package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

const (
	registryPath = `SOFTWARE\PinShare`

	// Default ports
	defaultIPFSAPIPort    = 5001
	defaultIPFSGatewayPort = 8080
	defaultIPFSSwarmPort  = 4001
	defaultPinShareAPIPort = 9090
	defaultPinShareP2PPort = 50001
	defaultUIPort         = 8888
)

type ServiceConfig struct {
	// Installation paths
	InstallDirectory string `json:"install_directory"`
	DataDirectory    string `json:"data_directory"`

	// Binary paths
	IPFSBinary     string `json:"ipfs_binary"`
	PinShareBinary string `json:"pinshare_binary"`

	// Ports
	IPFSAPIPort      int `json:"ipfs_api_port"`
	IPFSGatewayPort  int `json:"ipfs_gateway_port"`
	IPFSSwarmPort    int `json:"ipfs_swarm_port"`
	PinShareAPIPort  int `json:"pinshare_api_port"`
	PinShareP2PPort  int `json:"pinshare_p2p_port"`
	UIPort           int `json:"ui_port"`

	// PinShare configuration
	OrgName    string `json:"org_name"`
	GroupName  string `json:"group_name"`

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

// LoadConfig loads configuration from registry or file
func LoadConfig() (*ServiceConfig, error) {
	// Try loading from registry first
	config, err := loadFromRegistry()
	if err == nil {
		return config, nil
	}

	// Fall back to file-based config
	config, err = loadFromFile()
	if err == nil {
		return config, nil
	}

	// Use defaults if both fail
	return getDefaultConfig()
}

// loadFromRegistry loads configuration from Windows registry
func loadFromRegistry() (*ServiceConfig, error) {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, registryPath, registry.QUERY_VALUE)
	if err != nil {
		return nil, fmt.Errorf("failed to open registry key: %w", err)
	}
	defer key.Close()

	config := &ServiceConfig{}

	// Read string values
	config.InstallDirectory, _, _ = key.GetStringValue("InstallDirectory")
	config.DataDirectory, _, _ = key.GetStringValue("DataDirectory")
	config.IPFSBinary, _, _ = key.GetStringValue("IPFSBinary")
	config.PinShareBinary, _, _ = key.GetStringValue("PinShareBinary")
	config.OrgName, _, _ = key.GetStringValue("OrgName")
	config.GroupName, _, _ = key.GetStringValue("GroupName")
	config.VirusTotalToken, _, _ = key.GetStringValue("VirusTotalToken")
	config.EncryptionKey, _, _ = key.GetStringValue("EncryptionKey")
	config.LogLevel, _, _ = key.GetStringValue("LogLevel")
	config.LogFilePath, _, _ = key.GetStringValue("LogFilePath")

	// Read integer values
	ipfsAPIPort, _, err := key.GetIntegerValue("IPFSAPIPort")
	if err == nil {
		config.IPFSAPIPort = int(ipfsAPIPort)
	}

	ipfsGatewayPort, _, err := key.GetIntegerValue("IPFSGatewayPort")
	if err == nil {
		config.IPFSGatewayPort = int(ipfsGatewayPort)
	}

	ipfsSwarmPort, _, err := key.GetIntegerValue("IPFSSwarmPort")
	if err == nil {
		config.IPFSSwarmPort = int(ipfsSwarmPort)
	}

	pinshareAPIPort, _, err := key.GetIntegerValue("PinShareAPIPort")
	if err == nil {
		config.PinShareAPIPort = int(pinshareAPIPort)
	}

	pinshareP2PPort, _, err := key.GetIntegerValue("PinShareP2PPort")
	if err == nil {
		config.PinShareP2PPort = int(pinshareP2PPort)
	}

	uiPort, _, err := key.GetIntegerValue("UIPort")
	if err == nil {
		config.UIPort = int(uiPort)
	}

	// Read boolean values (stored as integers 0/1, with fallback to string for backwards compatibility)
	skipVT, _, err := key.GetIntegerValue("SkipVirusTotal")
	if err == nil {
		config.SkipVirusTotal = skipVT != 0
	} else {
		// Fallback: try reading as string (for old installs that used REG_SZ)
		if strVal, _, strErr := key.GetStringValue("SkipVirusTotal"); strErr == nil && strVal != "" {
			config.SkipVirusTotal = strVal == "1" || strVal == "true"
		}
	}

	enableCache, _, err := key.GetIntegerValue("EnableCache")
	if err == nil {
		config.EnableCache = enableCache != 0
	} else {
		// Fallback: try reading as string (for old installs that used REG_SZ)
		if strVal, _, strErr := key.GetStringValue("EnableCache"); strErr == nil && strVal != "" {
			config.EnableCache = strVal == "1" || strVal == "true"
		}
	}

	archiveNode, _, err := key.GetIntegerValue("ArchiveNode")
	if err == nil {
		config.ArchiveNode = archiveNode != 0
	} else {
		// Fallback: try reading as string (for old installs that used REG_SZ)
		if strVal, _, strErr := key.GetStringValue("ArchiveNode"); strErr == nil && strVal != "" {
			config.ArchiveNode = strVal == "1" || strVal == "true"
		}
	}

	// Apply defaults for missing values
	config.applyDefaults()

	return config, nil
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

		IPFSAPIPort:      defaultIPFSAPIPort,
		IPFSGatewayPort:  defaultIPFSGatewayPort,
		IPFSSwarmPort:    defaultIPFSSwarmPort,
		PinShareAPIPort:  defaultPinShareAPIPort,
		PinShareP2PPort:  defaultPinShareP2PPort,
		UIPort:           defaultUIPort,

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
		c.IPFSAPIPort = defaultIPFSAPIPort
	}
	if c.IPFSGatewayPort == 0 {
		c.IPFSGatewayPort = defaultIPFSGatewayPort
	}
	if c.IPFSSwarmPort == 0 {
		c.IPFSSwarmPort = defaultIPFSSwarmPort
	}
	if c.PinShareAPIPort == 0 {
		c.PinShareAPIPort = defaultPinShareAPIPort
	}
	if c.PinShareP2PPort == 0 {
		c.PinShareP2PPort = defaultPinShareP2PPort
	}
	if c.UIPort == 0 {
		c.UIPort = defaultUIPort
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

// SaveToRegistry saves the configuration to Windows registry
func (c *ServiceConfig) SaveToRegistry() error {
	key, _, err := registry.CreateKey(registry.LOCAL_MACHINE, registryPath, registry.ALL_ACCESS)
	if err != nil {
		return fmt.Errorf("failed to create registry key: %w", err)
	}
	defer key.Close()

	// Write string values
	_ = key.SetStringValue("InstallDirectory", c.InstallDirectory)
	_ = key.SetStringValue("DataDirectory", c.DataDirectory)
	_ = key.SetStringValue("IPFSBinary", c.IPFSBinary)
	_ = key.SetStringValue("PinShareBinary", c.PinShareBinary)
	_ = key.SetStringValue("OrgName", c.OrgName)
	_ = key.SetStringValue("GroupName", c.GroupName)
	_ = key.SetStringValue("VirusTotalToken", c.VirusTotalToken)
	_ = key.SetStringValue("EncryptionKey", c.EncryptionKey)
	_ = key.SetStringValue("LogLevel", c.LogLevel)
	_ = key.SetStringValue("LogFilePath", c.LogFilePath)

	// Write integer values
	_ = key.SetDWordValue("IPFSAPIPort", uint32(c.IPFSAPIPort))
	_ = key.SetDWordValue("IPFSGatewayPort", uint32(c.IPFSGatewayPort))
	_ = key.SetDWordValue("IPFSSwarmPort", uint32(c.IPFSSwarmPort))
	_ = key.SetDWordValue("PinShareAPIPort", uint32(c.PinShareAPIPort))
	_ = key.SetDWordValue("PinShareP2PPort", uint32(c.PinShareP2PPort))
	_ = key.SetDWordValue("UIPort", uint32(c.UIPort))

	// Write boolean values (as integers 0/1)
	skipVT := uint32(0)
	if c.SkipVirusTotal {
		skipVT = 1
	}
	_ = key.SetDWordValue("SkipVirusTotal", skipVT)

	enableCache := uint32(0)
	if c.EnableCache {
		enableCache = 1
	}
	_ = key.SetDWordValue("EnableCache", enableCache)

	archiveNode := uint32(0)
	if c.ArchiveNode {
		archiveNode = 1
	}
	_ = key.SetDWordValue("ArchiveNode", archiveNode)

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

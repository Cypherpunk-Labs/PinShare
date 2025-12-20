package main

import (
	"encoding/json"
	"os"
	"path/filepath"

	"pinshare/internal/winservice"
)

// TrayConfig holds configuration values needed by the tray application
type TrayConfig struct {
	IPFSAPIPort     int `json:"ipfs_api_port"`
	PinShareAPIPort int `json:"pinshare_api_port"`
}

// Global config instance
var appConfig *TrayConfig

// loadConfig loads configuration from config.json
// Falls back to defaults if config file doesn't exist or can't be read
func loadConfig() *TrayConfig {
	config := &TrayConfig{
		IPFSAPIPort:     winservice.DefaultIPFSAPIPort,
		PinShareAPIPort: winservice.DefaultPinShareAPIPort,
	}

	programData := os.Getenv("PROGRAMDATA")
	if programData == "" {
		programData = `C:\ProgramData`
	}

	configPath := filepath.Join(programData, "PinShare", "config.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		// Config file doesn't exist yet, use defaults
		return config
	}

	// Parse only the fields we need
	if err := json.Unmarshal(data, config); err != nil {
		// Invalid JSON, use defaults
		return config
	}

	// Apply defaults for zero values
	if config.IPFSAPIPort == 0 {
		config.IPFSAPIPort = winservice.DefaultIPFSAPIPort
	}
	if config.PinShareAPIPort == 0 {
		config.PinShareAPIPort = winservice.DefaultPinShareAPIPort
	}

	return config
}

// getConfig returns the current configuration, loading it if necessary
func getConfig() *TrayConfig {
	if appConfig == nil {
		appConfig = loadConfig()
	}
	return appConfig
}

// reloadConfig forces a reload of the configuration
func reloadConfig() {
	appConfig = loadConfig()
}

// Reload reloads the configuration from disk
func (c *TrayConfig) Reload() {
	newConfig := loadConfig()
	c.IPFSAPIPort = newConfig.IPFSAPIPort
	c.PinShareAPIPort = newConfig.PinShareAPIPort
}

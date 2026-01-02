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

// getUserDataDirectory returns the user's PinShare data directory
func getUserDataDirectory() string {
	localAppData := os.Getenv(envLocalAppData)
	if localAppData == "" {
		// Fall back to constructing from USERPROFILE
		userProfile := os.Getenv(envUserProfile)
		if userProfile != "" {
			localAppData = filepath.Join(userProfile, "AppData", "Local")
		} else {
			// Last resort
			localAppData = defaultLocalAppDataPath
		}
	}
	return filepath.Join(localAppData, appName)
}

// loadConfig loads configuration from config.json in user's LOCALAPPDATA
// Falls back to defaults if config file doesn't exist or can't be read
func loadConfig() *TrayConfig {
	config := &TrayConfig{
		IPFSAPIPort:     winservice.DefaultIPFSAPIPort,
		PinShareAPIPort: winservice.DefaultPinShareAPIPort,
	}

	dataDir := getUserDataDirectory()
	configPath := filepath.Join(dataDir, fileConfig)

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

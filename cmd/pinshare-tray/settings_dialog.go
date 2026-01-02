package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"pinshare/internal/winservice"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"golang.org/x/sys/windows"
)

// Settings dialog constants
const (
	// Default organization and group names for new installations
	defaultOrgName   = "MyOrganization"
	defaultGroupName = "MyGroup"

	// Default log level
	defaultLogLevel = "info"

	// Elevated copy operation timeouts
	elevatedCopyMaxWait      = 30 * time.Second
	elevatedCopyPollInterval = 500 * time.Millisecond

	// Lock file removal settings
	lockFileRemovalMaxAttempts = 3
	lockFileRemovalRetryDelay  = 1 * time.Second

	// Dialog dimensions
	dialogMinWidth  = 480
	dialogMinHeight = 440
	dialogWidth     = 500
	dialogHeight    = 460
	gridSpacing     = 10
	gridColumns     = 3

	// Port input constraints (standard TCP port range)
	portMinValue = 1
	portMaxValue = 65535
)

// logLevels defines the available log levels in order
var logLevels = []string{"debug", "info", "warn", "error"}

// FullConfig holds all configuration values from config.json
type FullConfig struct {
	// Installation paths (read-only)
	InstallDirectory string `json:"install_directory"`
	DataDirectory    string `json:"data_directory"`
	IPFSBinary       string `json:"ipfs_binary"`
	PinShareBinary   string `json:"pinshare_binary"`

	// Ports
	IPFSAPIPort     int `json:"ipfs_api_port"`
	IPFSGatewayPort int `json:"ipfs_gateway_port"`
	IPFSSwarmPort   int `json:"ipfs_swarm_port"`
	PinShareAPIPort int `json:"pinshare_api_port"`
	PinShareP2PPort int `json:"pinshare_p2p_port"`
	UIPort          int `json:"ui_port"`

	// Organization
	OrgName   string `json:"org_name"`
	GroupName string `json:"group_name"`

	// Features
	SkipVirusTotal bool `json:"skip_virus_total"`
	EnableCache    bool `json:"enable_cache"`
	ArchiveNode    bool `json:"archive_node"`

	// Security
	VirusTotalToken string `json:"virus_total_token,omitempty"`
	EncryptionKey   string `json:"encryption_key,omitempty"`

	// Logging
	LogLevel    string `json:"log_level"`
	LogFilePath string `json:"log_file_path,omitempty"`
}

// SettingsDialog manages the native Windows settings dialog
type SettingsDialog struct {
	config     *FullConfig
	configPath string

	// Dialog and main controls
	dlg        *walk.Dialog
	tabWidget  *walk.TabWidget
	saveButton *walk.PushButton

	// Port fields
	ipfsAPIPortEdit     *walk.NumberEdit
	ipfsGatewayPortEdit *walk.NumberEdit
	ipfsSwarmPortEdit   *walk.NumberEdit
	pinshareAPIPortEdit *walk.NumberEdit
	pinshareP2PPortEdit *walk.NumberEdit
	uiPortEdit          *walk.NumberEdit

	// Organization fields
	orgNameEdit   *walk.LineEdit
	groupNameEdit *walk.LineEdit

	// Feature checkboxes
	skipVTCheckbox     *walk.CheckBox
	enableCacheCheckbox *walk.CheckBox
	archiveNodeCheckbox *walk.CheckBox

	// Security fields
	vtTokenEdit   *walk.LineEdit
	logLevelCombo *walk.ComboBox
}

// loadFullConfig loads the complete configuration from config.json in user's LOCALAPPDATA
func loadFullConfig() (*FullConfig, string, error) {
	dataDir := getUserDataDirectory()
	configPath := filepath.Join(dataDir, fileConfig)

	config := &FullConfig{
		// Defaults
		IPFSAPIPort:     winservice.DefaultIPFSAPIPort,
		IPFSGatewayPort: winservice.DefaultIPFSGatewayPort,
		IPFSSwarmPort:   winservice.DefaultIPFSSwarmPort,
		PinShareAPIPort: winservice.DefaultPinShareAPIPort,
		PinShareP2PPort: winservice.DefaultPinShareP2PPort,
		UIPort:          winservice.DefaultUIPort,
		OrgName:         defaultOrgName,
		GroupName:       defaultGroupName,
		SkipVirusTotal:  false,
		EnableCache:     true,
		ArchiveNode:     false,
		LogLevel:        defaultLogLevel,
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		// Config doesn't exist, use defaults
		return config, configPath, nil
	}

	if err := json.Unmarshal(data, config); err != nil {
		log.Printf("Failed to parse config, using defaults: %v", err)
		return config, configPath, nil
	}

	// Apply defaults for any zero values
	if config.IPFSAPIPort == 0 {
		config.IPFSAPIPort = winservice.DefaultIPFSAPIPort
	}
	if config.IPFSGatewayPort == 0 {
		config.IPFSGatewayPort = winservice.DefaultIPFSGatewayPort
	}
	if config.IPFSSwarmPort == 0 {
		config.IPFSSwarmPort = winservice.DefaultIPFSSwarmPort
	}
	if config.PinShareAPIPort == 0 {
		config.PinShareAPIPort = winservice.DefaultPinShareAPIPort
	}
	if config.PinShareP2PPort == 0 {
		config.PinShareP2PPort = winservice.DefaultPinShareP2PPort
	}
	if config.UIPort == 0 {
		config.UIPort = winservice.DefaultUIPort
	}
	if config.OrgName == "" {
		config.OrgName = defaultOrgName
	}
	if config.GroupName == "" {
		config.GroupName = defaultGroupName
	}
	if config.LogLevel == "" {
		config.LogLevel = defaultLogLevel
	}

	return config, configPath, nil
}

// saveConfig saves the configuration to config.json
// It first tries direct write, then falls back to elevated copy if access is denied
func (sd *SettingsDialog) saveConfig() error {
	// Update config from UI controls
	sd.config.IPFSAPIPort = int(sd.ipfsAPIPortEdit.Value())
	sd.config.IPFSGatewayPort = int(sd.ipfsGatewayPortEdit.Value())
	sd.config.IPFSSwarmPort = int(sd.ipfsSwarmPortEdit.Value())
	sd.config.PinShareAPIPort = int(sd.pinshareAPIPortEdit.Value())
	sd.config.PinShareP2PPort = int(sd.pinshareP2PPortEdit.Value())
	sd.config.UIPort = int(sd.uiPortEdit.Value())

	sd.config.OrgName = sd.orgNameEdit.Text()
	sd.config.GroupName = sd.groupNameEdit.Text()

	sd.config.SkipVirusTotal = sd.skipVTCheckbox.Checked()
	sd.config.EnableCache = sd.enableCacheCheckbox.Checked()
	sd.config.ArchiveNode = sd.archiveNodeCheckbox.Checked()

	// Only update VT token if something was entered
	if token := sd.vtTokenEdit.Text(); token != "" {
		sd.config.VirusTotalToken = token
	}

	if idx := sd.logLevelCombo.CurrentIndex(); idx >= 0 {
		sd.config.LogLevel = logLevels[idx]
	}

	// Serialize to JSON
	data, err := json.MarshalIndent(sd.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	// Try direct write first
	err = os.WriteFile(sd.configPath, data, 0644)
	if err == nil {
		log.Printf("Config saved directly to %s", sd.configPath)
		return nil
	}

	// Check if it's an access denied error
	if !isAccessDeniedError(err) {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	log.Printf("Direct write failed (access denied), trying elevated copy...")

	// Write to temp file first
	tempFile, err := os.CreateTemp("", "pinshare-config-*.json")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tempPath := tempFile.Name()

	if _, err := tempFile.Write(data); err != nil {
		tempFile.Close()
		os.Remove(tempPath)
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	tempFile.Close()

	// Use elevated copy command
	err = saveConfigElevated(tempPath, sd.configPath)
	os.Remove(tempPath) // Clean up temp file

	if err != nil {
		return fmt.Errorf("failed to save config with elevation: %w", err)
	}

	log.Printf("Config saved with elevation to %s", sd.configPath)
	return nil
}

// saveConfigElevated copies the config file using an elevated process
func saveConfigElevated(srcPath, dstPath string) error {
	// Get the expected file size from source
	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		return fmt.Errorf("failed to stat source file: %w", err)
	}
	expectedSize := srcInfo.Size()

	// Get the original modification time of destination (if it exists)
	var originalModTime int64
	if dstInfo, err := os.Stat(dstPath); err == nil {
		originalModTime = dstInfo.ModTime().UnixNano()
	}

	// Use cmd.exe /c copy to copy the file with elevation
	args := fmt.Sprintf(`/c copy /Y "%s" "%s"`, srcPath, dstPath)

	verbPtr, _ := windows.UTF16PtrFromString("runas")
	exePtr, _ := windows.UTF16PtrFromString("cmd.exe")
	argsPtr, _ := windows.UTF16PtrFromString(args)

	err = windows.ShellExecute(0, verbPtr, exePtr, argsPtr, nil, windows.SW_HIDE)
	if err != nil {
		return fmt.Errorf("failed to execute elevated copy: %w", err)
	}

	// ShellExecute returns immediately, so poll for the file to be updated
	// Wait for UAC prompt + copy operation
	log.Printf("Waiting for elevated copy to complete...")
	deadline := time.Now().Add(elevatedCopyMaxWait)

	for time.Now().Before(deadline) {
		time.Sleep(elevatedCopyPollInterval)

		// Check if destination was updated
		dstInfo, err := os.Stat(dstPath)
		if err != nil {
			continue // File might not exist yet
		}

		// Check if modification time changed and size matches expected
		if dstInfo.ModTime().UnixNano() != originalModTime && dstInfo.Size() == expectedSize {
			log.Printf("Config file updated: %s (size: %d bytes)", dstPath, dstInfo.Size())
			return nil
		}
	}

	return fmt.Errorf("timeout waiting for elevated copy to complete (UAC may have been cancelled)")
}

// isAccessDeniedError checks if an error is an access denied error
func isAccessDeniedError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "Access is denied") ||
		strings.Contains(errStr, "access denied") ||
		strings.Contains(errStr, "permission denied")
}

// createNetworkPortsTab creates the Network Ports tab page
func (sd *SettingsDialog) createNetworkPortsTab(config *FullConfig) TabPage {
	return TabPage{
		Title:  "Network Ports",
		Layout: Grid{Columns: gridColumns, Spacing: gridSpacing},
		Children: []Widget{
			Label{Text: "IPFS API Port:"},
			NumberEdit{
				AssignTo: &sd.ipfsAPIPortEdit,
				Value:    float64(config.IPFSAPIPort),
				MinValue: portMinValue,
				MaxValue: portMaxValue,
				Decimals: 0,
			},
			Label{Text: fmt.Sprintf("(default: %d)", winservice.DefaultIPFSAPIPort), TextColor: walk.RGB(128, 128, 128)},

			Label{Text: "IPFS Gateway Port:"},
			NumberEdit{
				AssignTo: &sd.ipfsGatewayPortEdit,
				Value:    float64(config.IPFSGatewayPort),
				MinValue: portMinValue,
				MaxValue: portMaxValue,
				Decimals: 0,
			},
			Label{Text: fmt.Sprintf("(default: %d)", winservice.DefaultIPFSGatewayPort), TextColor: walk.RGB(128, 128, 128)},

			Label{Text: "IPFS Swarm Port:"},
			NumberEdit{
				AssignTo: &sd.ipfsSwarmPortEdit,
				Value:    float64(config.IPFSSwarmPort),
				MinValue: portMinValue,
				MaxValue: portMaxValue,
				Decimals: 0,
			},
			Label{Text: fmt.Sprintf("(default: %d)", winservice.DefaultIPFSSwarmPort), TextColor: walk.RGB(128, 128, 128)},

			Label{Text: "PinShare API Port:"},
			NumberEdit{
				AssignTo: &sd.pinshareAPIPortEdit,
				Value:    float64(config.PinShareAPIPort),
				MinValue: portMinValue,
				MaxValue: portMaxValue,
				Decimals: 0,
			},
			Label{Text: fmt.Sprintf("(default: %d)", winservice.DefaultPinShareAPIPort), TextColor: walk.RGB(128, 128, 128)},

			Label{Text: "PinShare P2P Port:"},
			NumberEdit{
				AssignTo: &sd.pinshareP2PPortEdit,
				Value:    float64(config.PinShareP2PPort),
				MinValue: portMinValue,
				MaxValue: portMaxValue,
				Decimals: 0,
			},
			Label{Text: fmt.Sprintf("(default: %d)", winservice.DefaultPinShareP2PPort), TextColor: walk.RGB(128, 128, 128)},

			Label{Text: "UI Port:"},
			NumberEdit{
				AssignTo: &sd.uiPortEdit,
				Value:    float64(config.UIPort),
				MinValue: portMinValue,
				MaxValue: portMaxValue,
				Decimals: 0,
			},
			Label{Text: fmt.Sprintf("(default: %d)", winservice.DefaultUIPort), TextColor: walk.RGB(128, 128, 128)},

			// Warning label spanning all columns
			VSpacer{Size: 10},
			VSpacer{Size: 10},
			VSpacer{Size: 10},
			Label{
				Text:       "Note: Changing ports requires a service restart.",
				TextColor:  walk.RGB(255, 140, 0),
				ColumnSpan: 3,
			},
		},
	}
}

// createOrganizationTab creates the Organization tab page
func (sd *SettingsDialog) createOrganizationTab(config *FullConfig) TabPage {
	return TabPage{
		Title:  "Organization",
		Layout: Grid{Columns: 2, Spacing: 10},
		Children: []Widget{
			Label{Text: "Organization Name:"},
			LineEdit{
				AssignTo: &sd.orgNameEdit,
				Text:     config.OrgName,
			},

			Label{Text: "Group Name:"},
			LineEdit{
				AssignTo: &sd.groupNameEdit,
				Text:     config.GroupName,
			},

			VSpacer{Size: 10},
			VSpacer{Size: 10},

			Label{
				Text:       "Organization and Group form the gossip topic for peer discovery.",
				ColumnSpan: 2,
			},
			Label{
				Text:       "All PinShare nodes with the same Organization and Group will",
				ColumnSpan: 2,
			},
			Label{
				Text:       "automatically discover and connect to each other.",
				ColumnSpan: 2,
			},
		},
	}
}

// createFeaturesTab creates the Features tab page
func (sd *SettingsDialog) createFeaturesTab(config *FullConfig) TabPage {
	return TabPage{
		Title:  "Features",
		Layout: VBox{Spacing: 10, MarginsZero: false},
		Children: []Widget{
			CheckBox{
				AssignTo: &sd.skipVTCheckbox,
				Text:     "Skip VirusTotal scanning",
				Checked:  config.SkipVirusTotal,
			},
			Label{
				Text:      "  Disable virus scanning for uploaded files",
				TextColor: walk.RGB(128, 128, 128),
			},

			VSpacer{Size: 5},

			CheckBox{
				AssignTo: &sd.enableCacheCheckbox,
				Text:     "Enable file caching",
				Checked:  config.EnableCache,
			},
			Label{
				Text:      "  Cache downloaded files locally for faster access",
				TextColor: walk.RGB(128, 128, 128),
			},

			VSpacer{Size: 5},

			CheckBox{
				AssignTo: &sd.archiveNodeCheckbox,
				Text:     "Archive node mode",
				Checked:  config.ArchiveNode,
			},
			Label{
				Text:      "  Pin all content shared by network peers (requires significant disk space)",
				TextColor: walk.RGB(128, 128, 128),
			},

			VSpacer{},
		},
	}
}

// createSecurityTab creates the Security tab page
func (sd *SettingsDialog) createSecurityTab(config *FullConfig, logLevelIndex int) TabPage {
	return TabPage{
		Title:  "Security",
		Layout: Grid{Columns: 2, Spacing: 10},
		Children: []Widget{
			Label{Text: "VirusTotal API Token:"},
			LineEdit{
				AssignTo:     &sd.vtTokenEdit,
				Text:         config.VirusTotalToken,
				PasswordMode: true,
			},
			Label{},
			Label{
				Text:      "Get a free API key from virustotal.com",
				TextColor: walk.RGB(128, 128, 128),
			},

			VSpacer{Size: 10},
			VSpacer{Size: 10},

			Label{Text: "Log Level:"},
			ComboBox{
				AssignTo:     &sd.logLevelCombo,
				Model:        logLevels,
				CurrentIndex: logLevelIndex,
			},
			Label{},
			Label{
				Text:      "debug = verbose, info = normal, warn/error = minimal",
				TextColor: walk.RGB(128, 128, 128),
			},
		},
	}
}

// createInfoTab creates the Info tab page (read-only)
func (sd *SettingsDialog) createInfoTab(config *FullConfig, configPath string) TabPage {
	return TabPage{
		Title:  "Info",
		Layout: Grid{Columns: 2, Spacing: 10},
		Children: []Widget{
			Label{Text: "Install Directory:"},
			LineEdit{
				Text:     config.InstallDirectory,
				ReadOnly: true,
			},

			Label{Text: "Data Directory:"},
			LineEdit{
				Text:     config.DataDirectory,
				ReadOnly: true,
			},

			Label{Text: "Config File:"},
			LineEdit{
				Text:     configPath,
				ReadOnly: true,
			},

			VSpacer{Size: 10},
			VSpacer{Size: 10},

			Label{
				Text:       "These paths are set during installation and cannot be changed here.",
				TextColor:  walk.RGB(128, 128, 128),
				ColumnSpan: 2,
			},
		},
	}
}

// createButtonsBar creates the Save/Cancel buttons composite
func (sd *SettingsDialog) createButtonsBar(saved *bool) Composite {
	return Composite{
		Layout: HBox{},
		Children: []Widget{
			HSpacer{},
			PushButton{
				AssignTo: &sd.saveButton,
				Text:     "Save",
				OnClicked: func() {
					if err := sd.validate(); err != nil {
						walk.MsgBox(sd.dlg, "Validation Error", err.Error(), walk.MsgBoxIconWarning)
						return
					}

					if err := sd.saveConfig(); err != nil {
						walk.MsgBox(sd.dlg, "Error", fmt.Sprintf("Failed to save settings: %v", err), walk.MsgBoxIconError)
						return
					}

					*saved = true
					sd.dlg.Accept()
				},
			},
			PushButton{
				Text: "Cancel",
				OnClicked: func() {
					sd.dlg.Cancel()
				},
			},
		},
	}
}

// validate checks that all fields have valid values
func (sd *SettingsDialog) validate() error {
	// Validate ports
	ports := map[string]float64{
		"IPFS API Port":     sd.ipfsAPIPortEdit.Value(),
		"IPFS Gateway Port": sd.ipfsGatewayPortEdit.Value(),
		"IPFS Swarm Port":   sd.ipfsSwarmPortEdit.Value(),
		"PinShare API Port": sd.pinshareAPIPortEdit.Value(),
		"PinShare P2P Port": sd.pinshareP2PPortEdit.Value(),
		"UI Port":           sd.uiPortEdit.Value(),
	}

	for name, port := range ports {
		if port < 1 || port > 65535 {
			return fmt.Errorf("%s must be between 1 and 65535", name)
		}
	}

	// Validate org/group names
	if sd.orgNameEdit.Text() == "" {
		return fmt.Errorf("Organization Name cannot be empty")
	}
	if sd.groupNameEdit.Text() == "" {
		return fmt.Errorf("Group Name cannot be empty")
	}

	return nil
}

// showNativeSettingsDialog shows the native Windows settings dialog.
// Returns true if settings were changed and saved.
func showNativeSettingsDialog() (bool, error) {
	config, configPath, err := loadFullConfig()
	if err != nil {
		return false, err
	}

	sd := &SettingsDialog{
		config:     config,
		configPath: configPath,
	}

	var saved bool

	logLevelIndex := 1 // default to "info"
	for i, level := range logLevels {
		if config.LogLevel == level {
			logLevelIndex = i
			break
		}
	}

	_, err = Dialog{
		AssignTo: &sd.dlg,
		Title:    "PinShare Settings",
		MinSize:  Size{Width: dialogMinWidth, Height: dialogMinHeight},
		Size:     Size{Width: dialogWidth, Height: dialogHeight},
		Layout:   VBox{},
		Children: []Widget{
			TabWidget{
				AssignTo: &sd.tabWidget,
				Pages: []TabPage{
					sd.createNetworkPortsTab(config),
					sd.createOrganizationTab(config),
					sd.createFeaturesTab(config),
					sd.createSecurityTab(config, logLevelIndex),
					sd.createInfoTab(config, configPath),
				},
			},
			sd.createButtonsBar(&saved),
		},
	}.Run(nil)

	if err != nil {
		return false, err
	}

	return saved, nil
}

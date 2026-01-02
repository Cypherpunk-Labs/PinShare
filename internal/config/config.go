package config

import (
	"os"
	"strconv"
	"time"

	"pinshare/internal/types"
)

// Environment variable names for configuration
const (
	// Organization settings
	EnvOrgName   = "PS_ORGNAME"
	EnvGroupName = "PS_GROUPNAME"

	// Path settings
	EnvUploadFolder    = "PS_UPLOAD_FOLDER"
	EnvCacheFolder     = "PS_CACHE_FOLDER"
	EnvRejectFolder    = "PS_REJECT_FOLDER"
	EnvMetadataFile    = "PS_METADATA_FILE"
	EnvIdentityKeyFile = "PS_IDENTITY_KEY_FILE"

	// Network settings
	EnvLibp2pPort = "PS_LIBP2P_PORT"

	// Feature flags
	EnvFFArchiveNode             = "PS_FF_ARCHIVE_NODE"
	EnvFFCache                   = "PS_FF_CACHE"
	EnvFFMoveUpload              = "PS_FF_MOVE_UPLOAD"
	EnvFFSendFileVT              = "PS_FF_SENDFILE_VT"
	EnvFFSkipVT                  = "PS_FF_SKIP_VT"
	EnvFFIgnoreUploadsInMetadata = "PS_FF_IGNORE_UPLOADS_IN_METADATA"
	EnvFFP2Pcircuit              = "PS_FF_P2PCIRCUIT"
	EnvFFTransportWS             = "PS_FF_TRNSPT_WS"
	EnvFFTransportTCP            = "PS_FF_TRNSPT_TCP"
	EnvFFTransportQUIC           = "PS_FF_TRNSPT_QUIC"
	EnvFFTransportWEBRTC         = "PS_FF_TRNSPT_WEBRTC"
)

// Default values for configuration
const (
	defaultUploadFolder     = "./upload"
	defaultCacheFolder      = "./cache"
	defaultRejectFolder     = "./rejected"
	defaultMetaDataFile     = "metadata.json"
	defaultIdentityKeyFile  = "identity.key"
	defaultLibp2pPort       = 50001
	defaultWatchInterval    = 2 * time.Minute
	defaultOrgName          = "Cypherpunk"
	defaultGroupName        = "TestLab"
	defaultMetadataTopicID  = "/metadata-sync/1.0.0"
	defaultFilteringTopicID = "/filtering-sync/1.0.0"
)

// Default values for Feature Flags
const (
	defaultFFArchiveNode             = false
	defaultFFCache                   = false
	defaultFFMoveUpload              = false
	defaultFFSendFileVT              = false
	defaultFFSkipVT                  = false
	defaultFFIgnoreUploadsInMetadata = true
	defaultFFP2Pcircuit              = true
	defaultFFTransportWS             = true
	defaultFFTransportTCP            = true
	defaultFFTransportQUIC           = true
	defaultFFTransportWEBRTC         = true
)

// AppConfig holds all configuration for the application.
type AppConfig struct {
	Version                   string
	SecurityCapability        types.SecurityCapability
	UploadFolder              string
	CacheFolder               string
	RejectFolder              string
	MetaDataFile              string
	IdentityKeyFile           string
	Libp2pPort                int
	WatchInterval             time.Duration
	OrgName                   string
	GroupName                 string
	MetadataTopicID           string
	FilteringTopicID          string
	FFArchiveNode             bool
	FFCache                   bool
	FFMoveUpload              bool
	FFSendFileVT              bool
	FFSkipVT                  bool
	FFIgnoreUploadsInMetadata bool
	FFP2Pcircuit              bool
	FFTransportWS             bool
	FFTransportTCP            bool
	FFTransportQUIC           bool
	FFTransportWEBRTC         bool
}

// LoadConfig loads configuration from environment variables, falling back to defaults.
// It returns a populated AppConfig struct. Errors are returned if environment
// variables are set but have invalid formats.
func LoadConfig() (*AppConfig, error) {
	conf := &AppConfig{
		Version:                   "dev0.1.3",
		SecurityCapability:        types.SecurityCapabilityNone,
		UploadFolder:              defaultUploadFolder,
		CacheFolder:               defaultCacheFolder,
		RejectFolder:              defaultRejectFolder,
		MetaDataFile:              defaultMetaDataFile,
		IdentityKeyFile:           defaultIdentityKeyFile,
		Libp2pPort:                defaultLibp2pPort,
		WatchInterval:             defaultWatchInterval,
		OrgName:                   defaultOrgName,
		GroupName:                 defaultGroupName,
		MetadataTopicID:           defaultMetadataTopicID,
		FilteringTopicID:          defaultFilteringTopicID,
		FFArchiveNode:             defaultFFArchiveNode,
		FFCache:                   defaultFFCache,
		FFMoveUpload:              defaultFFMoveUpload,
		FFSendFileVT:              defaultFFSendFileVT,
		FFSkipVT:                  defaultFFSkipVT,
		FFIgnoreUploadsInMetadata: defaultFFIgnoreUploadsInMetadata,
		FFP2Pcircuit:              defaultFFP2Pcircuit,
		FFTransportWS:             defaultFFTransportWS,
		FFTransportTCP:            defaultFFTransportTCP,
		FFTransportQUIC:           defaultFFTransportQUIC,
		FFTransportWEBRTC:         defaultFFTransportWEBRTC,
	}

	// Helper function to parse boolean environment variables
	parseBoolEnv := func(key string, target *bool) error {
		if val, ok := os.LookupEnv(key); ok {
			b, err := strconv.ParseBool(val)
			if err != nil {
				return err
			}
			*target = b
		}
		return nil
	}

	// Helper function to parse string environment variables
	parseStringEnv := func(key string, target *string) error {
		if val, ok := os.LookupEnv(key); ok {
			*target = val
		}
		return nil
	}

	// Helper function to parse int environment variables
	parseIntEnv := func(key string, target *int) error {
		if val, ok := os.LookupEnv(key); ok {
			b, err := strconv.ParseInt(val, 0, 64)
			if err != nil {
				return err
			}
			*target = int(b)
		}
		return nil
	}

	// Load organization and group names
	if err := parseStringEnv(EnvOrgName, &conf.OrgName); err != nil {
		return nil, err
	}
	if err := parseStringEnv(EnvGroupName, &conf.GroupName); err != nil {
		return nil, err
	}
	conf.MetadataTopicID = "/" + conf.OrgName + "/" + conf.GroupName + conf.MetadataTopicID
	conf.FilteringTopicID = "/" + conf.OrgName + "/" + conf.GroupName + conf.FilteringTopicID

	// Environment variable config overrides
	if err := parseStringEnv(EnvUploadFolder, &conf.UploadFolder); err != nil {
		return nil, err
	}
	if err := parseStringEnv(EnvCacheFolder, &conf.CacheFolder); err != nil {
		return nil, err
	}
	if err := parseStringEnv(EnvRejectFolder, &conf.RejectFolder); err != nil {
		return nil, err
	}
	if err := parseStringEnv(EnvMetadataFile, &conf.MetaDataFile); err != nil {
		return nil, err
	}
	if err := parseStringEnv(EnvIdentityKeyFile, &conf.IdentityKeyFile); err != nil {
		return nil, err
	}

	if err := parseIntEnv(EnvLibp2pPort, &conf.Libp2pPort); err != nil {
		return nil, err
	}

	// Load feature flags
	if err := parseBoolEnv(EnvFFArchiveNode, &conf.FFArchiveNode); err != nil {
		return nil, err
	}
	if err := parseBoolEnv(EnvFFCache, &conf.FFCache); err != nil {
		return nil, err
	}
	if err := parseBoolEnv(EnvFFMoveUpload, &conf.FFMoveUpload); err != nil {
		return nil, err
	}
	if err := parseBoolEnv(EnvFFSendFileVT, &conf.FFSendFileVT); err != nil {
		return nil, err
	}
	if err := parseBoolEnv(EnvFFSkipVT, &conf.FFSkipVT); err != nil {
		return nil, err
	}
	if err := parseBoolEnv(EnvFFIgnoreUploadsInMetadata, &conf.FFIgnoreUploadsInMetadata); err != nil {
		return nil, err
	}
	if err := parseBoolEnv(EnvFFP2Pcircuit, &conf.FFP2Pcircuit); err != nil {
		return nil, err
	}
	if err := parseBoolEnv(EnvFFTransportWS, &conf.FFTransportWS); err != nil {
		return nil, err
	}
	if err := parseBoolEnv(EnvFFTransportTCP, &conf.FFTransportTCP); err != nil {
		return nil, err
	}
	if err := parseBoolEnv(EnvFFTransportQUIC, &conf.FFTransportQUIC); err != nil {
		return nil, err
	}
	if err := parseBoolEnv(EnvFFTransportWEBRTC, &conf.FFTransportWEBRTC); err != nil {
		return nil, err
	}

	return conf, nil
}

package config

import (
	"os"
	"strconv"
	"time"
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
	defaultFF                        = false // ENVVAR NAME
	defaultFFArchiveNode             = false // PS_FF_ARCHIVE_NODE
	defaultFFCache                   = false // PS_FF_CACHE
	defaultFFMoveUpload              = false // PS_FF_MOVE_UPLOAD
	defaultFFSendFileVT              = false // PS_FF_SENDFILE_VT
	defaultFFSkipVT                  = false // PS_FF_SKIP_VT
	defaultFFIgnoreUploadsInMetadata = true  // PS_FF_IGNORE_UPLOADS_IN_METADATA
)

// AppConfig holds all configuration for the application.
type AppConfig struct {
	SecurityCapability        int
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
}

// LoadConfig loads configuration from environment variables, falling back to defaults.
// It returns a populated AppConfig struct. Errors are returned if environment
// variables are set but have invalid formats.
func LoadConfig() (*AppConfig, error) {
	conf := &AppConfig{
		SecurityCapability:        0,
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

	//TODO: Loadin the Org/Group names
	if err := parseStringEnv("PS_ORGNAME", &conf.OrgName); err != nil {
		return nil, err
	}
	if err := parseStringEnv("PS_GROUPNAME", &conf.GroupName); err != nil {
		return nil, err
	}
	conf.MetadataTopicID = "/" + conf.OrgName + "/" + conf.GroupName + conf.MetadataTopicID
	conf.FilteringTopicID = "/" + conf.OrgName + "/" + conf.GroupName + conf.FilteringTopicID

	if err := parseIntEnv("PS_LIBP2P_PORT", &conf.Libp2pPort); err != nil {
		return nil, err
	}
	if err := parseBoolEnv("PS_FF_MOVE_UPLOAD", &conf.FFMoveUpload); err != nil {
		return nil, err
	}
	if err := parseBoolEnv("PS_FF_SENDFILE_VT", &conf.FFSendFileVT); err != nil {
		return nil, err
	}
	if err := parseBoolEnv("PS_FF_SKIP_VT", &conf.FFSkipVT); err != nil {
		return nil, err
	}
	if err := parseBoolEnv("PS_FF_IGNORE_UPLOADS_IN_METADATA", &conf.FFIgnoreUploadsInMetadata); err != nil {
		return nil, err
	}

	return conf, nil
}

package config

import (
	"os"
	"strconv"
	"time"
)

// Default values for configuration
const (
	defaultUploadFolder    = "./upload"
	defaultCacheFolder     = "./cache"
	defaultRejectFolder    = "./rejected"
	defaultMetaDataFile    = "metadata.json"
	defaultIdentityKeyFile = "identity.key"
	defaultLibp2pPort      = 50001
	defaultWatchInterval   = 2 * time.Minute
	defaultOrgName         = "Cypherpunk"
	defaultGroupName       = "TestLab"
)

// Default values for Feature Flags
const (
	defaultFF                        = false // ENVVAR NAME
	defaultFFMoveUpload              = false // PS_FF_MOVE_UPLOAD
	defaultFFSendFileVT              = false // PS_FF_SENDFILE_VT
	defaultFFSkipVT                  = false // PS_FF_SKIP_VT
	defaultFFIgnoreUploadsInMetadata = true  // PS_FF_IGNORE_UPLOADS_IN_METADATA
)

// AppConfig holds all configuration for the application.
type AppConfig struct {
	UploadFolder              string
	CacheFolder               string
	RejectFolder              string
	MetaDataFile              string
	IdentityKeyFile           string
	Libp2pPort                int
	WatchInterval             time.Duration
	OrgName                   string
	GroupName                 string
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
		UploadFolder:              defaultUploadFolder,
		CacheFolder:               defaultCacheFolder,
		RejectFolder:              defaultRejectFolder,
		MetaDataFile:              defaultMetaDataFile,
		IdentityKeyFile:           defaultIdentityKeyFile,
		Libp2pPort:                defaultLibp2pPort,
		WatchInterval:             defaultWatchInterval,
		OrgName:                   defaultOrgName,
		GroupName:                 defaultGroupName,
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

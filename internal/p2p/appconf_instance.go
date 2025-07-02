package p2p

import "pinshare/internal/config"

var appconfInstance *config.AppConfig

func SetAppConfig(config *config.AppConfig) {
	appconfInstance = config
}

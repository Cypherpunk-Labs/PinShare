package config

import "time"

// Configs
const UploadFolder = "./upload"
const CacheFolder = "./cache"
const RejectFolder = "./rejected"
const DataFile = "metadata.json"
const IdentityKeyFile = "identity.key"

const Libp2pPort = 50001
const WatchInterval = 2 * time.Minute // Interval to scan the folder

// Feature Flags
const FF_move_upload = false
const FF_sendfile_vt = false
const FF_skip_vt = true
const FF_ignore_uploads_in_metadata = true

// TODO: Load and save config to file on start/exit.

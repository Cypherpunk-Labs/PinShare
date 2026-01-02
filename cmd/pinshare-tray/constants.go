package main

// Application identity
const (
	appName    = "PinShare"
	appTooltip = "PinShare - Decentralized IPFS Pinning"
)

// Windows MessageBox flags and return values
const (
	MB_OK              = 0x00000000
	MB_YESNO           = 0x00000004
	MB_YESNOCANCEL     = 0x00000003
	MB_ICONINFORMATION = 0x00000040
	MB_ICONERROR       = 0x00000010
	MB_ICONWARNING     = 0x00000030
	MB_ICONQUESTION    = 0x00000020

	IDYES    = 6
	IDNO     = 7
	IDCANCEL = 2
)

// Directory names within the data directory
const (
	dirIPFS     = "ipfs"
	dirPinShare = "pinshare"
	dirUpload   = "upload"
	dirCache    = "cache"
	dirRejected = "rejected"
	dirLogs     = "logs"
)

// File names
const (
	fileConfig  = "config.json"
	fileSession = "session.json"
)

// Environment variable names
const (
	envLocalAppData = "LOCALAPPDATA"
	envUserProfile  = "USERPROFILE"
	envProgramData  = "PROGRAMDATA"
	envUsername     = "USERNAME"
)

// Default paths when environment variables are not available
const (
	defaultLocalAppDataPath = `C:\Users\Default\AppData\Local`
	defaultProgramDataPath  = `C:\ProgramData`
)

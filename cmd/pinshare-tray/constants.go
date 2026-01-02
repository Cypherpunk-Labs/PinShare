package main

import "pinshare/internal/winservice"

// Tray-specific constants (aliases to shared constants for convenience)
const (
	appName    = winservice.AppName
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

// Aliases to shared constants for package-level convenience
const (
	dirIPFS     = winservice.DirIPFS
	dirPinShare = winservice.DirPinShare
	dirUpload   = winservice.DirUpload
	dirCache    = winservice.DirCache
	dirRejected = winservice.DirRejected
	dirLogs     = winservice.DirLogs
)

const (
	fileConfig  = winservice.FileConfig
	fileSession = winservice.FileSession
)

const (
	envLocalAppData = winservice.EnvLocalAppData
	envUserProfile  = winservice.EnvUserProfile
	envProgramData  = winservice.EnvProgramData
	envUsername     = winservice.EnvUsername
)

const (
	defaultLocalAppDataPath = winservice.DefaultLocalAppDataPath
	defaultProgramDataPath  = winservice.DefaultProgramDataPath
)

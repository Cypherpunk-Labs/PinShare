# PinShare Windows Architecture

This document describes the architecture of PinShare when deployed on Windows.

## Process Hierarchy

```
┌─────────────────────────────────────────────────────────────────────┐
│                     Windows Service Manager                          │
│                    (runs at system startup)                          │
└───────────────────────────┬─────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    pinsharesvc.exe (Windows Service)                 │
│                    "PinShareService"                                 │
│                                                                      │
│  • Runs as SYSTEM account (no user login required)                  │
│  • Manages child processes (keeps them alive)                       │
│  • Monitors health & auto-restarts crashed processes                │
│  • Embedded UI server on port 8888                                  │
│                                                                      │
│  ┌─────────────────────┐    ┌─────────────────────┐                │
│  │   ipfs.exe          │    │   pinshare.exe      │                │
│  │   (child process)   │    │   (child process)   │                │
│  │                     │    │                     │                │
│  │ • IPFS daemon       │    │ • libp2p host       │                │
│  │ • Port 5001 (API)   │◄───│ • PubSub messaging  │                │
│  │ • Port 4001 (swarm) │    │ • File watcher      │                │
│  │ • Port 8080 (gw)    │    │ • API on port 9090  │                │
│  └─────────────────────┘    └─────────────────────┘                │
└─────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────┐
│                    pinshare-tray.exe (User Process)                  │
│                    (runs at user login via Startup folder)           │
│                                                                      │
│  • Runs in USER context (per-user, after login)                     │
│  • System tray icon for user interaction                            │
│  • NOT managed by service - completely independent                  │
│  • Talks to service via HTTP APIs                                   │
└─────────────────────────────────────────────────────────────────────┘
```

## Components

### pinsharesvc.exe (Windows Service Wrapper)

The main Windows service that orchestrates all PinShare components.

**Responsibilities:**
- Registers as a Windows Service ("PinShareService")
- Starts and monitors IPFS daemon
- Starts and monitors PinShare backend
- Runs embedded UI server (serves React UI)
- Health checking with automatic restart on failure
- Graceful shutdown of all components

**Ports:**
- 8888: UI Server (serves React frontend, proxies API requests)

**Source:** `cmd/pinsharesvc/`

### pinshare.exe (Main Daemon)

The core PinShare application with libp2p networking.

**Responsibilities:**
- libp2p host for P2P communication
- PubSub for metadata synchronization
- File watcher for upload folder
- REST API for external integrations
- Connects to IPFS daemon for storage

**Ports:**
- 9090: REST API
- 50001: libp2p P2P port

**Source:** `internal/` (main application code)

### ipfs.exe (IPFS Kubo Daemon)

Standard IPFS daemon for content-addressed storage.

**Ports:**
- 5001: IPFS API
- 4001: IPFS Swarm (P2P)
- 8080: IPFS Gateway

**Source:** Downloaded from https://dist.ipfs.tech/kubo/

### pinshare-tray.exe (System Tray Application)

User-facing system tray application for easy interaction.

**Responsibilities:**
- System tray icon with context menu
- Open web UI in browser
- Start/Stop/Restart service
- Show service status

**Note:** This runs independently of the service, launched via Windows Startup folder.

**Source:** `cmd/pinshare-tray/`

## Data Flow

```
User clicks tray icon
        │
        ▼
pinshare-tray.exe ──HTTP──► pinsharesvc.exe (port 8888)
                                   │
                                   ├──proxy──► pinshare.exe API (port 9090)
                                   │                  │
                                   │                  ▼
                                   │           libp2p network
                                   │                  │
                                   └──────────► ipfs.exe (port 5001)
                                                      │
                                                      ▼
                                               IPFS network
```

## Installed Files

```
C:\Program Files\PinShare\
├── pinsharesvc.exe    # Windows service wrapper
├── pinshare.exe       # Main daemon (managed by service)
├── pinshare-tray.exe  # User tray app (independent)
├── ipfs.exe           # IPFS daemon (managed by service)
└── ui\                # React web UI (served by service)
    ├── index.html
    ├── assets\
    └── ...

C:\ProgramData\PinShare\
├── config.json        # Configuration file
├── ipfs\              # IPFS repository
│   ├── config
│   ├── datastore\
│   └── ...
├── pinshare\          # PinShare data
│   ├── identity.key   # libp2p identity
│   ├── metadata.json  # File metadata store
│   └── pinshare.db    # SQLite database
├── upload\            # Watch folder for new files
├── cache\             # Downloaded/processed files
├── rejected\          # Files that failed security scan
└── logs\              # Log files
    ├── service.log
    ├── ipfs.log
    └── pinshare.log
```

## Registry Configuration

Configuration is stored in Windows Registry at:
```
HKEY_LOCAL_MACHINE\SOFTWARE\PinShare\
```

| Key | Type | Description |
|-----|------|-------------|
| InstallDirectory | REG_SZ | Installation path |
| DataDirectory | REG_SZ | Data directory path |
| IPFSAPIPort | REG_DWORD | IPFS API port (default: 5001) |
| IPFSGatewayPort | REG_DWORD | IPFS Gateway port (default: 8080) |
| IPFSSwarmPort | REG_DWORD | IPFS Swarm port (default: 4001) |
| PinShareAPIPort | REG_DWORD | PinShare API port (default: 9090) |
| PinShareP2PPort | REG_DWORD | libp2p port (default: 50001) |
| UIPort | REG_DWORD | UI server port (default: 8888) |
| OrgName | REG_SZ | Organization name for topic |
| GroupName | REG_SZ | Group name for topic |
| SkipVirusTotal | REG_DWORD | Skip virus scanning (0/1) |
| EnableCache | REG_DWORD | Enable file caching (0/1) |
| ArchiveNode | REG_DWORD | Run as archive node (0/1) |

## Service Management

### Install Service
```batch
pinsharesvc.exe install
```

### Uninstall Service
```batch
pinsharesvc.exe uninstall
```

### Start/Stop Service
```batch
pinsharesvc.exe start
pinsharesvc.exe stop
pinsharesvc.exe restart
```

### Debug Mode (Console)
```batch
pinsharesvc.exe debug
```

### Using Windows Service Manager
```batch
net start PinShareService
net stop PinShareService
sc query PinShareService
```

## Health Monitoring

The service includes a health checker that:
- Checks IPFS health every 30 seconds via `http://localhost:5001/api/v0/version`
- Checks PinShare health every 30 seconds via `http://localhost:9090/api/health`
- Automatically restarts failed components (up to 3 times)
- Logs all health events to Windows Event Log

## Security Capabilities

PinShare supports multiple security scanning backends:

| Capability | Description | Requirements |
|------------|-------------|--------------|
| 0 | No scanning (fails startup) | - |
| 1 | P2P-Sec service | Port 36939 running |
| 2 | VirusTotal API | VT_TOKEN env var |
| 3 | ClamAV | clamscan in PATH |
| 4 | VirusTotal via browser | Chromium installed |

Set `SkipVirusTotal=1` in registry to bypass all scanning (for testing).

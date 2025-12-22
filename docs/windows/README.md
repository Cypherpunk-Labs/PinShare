# PinShare for Windows

Complete guide for installing and using PinShare on Windows.

## Table of Contents

- [Installation](#installation)
- [Getting Started](#getting-started)
- [Configuration](#configuration)
- [Using PinShare](#using-pinshare)
- [Troubleshooting](#troubleshooting)
- [Uninstallation](#uninstallation)
- [Advanced Topics](#advanced-topics)

## Installation

### System Requirements

- **Windows 10** or **Windows 11** (64-bit)
- **4 GB RAM** minimum (8 GB recommended)
- **10 GB free disk space** (more for IPFS storage)
- **Administrator privileges** for installation

### Installation Steps

1. **Download the installer**
   - Download `PinShare-Setup.msi` from the releases page
   - Or build from source (see [Building from Source](#building-from-source))

2. **Run the installer**
   - Double-click `PinShare-Setup.msi`
   - Click "Next" through the installation wizard
   - Accept the license agreement
   - Choose installation directory (default: `C:\Program Files\PinShare`)
   - Click "Install"

3. **Complete installation**
   - The installer will:
     - Install all required binaries
     - Create data directories
     - Install and start the Windows service
     - Add system tray application to startup

4. **First launch**
   - The PinShare service should start automatically
   - Look for the PinShare icon in your system tray (bottom-right corner)

## Getting Started

### First-Time Setup

When PinShare first starts:

1. The IPFS repository will be initialized automatically
2. PinShare will generate a unique identity key
3. You can start sharing files immediately

### Sharing Files

1. Copy or move files to the upload folder (default: `C:\ProgramData\PinShare\upload`)
2. Files are automatically:
   - Scanned for malware (if configured)
   - Added to IPFS
   - Metadata is shared with peers via libp2p PubSub

## Configuration

### Default Settings

PinShare uses these default settings:

| Setting | Default Value | Description |
|---------|---------------|-------------|
| API Port | 9090 | Backend API port (localhost only) |
| IPFS API | 5001 | IPFS daemon API port (localhost only) |
| IPFS Gateway | 8080 | IPFS HTTP gateway (localhost only) |
| IPFS Swarm | 4001 | IPFS P2P port (public) |
| libp2p Port | 50001 | PinShare P2P port (public) |

**Note:** The API ports (9090, 5001, 8080) are bound to localhost (127.0.0.1) by default and are not exposed to the network. They will be needed for the future web UI integration.

### Changing Configuration

Edit: `C:\ProgramData\PinShare\config.json`

```json
{
  "install_directory": "C:\\Program Files\\PinShare",
  "data_directory": "C:\\ProgramData\\PinShare",
  "ipfs_binary": "C:\\Program Files\\PinShare\\ipfs.exe",
  "pinshare_binary": "C:\\Program Files\\PinShare\\pinshare.exe",
  "ipfs_api_port": 5001,
  "ipfs_gateway_port": 8080,
  "ipfs_swarm_port": 4001,
  "pinshare_api_port": 9090,
  "pinshare_p2p_port": 50001,
  "ui_port": 8888,
  "org_name": "MyOrganization",
  "group_name": "MyGroup",
  "skip_virus_total": false,
  "enable_cache": true,
  "archive_node": false,
  "virus_total_token": "",
  "log_level": "info",
  "log_file_path": "C:\\ProgramData\\PinShare\\logs\\service.log"
}
```

Then restart the service:
```cmd
net stop PinShareService
net start PinShareService
```

### Data Directories

PinShare stores data in:

```
C:\ProgramData\PinShare\
├── config.json          # Configuration file
├── logs\
│   ├── service.log      # Service logs
│   ├── ipfs.log         # IPFS logs
│   └── pinshare.log     # Backend logs
├── ipfs\                # IPFS repository
├── pinshare\
│   ├── metadata.json    # File metadata
│   └── identity.key     # libp2p identity (keep secure!)
├── upload\              # Upload directory
├── cache\               # File cache
└── rejected\            # Rejected files
```

**Security Note:** The `identity.key` file contains the libp2p private key (protobuf-encoded). This key identifies your node on the network. Keep this file secure and backed up. On Windows, the `C:\ProgramData\PinShare` directory inherits SYSTEM/Administrators permissions by default, restricting access to privileged users.

## Using PinShare

### System Tray Application

The system tray application provides quick access:

**Menu Options:**
- **Status** - Shows service status
- **Start/Stop/Restart Service** - Control the service
- **Settings** - Configure PinShare options
- **View Logs** - Opens log directory
- **Exit** - Closes tray app (service continues running)

### Service Management

#### Using Services Manager (GUI)

1. Press `Win + R`, type `services.msc`, press Enter
2. Find "PinShare Service"
3. Right-click → Start/Stop/Restart

#### Using Command Line

```cmd
# Start service
net start PinShareService

# Stop service
net stop PinShareService

# Restart service
net stop PinShareService && net start PinShareService

# Check status
sc query PinShareService
```

#### Using Service Wrapper

```cmd
cd "C:\Program Files\PinShare"

# Start
pinsharesvc.exe start

# Stop
pinsharesvc.exe stop

# Restart
pinsharesvc.exe restart

# Debug mode (console)
pinsharesvc.exe debug
```

### Firewall Configuration

- Port **4001** - IPFS swarm (P2P file sharing)
- Port **50001** - PinShare libp2p (peer discovery)

To add firewall rules (run in Git Bash as Administrator):

```bash
# Add firewall rules using PowerShell
powershell -Command "New-NetFirewallRule -DisplayName 'IPFS Swarm' -Direction Inbound -Protocol TCP -LocalPort 4001 -Action Allow"
powershell -Command "New-NetFirewallRule -DisplayName 'PinShare P2P' -Direction Inbound -Protocol TCP -LocalPort 50001 -Action Allow"
```

**Testing port reachability** (from another machine or using online port checkers):

```bash
# Test if ports are accessible
nc -zv your-public-ip 4001
nc -zv your-public-ip 50001
```

### Security Scanning

PinShare supports multiple virus scanning options (in priority order):

1. **P2P-Sec Service** (port 36939) - Preferred
2. **VirusTotal API** - Requires API token
3. **ClamAV** - Local scanning

To configure VirusTotal:

1. Get API token from https://www.virustotal.com/
2. Add to config.json: `"virus_total_token": "your_api_token"`
3. Restart service

## Troubleshooting

### Service Won't Start

**Check Event Viewer:**
1. Press `Win + R`, type `eventvwr.msc`
2. Go to: Windows Logs → Application
3. Look for PinShare errors

**Common issues:**

1. **Port already in use**
   - Check if another app is using ports 9090, 5001, 4001
   - Change ports in configuration

2. **IPFS failed to initialize**
   - Check logs: `C:\ProgramData\PinShare\logs\ipfs.log`
   - Delete IPFS repo: `C:\ProgramData\PinShare\ipfs`
   - Restart service (will re-initialize)

3. **Permission denied**
   - Ensure service has write access to `C:\ProgramData\PinShare`
   - Check antivirus isn't blocking executables

### High CPU/Memory Usage

**IPFS repository cleanup:**

```cmd
cd "C:\Program Files\PinShare"

# Run IPFS garbage collection
ipfs.exe --repo-dir="C:\ProgramData\PinShare\ipfs" repo gc
```

**Limit IPFS resource usage:**

1. Edit: `C:\ProgramData\PinShare\ipfs\config`
2. Modify `Swarm.ConnMgr`:
   ```json
   "ConnMgr": {
     "HighWater": 300,
     "LowWater": 150
   }
   ```

### Can't Connect to Peers

1. **Check firewall** - Ensure ports 4001 and 50001 are open
2. **Check NAT** - PinShare uses relay for NAT traversal
3. **View peer status** (from Git Bash):
   ```bash
   curl http://localhost:9090/api/status
   ```

### Logs and Debugging

**View logs (Git Bash):**

```bash
# Service log
cat "/c/ProgramData/PinShare/logs/service.log"

# IPFS log
cat "/c/ProgramData/PinShare/logs/ipfs.log"

# PinShare log
cat "/c/ProgramData/PinShare/logs/pinshare.log"
```

**Enable debug mode (Git Bash):**

1. Stop the service: `net stop PinShareService`
2. Run in console mode:
   ```bash
   cd "/c/Program Files/PinShare"
   ./pinsharesvc.exe debug
   ```
3. Watch console output

**Tail logs (Git Bash):**

```bash
tail -f "/c/ProgramData/PinShare/logs/service.log"
```

## Uninstallation

### Using Control Panel

1. Open Settings → Apps → Installed apps
2. Find "PinShare"
3. Click "Uninstall"

### Using Installer

```cmd
msiexec /x PinShare-Setup.msi
```

### Manual Cleanup (if needed)

The uninstaller preserves data. To completely remove:

```cmd
# Remove program files
rmdir /s "C:\Program Files\PinShare"

# Remove data (WARNING: Deletes all pins and configuration)
rmdir /s "C:\ProgramData\PinShare"
```

## Advanced Topics

### Windows 11 Compatibility

PinShare is fully compatible with Windows 11. The installer includes enhanced service startup logic to handle Windows 11's stricter security policies.

For technical details about Windows 11 specific improvements, see [Windows 11 Compatibility Guide](WINDOWS11_COMPATIBILITY.md).

### Running Multiple Instances

To run multiple PinShare instances:

1. Install normally (first instance)
2. For additional instances:
   - Copy installation directory
   - Change all ports in configuration
   - Install as separate service with different name

### Backup and Restore

**Backup:**

```cmd
# Stop service
net stop PinShareService

# Backup data directory
xcopy "C:\ProgramData\PinShare" "D:\Backup\PinShare\" /E /I /H

# Restart service
net start PinShareService
```

**Restore:**

```cmd
# Stop service
net stop PinShareService

# Restore data
xcopy "D:\Backup\PinShare\" "C:\ProgramData\PinShare\" /E /I /H /Y

# Restart service
net start PinShareService
```

### Performance Tuning

**For archive nodes** (storing many files):
- Increase disk space for IPFS repo
- Disable automatic garbage collection
- Add to config: `"archive_node": true`

**For low-resource systems:**
- Reduce IPFS connection limits
- Disable caching: `"enable_cache": false`
- Enable VirusTotal skip: `"skip_virus_total": true`

### Network Configuration

**Public P2P Ports:**

For PinShare to work optimally with other peers, the following ports must be publicly accessible:

| Port | Protocol | Purpose |
|------|----------|---------|
| 4001 | TCP/UDP | IPFS Swarm (file sharing) |
| 50001 | TCP | PinShare libp2p (peer discovery) |

**Options for public access:**
- **UPnP** - Automatically opens ports if your router supports it
- **Port forwarding** - Manually configure your router to forward these ports
- **NAT traversal** - PinShare uses relay servers as fallback when direct connections fail

**Note:** The API ports (5001, 8080, 9090) are bound to localhost by default and should remain so for security.

## Building from Source

See [BUILD.md](BUILD.md) for complete build instructions.

Quick start (Git Bash):

```bash
# Prerequisites: Go 1.24+, Git Bash

# Clone repository
git clone https://github.com/Cypherpunk-Labs/PinShare.git
cd PinShare

# Build all components
./build-windows.sh
```

## Support

- **Issues**: https://github.com/Cypherpunk-Labs/PinShare/issues
- **Documentation**: https://github.com/Cypherpunk-Labs/PinShare/tree/main/docs
- **Logs**: `C:\ProgramData\PinShare\logs`

## License

PinShare is released under the MIT License. See LICENSE file for details.

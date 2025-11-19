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
   - Click "Open PinShare UI" to access the web interface

## Getting Started

### Accessing the UI

After installation, access PinShare through:

1. **System Tray**
   - Right-click the PinShare icon
   - Select "Open PinShare UI"

2. **Browser**
   - Navigate to: http://localhost:8888

3. **Start Menu**
   - Start Menu → PinShare → Open PinShare UI

### First-Time Setup

When you first open PinShare:

1. The IPFS repository will be initialized automatically
2. PinShare will generate a unique identity key
3. You can start uploading files immediately

### Uploading Files

1. Click "Upload" or drag files to the upload area
2. Files are automatically:
   - Scanned for malware (if configured)
   - Added to IPFS
   - Shared with peers via libp2p

## Configuration

### Default Settings

PinShare uses these default settings:

| Setting | Default Value | Description |
|---------|---------------|-------------|
| UI Port | 8888 | Web interface port |
| API Port | 9090 | Backend API port |
| IPFS API | 5001 | IPFS daemon API port |
| IPFS Gateway | 8080 | IPFS HTTP gateway |
| IPFS Swarm | 4001 | IPFS P2P port |
| libp2p Port | 50001 | PinShare P2P port |

### Changing Configuration

#### Method 1: Registry Editor (Advanced)

1. Press `Win + R`, type `regedit`, press Enter
2. Navigate to: `HKEY_LOCAL_MACHINE\SOFTWARE\PinShare`
3. Modify values:
   - `UIPort` - Change web UI port
   - `OrgName` - Your organization name
   - `GroupName` - Your group name
   - `SkipVirusTotal` - 1 to skip virus scanning
   - `EnableCache` - 1 to enable caching

4. Restart the service:
   ```cmd
   net stop PinShareService
   net start PinShareService
   ```

#### Method 2: Configuration File

Edit: `C:\ProgramData\PinShare\config.json`

```json
{
  "ui_port": 8888,
  "pinshare_api_port": 9090,
  "org_name": "MyOrganization",
  "group_name": "MyGroup",
  "skip_virus_total": true,
  "enable_cache": true
}
```

Then restart the service.

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
│   ├── pinshare.db      # SQLite database
│   ├── metadata.json    # File metadata
│   └── identity.key     # libp2p identity
├── upload\              # Upload directory
├── cache\               # File cache
└── rejected\            # Rejected files
```

## Using PinShare

### System Tray Application

The system tray application provides quick access:

**Menu Options:**
- **Open PinShare UI** - Opens web interface
- **Status** - Shows service status
- **Start/Stop/Restart Service** - Control the service
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

PinShare needs these ports open:

**Outbound** (usually allowed by default):
- All ports for IPFS swarm connections

**Inbound** (may need firewall rules):
- Port **4001** - IPFS swarm (P2P file sharing)
- Port **50001** - PinShare libp2p (peer discovery)

To add firewall rules:

```powershell
# Run as Administrator
New-NetFirewallRule -DisplayName "IPFS Swarm" -Direction Inbound -Protocol TCP -LocalPort 4001 -Action Allow
New-NetFirewallRule -DisplayName "PinShare P2P" -Direction Inbound -Protocol TCP -LocalPort 50001 -Action Allow
```

### Security Scanning

PinShare supports multiple virus scanning options (in priority order):

1. **P2P-Sec Service** (port 36939) - Preferred
2. **VirusTotal API** - Requires API token
3. **ClamAV** - Local scanning

To configure VirusTotal:

1. Get API token from https://www.virustotal.com/
2. Add to registry: `HKLM\SOFTWARE\PinShare\VirusTotalToken`
3. Restart service

## Troubleshooting

### Service Won't Start

**Check Event Viewer:**
1. Press `Win + R`, type `eventvwr.msc`
2. Go to: Windows Logs → Application
3. Look for PinShare errors

**Common issues:**

1. **Port already in use**
   - Check if another app is using ports 8888, 9090, 5001
   - Change ports in configuration

2. **IPFS failed to initialize**
   - Check logs: `C:\ProgramData\PinShare\logs\ipfs.log`
   - Delete IPFS repo: `C:\ProgramData\PinShare\ipfs`
   - Restart service (will re-initialize)

3. **Permission denied**
   - Ensure service has write access to `C:\ProgramData\PinShare`
   - Check antivirus isn't blocking executables

### UI Not Loading

1. **Check service status**
   ```cmd
   sc query PinShareService
   ```

2. **Verify UI server is running**
   - Open: http://localhost:8888
   - If connection refused, check `service.log`

3. **Check browser console**
   - Press F12 in browser
   - Look for JavaScript errors

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
3. **View peer status**:
   ```cmd
   curl http://localhost:9090/api/status
   ```

### Logs and Debugging

**View logs:**

```cmd
# Service log
type "C:\ProgramData\PinShare\logs\service.log"

# IPFS log
type "C:\ProgramData\PinShare\logs\ipfs.log"

# PinShare log
type "C:\ProgramData\PinShare\logs\pinshare.log"
```

**Enable debug mode:**

1. Stop the service
2. Run in console mode:
   ```cmd
   cd "C:\Program Files\PinShare"
   pinsharesvc.exe debug
   ```
3. Watch console output

**Tail logs in PowerShell:**

```powershell
Get-Content "C:\ProgramData\PinShare\logs\service.log" -Wait -Tail 50
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

# Remove registry entries
reg delete "HKLM\SOFTWARE\PinShare" /f
reg delete "HKCU\SOFTWARE\PinShare" /f
```

## Advanced Topics

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

**Static IP / Public Access:**

If you want to access PinShare from other computers:

1. **Change bind address** (advanced):
   - Modify service to bind to `0.0.0.0` instead of `localhost`
   - Add firewall rules for ports 8888, 9090
   - ⚠️ **Security risk** - Add authentication first!

2. **Use reverse proxy** (recommended):
   - Install nginx/Caddy
   - Proxy to `localhost:8888`
   - Add HTTPS and authentication

## Building from Source

See [BUILD.md](BUILD.md) for complete build instructions.

Quick start:

```cmd
# Install dependencies
# - Go 1.24+
# - Node.js 20+
# - MinGW-w64 (for CGO/SQLite)
# - WiX Toolset

# Clone repository
git clone https://github.com/Episk-pos/PinShare.git
cd PinShare

# Build all components
make -f Makefile.windows windows-all

# Build installer
cd installer
build.bat
```

## Support

- **Issues**: https://github.com/Episk-pos/PinShare/issues
- **Documentation**: https://github.com/Episk-pos/PinShare/docs
- **Logs**: `C:\ProgramData\PinShare\logs`

## License

PinShare is released under the MIT License. See LICENSE file for details.

## Windows Installer & Service Infrastructure

### Summary

This PR introduces a complete Windows installation and service management infrastructure for PinShare, including:

- **Windows Service Wrapper** (`pinsharesvc`) - Manages IPFS daemon and PinShare backend as a Windows service
- **System Tray Application** (`pinshare-tray`) - User-friendly tray icon for service status monitoring and control
- **WiX 6 Installer** - Professional MSI installer with configuration wizard
- **Build Infrastructure** - Batch scripts and PowerShell for building all Windows components

### Key Changes

#### New Components

- **`cmd/pinsharesvc/`** - Windows service that:
  - Manages IPFS daemon lifecycle
  - Runs PinShare backend API
  - Handles graceful shutdown and process cleanup
  - Loads configuration from registry/config files

- **`cmd/pinshare-tray/`** - System tray application that:
  - Displays real-time service status (Running/Stopped/Starting)
  - Shows IPFS and PinShare component health
  - Provides Start/Stop/Restart controls (with UAC elevation)
  - Uses Windows SCM API directly for status checks (no process spawning)
  - Temporarily disables "Open PinShare UI" menu item (commented for future re-enablement)

- **`installer/`** - WiX 6 MSI installer with:
  - Configuration wizard for upload directory and settings
  - Service installation and automatic startup
  - Proper uninstallation cleanup
  - GPLv3 license display

#### Build System

- `build-windows.bat` / `build-windows.ps1` - Automated build scripts
- `installer/build-wix6.bat` - WiX 6 installer compilation
- GitHub Actions workflow for CI/CD

#### Configuration & P2P Improvements

- Feature flags for relay and transport options
- IPv6 localhost publishing fix
- Expanded allowed file types (CAD formats)
- `/api/health` endpoint for service health checks

### Technical Notes

- Tray app uses `golang.org/x/sys/windows` SCM API with minimal permissions (`SC_MANAGER_CONNECT`, `SERVICE_QUERY_STATUS`) to avoid UI flickering from process spawning
- Service control operations use PowerShell UAC elevation (`-Verb RunAs`)
- Status polling occurs every 10 seconds via in-process Windows API calls

### Documentation

- `WINDOWS_SERVICE.md` - Service architecture documentation
- `docs/windows/` - Build, testing, and troubleshooting guides
- `INSTALLER-QUICKSTART.md` - Quick start for building installer

### Test Plan

- [ ] Build all components with `build-windows.bat`
- [ ] Install MSI on clean Windows machine
- [ ] Verify service starts automatically after install
- [ ] Verify tray app shows correct status
- [ ] Test Start/Stop/Restart from tray menu
- [ ] Verify clean uninstallation
- [ ] Test on Windows 10 and Windows 11

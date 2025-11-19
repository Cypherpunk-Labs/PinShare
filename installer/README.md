# PinShare Windows Installer

This directory contains the WiX Toolset configuration for building the PinShare Windows installer.

## Prerequisites

1. **WiX Toolset 3.x or 4.x**
   - Download from: https://wixtoolset.org/
   - Add WiX bin directory to PATH

2. **Visual C++ Redistributable** (for end users)
   - The installer should bundle this if CGO is used

## Building the Installer

### 1. Build all binaries first

```bash
# From repository root
make windows-all
```

This will create:
- `dist/windows/pinsharesvc.exe` - Windows service wrapper
- `dist/windows/pinshare.exe` - PinShare backend
- `dist/windows/pinshare-tray.exe` - System tray application
- `dist/windows/ipfs.exe` - IPFS Kubo daemon
- `dist/windows/ui/` - React UI static files

### 2. Run the installer build script

```cmd
cd installer
build.bat
```

This will:
1. Harvest UI files using WiX heat.exe
2. Compile WiX sources
3. Link to create MSI package
4. Output: `dist/PinShare-Setup.msi`

## Manual Build Steps

If you prefer to build manually:

```cmd
cd installer

# Harvest UI files
heat.exe dir "..\dist\windows\ui" -cg UIComponents -dr UIFolder -gg -g1 -sf -srd -var var.UISourceDir -out UIComponents.wxs

# Compile
candle.exe -ext WixUIExtension -ext WixUtilExtension -dUISourceDir="..\dist\windows\ui" Product.wxs UIComponents.wxs

# Link
light.exe -ext WixUIExtension -ext WixUtilExtension -out PinShare-Setup.msi Product.wixobj UIComponents.wixobj
```

## Installer Features

The installer will:

1. **Install binaries** to `C:\Program Files\PinShare\`
   - pinsharesvc.exe
   - pinshare.exe
   - pinshare-tray.exe
   - ipfs.exe
   - UI files

2. **Create data directory** at `C:\ProgramData\PinShare\`
   - logs/
   - ipfs/
   - pinshare/
   - upload/
   - cache/
   - rejected/

3. **Install Windows service** (PinShareService)
   - Set to start automatically
   - Configure recovery options

4. **Create registry entries** at `HKLM\SOFTWARE\PinShare`
   - Installation paths
   - Port configurations
   - Default settings

5. **Add to startup**
   - System tray application in user startup folder

6. **Create shortcuts** in Start Menu
   - Open PinShare UI
   - Uninstall PinShare

## Testing the Installer

1. **Install**
   ```cmd
   msiexec /i PinShare-Setup.msi
   ```

2. **Install with logging**
   ```cmd
   msiexec /i PinShare-Setup.msi /l*v install.log
   ```

3. **Uninstall**
   ```cmd
   msiexec /x PinShare-Setup.msi
   ```

## Customization

### Changing the UpgradeCode

Edit `Product.wxs`:
```xml
<?define UpgradeCode = "YOUR-GUID-HERE" ?>
```

Generate a new GUID:
```powershell
[guid]::NewGuid()
```

### Adding More Files

Use WiX heat.exe to harvest file lists, or manually add components to Product.wxs.

### Changing Default Ports

Edit the registry values in `Product.wxs`:
```xml
<RegistryValue Name="UIPort" Type="integer" Value="8888" />
```

## Code Signing (Optional)

To sign the installer:

```cmd
signtool sign /f certificate.pfx /p password /t http://timestamp.digicert.com PinShare-Setup.msi
```

## Troubleshooting

**Error: "candle.exe is not recognized"**
- Add WiX bin directory to PATH
- Default location: `C:\Program Files (x86)\WiX Toolset v3.x\bin`

**Error: "UI files not found"**
- Build the React UI first: `cd pinshare-ui && npm run build`

**Error: "IPFS binary not found"**
- Download from: https://dist.ipfs.tech/kubo/
- Extract ipfs.exe to `dist/windows/`

**Service fails to start after installation**
- Check Windows Event Viewer → Application logs
- Check `C:\ProgramData\PinShare\logs\service.log`
- Verify all binaries are present and not blocked by antivirus

## Architecture

The installer creates this structure:

```
C:\Program Files\PinShare\
├── pinsharesvc.exe       # Service wrapper
├── pinshare.exe          # Backend binary
├── pinshare-tray.exe     # Tray application
├── ipfs.exe              # IPFS daemon
├── icon.ico
└── ui\                   # React static files
    ├── index.html
    ├── assets\
    └── ...

C:\ProgramData\PinShare\
├── config.json
├── logs\
├── ipfs\                 # IPFS repository
├── pinshare\             # Database, metadata
├── upload\
├── cache\
└── rejected\
```

## License

Same license as PinShare (MIT)

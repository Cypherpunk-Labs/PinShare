# Building PinShare for Windows

Complete guide for building PinShare Windows distribution from source.

## Prerequisites

### Required Tools

1. **Go 1.24 or later**
   - Download: https://golang.org/dl/
   - Verify: `go version`

2. **Git for Windows** (includes Git Bash)
   - Download: https://git-scm.com/
   - Verify: `git --version`
   - **Note:** Use Git Bash as the preferred shell for running build commands

### Platform-Specific Requirements

#### Building on Windows (Git Bash)

**Required:**
- **Git Bash** (preferred shell for build commands)
  - Included with Git for Windows

- **WiX Toolset 6** (for installer, installed via .NET tool)
  - Requires .NET SDK 6+: https://dotnet.microsoft.com/download
  - WiX is installed automatically by build scripts via `dotnet tool install --global wix`

#### Cross-Compiling from Linux (Debian/Ubuntu)

**Required packages:**
```bash
sudo apt-get update
sudo apt-get install -y unzip curl

# Optional: Wine for testing Windows binaries
sudo apt-get install -y wine64
```

#### Cross-Compiling from macOS

No additional dependencies required beyond Go and Git.

## Building Components

### Clone Repository

```bash
git clone https://github.com/Cypherpunk-Labs/PinShare.git
cd PinShare
```

### Option 1: Build Everything (Recommended)

Use the provided build script (preferred over make targets):

```bash
# On any platform (macOS, Linux, Windows Git Bash)
./build-windows.sh

# On Windows (CMD or PowerShell)
.\build-windows.bat
```

**Platform behavior:**
- **Windows (Git Bash):** `build-windows.sh` delegates to `build-windows.bat` for native builds
- **macOS/Linux:** `build-windows.sh` cross-compiles Windows binaries (CGO disabled)
- **Windows (CMD):** Use `build-windows.bat` directly

This will:
1. Build PinShare backend (`pinshare.exe`)
2. Build Windows service wrapper (`pinsharesvc.exe`)
3. Build system tray application (`pinshare-tray.exe`)
4. Download IPFS Kubo binary
5. Optionally build the MSI installer

Output: `dist/windows/`

### Option 2: Build Individual Components

#### 1. Backend Binary

```bash
# On Linux/macOS (cross-compile)
GOOS=windows GOARCH=amd64 go build -o dist/windows/pinshare.exe .

# On Windows (Git Bash)
GOOS=windows GOARCH=amd64 go build -o dist/windows/pinshare.exe .
```

**Note:** CGO is disabled by default for Windows builds.

#### 2. Windows Service Wrapper

```bash
# On Linux/macOS
GOOS=windows GOARCH=amd64 \
go build -o dist/windows/pinsharesvc.exe ./cmd/pinsharesvc

# On Windows
go build -o dist\windows\pinsharesvc.exe .\cmd\pinsharesvc
```

#### 3. System Tray Application

```bash
# On Linux/macOS
GOOS=windows GOARCH=amd64 \
go build -ldflags="-H windowsgui" \
-o dist/windows/pinshare-tray.exe ./cmd/pinshare-tray

# On Windows
go build -ldflags="-H windowsgui" -o dist\windows\pinshare-tray.exe .\cmd\pinshare-tray
```

The `-H windowsgui` flag prevents a console window from appearing.

#### 4. IPFS Kubo Binary

```bash
# Download and extract
IPFS_VERSION=v0.31.0
curl -L -o /tmp/kubo.zip \
  "https://dist.ipfs.tech/kubo/${IPFS_VERSION}/kubo_${IPFS_VERSION}_windows-amd64.zip"

unzip -j /tmp/kubo.zip "kubo/ipfs.exe" -d dist/windows/
```

On Windows (PowerShell):
```powershell
$IPFS_VERSION = "v0.31.0"
$URL = "https://dist.ipfs.tech/kubo/$IPFS_VERSION/kubo_${IPFS_VERSION}_windows-amd64.zip"
Invoke-WebRequest -Uri $URL -OutFile kubo.zip
Expand-Archive -Path kubo.zip -DestinationPath .
Move-Item kubo\ipfs.exe dist\windows\
Remove-Item kubo.zip
Remove-Item -Recurse kubo
```

## Building the Installer

### Prerequisites

**WiX Toolset 6 must be installed via .NET tool.**

Verify:
```bash
wix --version
```

### Build Steps

#### On Windows (Git Bash or CMD)

```bash
cd installer

# Using shell script (works in Git Bash, delegates to .bat)
./build-wix6.sh [version]

# Or use batch file directly (CMD/PowerShell)
.\build-wix6.bat [version]
```

This uses WiX 4.x/6.x toolset (installed via `dotnet tool install`) to build the MSI installer.

Output: `installer/bin/Release/PinShare-Setup.msi`

#### On macOS/Linux

WiX cannot run natively on macOS/Linux. Options:

1. **CI/CD Pipeline:** Use GitHub Actions with Windows runners (recommended)
2. **Windows VM:** Copy built binaries to Windows and run `build-wix6.bat`
3. **Cross-compile binaries locally, build MSI in CI:** The `build-windows.sh` script creates all binaries; the MSI can be built by GitHub Actions

#### Using CI/CD

For automated builds, use GitHub Actions with Windows runners. See `.github/workflows/build.yml` for an example.

## Troubleshooting Build Issues

### WiX Errors

**Error:** `wix: command not found`

**Solution:** Install WiX via .NET tool:
```bash
dotnet tool install --global wix
```

**Error:** `The system cannot find the file specified`

**Solution:** Ensure all binaries are built:
```cmd
dir ..\dist\windows\*.exe
```

All required files must exist before building the installer.

## Advanced Build Options

### Custom Build Flags

```bash
# Build with version info
go build -ldflags="-X main.Version=1.0.0 -X main.GitCommit=$(git rev-parse --short HEAD)" ...

# Build with optimizations
go build -ldflags="-s -w" ...  # Strip debug info

# Static linking
go build -ldflags="-extldflags=-static" ...
```

### Code Signing

To sign binaries (requires code signing certificate):

```cmd
REM Sign executables
signtool sign /f certificate.pfx /p password /t http://timestamp.digicert.com dist\windows\*.exe

REM Sign installer
signtool sign /f certificate.pfx /p password /t http://timestamp.digicert.com dist\PinShare-Setup.msi
```

### Reproducible Builds

For deterministic builds:

```bash
# Set build timestamp
export SOURCE_DATE_EPOCH=1234567890

# Disable build ID
go build -ldflags="-buildid=" ...

# Use specific Go version
go1.24.0 build ...
```

## Testing Builds

### On Windows

```cmd
REM Install
msiexec /i PinShare-Setup.msi /l*v install.log

REM Test service
sc query PinShareService

REM Uninstall
msiexec /x PinShare-Setup.msi
```

### On Linux (with Wine)

```bash
# Test executables
wine64 dist/windows/pinsharesvc.exe

# Note: Full service functionality won't work in Wine
```

## Build Output

Successful build produces:

```
dist/
├── windows/
│   ├── pinshare.exe         (~50 MB with dependencies)
│   ├── pinsharesvc.exe      (~15 MB)
│   ├── pinshare-tray.exe    (~10 MB)
│   ├── ipfs.exe             (~85 MB)
│   └── resources/           (tray app resources)
└── PinShare-Setup.msi       (~100 MB)
```

## Next Steps

After building:

1. **Test the installer** on a clean Windows VM
2. **Verify all components** start and function correctly
3. **Check logs** for errors
4. **Document** any configuration changes
5. **Create release** with installer and checksums

## Resources

- [Go Cross Compilation](https://golang.org/doc/install/source#environment)
- [WiX Documentation](https://wixtoolset.org/documentation/)
- [IPFS Kubo Releases](https://dist.ipfs.tech/)

## Support

For build issues, check:
- [Troubleshooting](#troubleshooting-build-issues)
- [GitHub Issues](https://github.com/Cypherpunk-Labs/PinShare/issues)
- Build logs in `dist/build.log`

# Building PinShare for Windows

Complete guide for building PinShare Windows distribution from source.

## Prerequisites

### Required Tools

1. **Go 1.24 or later**
   - Download: https://golang.org/dl/
   - Verify: `go version`

2. **Node.js 20 or later**
   - Download: https://nodejs.org/
   - Verify: `node --version` and `npm --version`

3. **Git**
   - Download: https://git-scm.com/
   - Verify: `git --version`

### Platform-Specific Requirements

#### Building on Windows

**Required:**
- **TDM-GCC** or **MinGW-w64** (for CGO/SQLite)
  - TDM-GCC: https://jmeubank.github.io/tdm-gcc/
  - Or MinGW-w64: https://www.mingw-w64.org/

- **WiX Toolset 3.x or 4.x** (for installer)
  - Download: https://wixtoolset.org/
  - Add to PATH: `C:\Program Files (x86)\WiX Toolset v3.x\bin`

**Optional:**
- **Visual Studio Build Tools** (alternative to MinGW)
  - Download: https://visualstudio.microsoft.com/downloads/
  - Install "Desktop development with C++" workload

#### Cross-Compiling from Linux

**Required packages:**
```bash
sudo apt-get update
sudo apt-get install -y \
  gcc-mingw-w64-x86-64 \
  wine64 \
  unzip \
  curl
```

**For Debian/Ubuntu:**
```bash
# Add i386 architecture for Wine
sudo dpkg --add-architecture i386
sudo apt-get update
sudo apt-get install wine64 wine32
```

#### Cross-Compiling from macOS

**Required:**
```bash
# Install Homebrew if not already installed
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Install MinGW-w64
brew install mingw-w64
```

## Building Components

### Clone Repository

```bash
git clone https://github.com/Episk-pos/PinShare.git
cd PinShare
git checkout infra/refactor
```

### Option 1: Build Everything (Recommended)

```bash
# On Linux/macOS
make -f Makefile.windows windows-all

# On Windows
mingw32-make -f Makefile.windows windows-all
```

This will:
1. Build PinShare backend (`pinshare.exe`)
2. Build Windows service wrapper (`pinsharesvc.exe`)
3. Build system tray application (`pinshare-tray.exe`)
4. Build React UI (static files)
5. Download IPFS Kubo binary

Output: `dist/windows/`

### Option 2: Build Individual Components

#### 1. Backend Binary

```bash
# On Linux/macOS (cross-compile)
CGO_ENABLED=1 \
GOOS=windows \
GOARCH=amd64 \
CC=x86_64-w64-mingw32-gcc \
go build -o dist/windows/pinshare.exe .

# On Windows
set CGO_ENABLED=1
set GOOS=windows
set GOARCH=amd64
go build -o dist\windows\pinshare.exe .
```

**Note:** CGO is required for SQLite (`mattn/go-sqlite3`).

**Alternative:** Use pure-Go SQLite to avoid CGO:
- Replace `github.com/mattn/go-sqlite3` with `modernc.org/sqlite`
- Build without CGO: `CGO_ENABLED=0`

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

#### 4. React UI

```bash
cd pinshare-ui

# Install dependencies
npm install

# Build for production
npm run build

# Copy to distribution
mkdir -p ../dist/windows/ui
cp -r dist/* ../dist/windows/ui/
```

On Windows:
```cmd
cd pinshare-ui
npm install
npm run build
xcopy /E /I dist ..\dist\windows\ui
```

#### 5. IPFS Kubo Binary

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

**WiX Toolset must be installed and in PATH.**

Verify:
```cmd
candle.exe -?
light.exe -?
```

### Build Steps

#### On Windows

```cmd
cd installer
build.bat
```

This will:
1. Harvest UI files using `heat.exe`
2. Compile WiX sources with `candle.exe`
3. Link MSI package with `light.exe`
4. Output: `dist/PinShare-Setup.msi`

#### Manual Build (Windows)

```cmd
cd installer

REM Harvest UI files
heat.exe dir "..\dist\windows\ui" ^
  -cg UIComponents ^
  -dr UIFolder ^
  -gg -g1 -sf -srd ^
  -var var.UISourceDir ^
  -out UIComponents.wxs

REM Compile
candle.exe ^
  -ext WixUIExtension ^
  -ext WixUtilExtension ^
  -dUISourceDir="..\dist\windows\ui" ^
  Product.wxs UIComponents.wxs

REM Link
light.exe ^
  -ext WixUIExtension ^
  -ext WixUtilExtension ^
  -out PinShare-Setup.msi ^
  Product.wixobj UIComponents.wixobj

REM Move to dist
move PinShare-Setup.msi ..\dist\
```

#### On Linux (using Wine)

**Not recommended.** WiX under Wine is unreliable. Better options:

1. **Build on Windows VM**
   - Use VirtualBox/VMware
   - Share `dist/windows` folder
   - Run `build.bat` inside VM

2. **Use CI/CD**
   - GitHub Actions has Windows runners
   - See `.github/workflows/build.yml` example below

## Troubleshooting Build Issues

### CGO Errors

**Error:** `gcc: command not found`

**Linux/macOS Solution:**
```bash
# Install MinGW
sudo apt-get install gcc-mingw-w64-x86-64  # Debian/Ubuntu
brew install mingw-w64                      # macOS
```

**Windows Solution:**
```
Install TDM-GCC or MinGW-w64, add to PATH
```

**Error:** `undefined reference to...` (SQLite linking)

**Solution 1:** Ensure CGO is enabled
```bash
export CGO_ENABLED=1
export CC=x86_64-w64-mingw32-gcc  # Linux
```

**Solution 2:** Use pure-Go SQLite
```bash
# In go.mod, replace:
# github.com/mattn/go-sqlite3
# with:
# modernc.org/sqlite

# Then build without CGO
CGO_ENABLED=0 go build ...
```

### UI Build Errors

**Error:** `npm: command not found`

**Solution:** Install Node.js from https://nodejs.org/

**Error:** `EACCES: permission denied`

**Solution:**
```bash
# Don't use sudo with npm
# Fix npm permissions:
mkdir ~/.npm-global
npm config set prefix '~/.npm-global'
export PATH=~/.npm-global/bin:$PATH
```

**Error:** Build fails in `pinshare-ui/`

**Solution:**
```bash
# Clean and rebuild
cd pinshare-ui
rm -rf node_modules dist
npm install
npm run build
```

### IPFS Download Errors

**Error:** `curl: (6) Could not resolve host`

**Solution:** Check internet connection, try:
```bash
# Use wget instead
wget https://dist.ipfs.tech/kubo/v0.31.0/kubo_v0.31.0_windows-amd64.zip
```

**Error:** `unzip: command not found`

**Linux Solution:**
```bash
sudo apt-get install unzip
```

**Windows Solution:** Use PowerShell's `Expand-Archive` (see above)

### WiX Errors

**Error:** `candle.exe: command not found`

**Solution:** Add WiX to PATH:
```cmd
set PATH=%PATH%;C:\Program Files (x86)\WiX Toolset v3.11\bin
```

**Error:** `The system cannot find the file specified`

**Solution:** Ensure all binaries are built:
```cmd
dir ..\dist\windows\*.exe
dir ..\dist\windows\ui\index.html
```

All required files must exist before running `build.bat`.

## CI/CD Integration

### GitHub Actions Example

Create `.github/workflows/build-windows.yml`:

```yaml
name: Build Windows

on:
  push:
    branches: [ main, infra/refactor ]
  pull_request:
    branches: [ main ]

jobs:
  build:
    runs-on: windows-latest

    steps:
    - uses: actions/checkout@v3

    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.24'

    - name: Set up Node.js
      uses: actions/setup-node@v3
      with:
        node-version: '20'

    - name: Install dependencies
      run: |
        choco install wix311 -y
        refreshenv

    - name: Build Windows components
      run: |
        make -f Makefile.windows windows-all

    - name: Build installer
      run: |
        cd installer
        .\build.bat

    - name: Upload artifacts
      uses: actions/upload-artifact@v3
      with:
        name: pinshare-windows
        path: |
          dist/windows/*.exe
          dist/PinShare-Setup.msi
```

## Advanced Build Options

### Custom Build Flags

```bash
# Build with version info
go build -ldflags="-X main.Version=1.0.0 -X main.GitCommit=$(git rev-parse --short HEAD)" ...

# Build with optimizations
go build -ldflags="-s -w" ...  # Strip debug info

# Static linking (requires CGO_ENABLED=0)
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

REM Test UI
start http://localhost:8888

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
│   └── ui/
│       ├── index.html
│       ├── assets/
│       └── ...
└── PinShare-Setup.msi       (~150 MB)
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
- [MinGW-w64](https://www.mingw-w64.org/)

## Support

For build issues, check:
- [Troubleshooting](#troubleshooting-build-issues)
- [GitHub Issues](https://github.com/Episk-pos/PinShare/issues)
- Build logs in `dist/build.log`

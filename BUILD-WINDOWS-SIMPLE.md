# Building PinShare on Windows (No Make Required!)

## Quick Start

### Option 1: PowerShell (Recommended)

```powershell
.\build-windows.ps1
```

### Option 2: Batch Script (Git Bash / CMD)

```batch
build-windows.bat
```

Both scripts will:
1. ✅ Build all Go binaries
2. ✅ Build React UI
3. ✅ Download IPFS
4. ✅ Offer to build MSI installer

## Prerequisites

### Required

1. **Go 1.24+**
   - Download: https://golang.org/dl/
   - Verify: `go version`

2. **Node.js 20+**
   - Download: https://nodejs.org/
   - Verify: `node --version`

3. **Git**
   - Download: https://git-scm.com/
   - Verify: `git --version`

### For Installer (Optional)

4. **.NET SDK 6+**
   ```powershell
   winget install Microsoft.DotNet.SDK.8
   ```

5. **WiX Tool**
   ```powershell
   dotnet tool install --global wix
   ```

## Build Steps

### 1. Clone Repository

```bash
git clone https://github.com/Episk-pos/PinShare.git
cd PinShare
git checkout claude/windows-service-wrapper-plan-01NFgPq7Z22pinZbjqPcFHVu
```

### 2. Build Everything

**PowerShell:**
```powershell
.\build-windows.ps1
```

**Batch (Git Bash or CMD):**
```batch
build-windows.bat
```

**Manual (if scripts don't work):**

```batch
REM Create output directory
mkdir dist\windows

REM Build backend
go build -o dist\windows\pinshare.exe .

REM Build service wrapper
go build -o dist\windows\pinsharesvc.exe .\cmd\pinsharesvc

REM Build tray app
go build -ldflags "-H windowsgui" -o dist\windows\pinshare-tray.exe .\cmd\pinshare-tray

REM Build UI
cd pinshare-ui
npm install
npm run build
xcopy /E /I dist ..\dist\windows\ui
cd ..

REM Download IPFS
REM Download from: https://dist.ipfs.tech/kubo/v0.31.0/kubo_v0.31.0_windows-amd64.zip
REM Extract ipfs.exe to dist\windows\
```

### 3. Build Installer (Optional)

```batch
cd installer
build-wix6.bat
```

Output: `installer\bin\Release\PinShare-Setup.msi`

## Testing Without Installing

You can test PinShare without building the installer:

```powershell
cd dist\windows

# Run in debug mode (console window)
.\pinsharesvc.exe debug
```

This will:
- Start IPFS daemon
- Start PinShare backend
- Start UI server on http://localhost:8888
- Show all logs in console

Press `Ctrl+C` to stop.

## Installing

### From MSI (Recommended)

```powershell
# Install with UI
msiexec /i installer\bin\Release\PinShare-Setup.msi

# Silent install
msiexec /i installer\bin\Release\PinShare-Setup.msi /quiet
```

### Manual Install (Advanced)

```powershell
# Copy binaries
Copy-Item -Recurse dist\windows\* "C:\Program Files\PinShare\"

# Install service
cd "C:\Program Files\PinShare"
.\pinsharesvc.exe install
.\pinsharesvc.exe start

# Open UI
start http://localhost:8888
```

## Troubleshooting

### Error: "go: command not found"

Install Go from https://golang.org/dl/

### Error: "npm: command not found"

Install Node.js from https://nodejs.org/

### Error: "CGO_ENABLED requires gcc"

**Option 1 - Install TDM-GCC:**
- Download: https://jmeubank.github.io/tdm-gcc/
- Install and add to PATH

**Option 2 - Use pure Go SQLite (no CGO):**
```batch
REM Edit go.mod to use modernc.org/sqlite instead of mattn/go-sqlite3
REM Then build with:
set CGO_ENABLED=0
go build -o dist\windows\pinshare.exe .
```

### UI build fails

```batch
cd pinshare-ui
rmdir /s node_modules
del package-lock.json
npm install
npm run build
```

### IPFS download fails

Manually download from:
https://dist.ipfs.tech/kubo/v0.31.0/kubo_v0.31.0_windows-amd64.zip

Extract `ipfs.exe` to `dist\windows\`

## Build Output

After successful build:

```
dist/windows/
├── pinshare.exe         (~50 MB)
├── pinsharesvc.exe      (~15 MB)
├── pinshare-tray.exe    (~10 MB)
├── ipfs.exe             (~85 MB)
└── ui/
    ├── index.html
    └── assets/
```

## Next Steps

- 📖 **User Guide**: `docs/windows/README.md`
- 🏗️ **Full Build Guide**: `docs/windows/BUILD.md`
- 🚀 **Installer Guide**: `INSTALLER-QUICKSTART.md`

## Quick Links

- Build scripts: `build-windows.ps1` or `build-windows.bat`
- Test without installing: `dist\windows\pinsharesvc.exe debug`
- Build installer: `installer\build-wix6.bat`
- Open UI: http://localhost:8888

---

**No `make` required!** ✅

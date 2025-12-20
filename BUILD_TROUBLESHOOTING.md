# Build Troubleshooting Guide

## Quick Test

Before building, you can run the test script to verify all fixes are working:

```cmd
.\test-build-fixes.bat
```

This will:
1. Check .NET SDK installation and PATH
2. Verify WiX tool is installed
3. Test the IPFS download script
4. Verify the downloaded IPFS executable

If all tests pass, you're ready to build!

---

## Issues Found and Fixed

### 1. PowerShell IPFS Download Failures ✓ FIXED

**Symptoms:**
- `New-Object : Exception calling ".ctor" with "3" argument(s): "End of Central Directory record could not be found."`
- `Copy-Item : Cannot find path ... because it does not exist`

**Root Cause:**
The inline PowerShell command in `build-windows.bat` was too complex and had issues with:
- Archive extraction using `Expand-Archive`
- Path handling for temporary files
- Error handling

**Solution:**
Created dedicated PowerShell script `installer/download-ipfs.ps1` with:
- Proper error handling
- File verification (size checks)
- Recursive search for ipfs.exe in extracted archive
- Cleanup on success and failure
- Better logging

**Files Modified:**
- Created: `installer/download-ipfs.ps1`
- Modified: `build-windows.bat` (line 141)

---

### 2. .NET SDK Not Found / PATH Issue ✓ FIXED

**Error:**
```
ERROR: .NET SDK not found
Please install .NET SDK 6.0 or later from https://dotnet.microsoft.com/download
ERROR: Installer build failed
```

**Root Cause:**
The WiX 6 MSI installer requires .NET SDK 6.0 or later to build. Even if .NET SDK is installed, it may not be in the PATH environment variable, especially in a newly opened terminal window.

**Solution Applied:**
The build scripts now automatically detect .NET SDK in common installation locations and add it to PATH:
- `C:\Program Files\dotnet\` (system-wide installation)
- `%USERPROFILE%\.dotnet\` (user installation)

**Files Modified:**
- [build-windows.bat](build-windows.bat#L167-L175) - Auto-detects .NET SDK
- [installer/build-wix6.bat](installer/build-wix6.bat#L19-L36) - Auto-detects .NET SDK

**Manual Verification:**
If you want to verify .NET SDK is working:
```cmd
dotnet --version
```
Should output: `8.0.x`, `7.0.x`, or `6.0.x`

**If .NET SDK is Not Installed:**
1. **Download .NET SDK:**
   - Visit: https://dotnet.microsoft.com/download
   - Recommended: .NET 8.0 SDK (LTS)
   - Minimum: .NET 6.0 SDK

2. **After Installation:**
   - Close and reopen your terminal
   - OR run the test script: `.\test-build-fixes.bat`

**Alternative - Build Without MSI:**
If you only need the binaries and can skip the MSI installer:
```cmd
.\build-windows-no-installer.bat
```

This will build all executables but skip the MSI creation step.

---

## Windows 10/11 Compatibility ✓ VERIFIED

All tooling is compatible with both Windows 10 and Windows 11:

### PowerShell Requirements
- **Required:** PowerShell 5.1+ (built into Windows 10/11)
- **Features Used:**
  - `Invoke-WebRequest` - Standard cmdlet
  - `Expand-Archive` - Standard cmdlet
  - No PowerShell Core (7+) required

### .NET Requirements for MSI Build
- **.NET SDK 6.0+** - Compatible with Windows 10 (1607+) and Windows 11
- **WiX Toolset 6.x** - Compatible with both Windows versions

### Go Build Requirements
- **Go 1.21+** recommended
- Produces Windows executables compatible with:
  - Windows 10 (all versions)
  - Windows 11
  - Windows Server 2016+

### Runtime Requirements (for end users)
The built binaries require:
- Windows 10 version 1809+ or Windows 11
- No .NET runtime required (Go produces native executables)
- No additional dependencies

---

## Build Options

### Option A: Full Build with MSI Installer
**Prerequisites:**
- Go 1.21+
- .NET SDK 6.0+
- Git
- Node.js/npm (for UI, optional)

**Command:**
```cmd
.\build-windows.bat
```
Answer `Y` when prompted to build MSI.

**Output:**
- `dist/windows/pinshare.exe`
- `dist/windows/pinsharesvc.exe`
- `dist/windows/pinshare-tray.exe`
- `dist/windows/ipfs.exe`
- `installer/bin/Release/PinShare-Setup.msi`

---

### Option B: Binaries Only (No MSI)
**Prerequisites:**
- Go 1.21+
- Git
- Node.js/npm (for UI, optional)

**Command:**
```cmd
.\build-windows-no-installer.bat
```

**Output:**
- `dist/windows/pinshare.exe`
- `dist/windows/pinsharesvc.exe`
- `dist/windows/pinshare-tray.exe`
- `dist/windows/ipfs.exe`

You can build the MSI later after installing .NET SDK:
```cmd
cd installer
build-wix6.bat 1.0.0
```

---

## Testing the Fixes

### Test 1: IPFS Download
```cmd
REM Delete any existing IPFS
del dist\windows\ipfs.exe

REM Run build - should download successfully
.\build-windows-no-installer.bat
```

Expected output:
```
Downloading IPFS Kubo...
Downloading IPFS Kubo v0.31.0...
URL: https://dist.ipfs.tech/kubo/v0.31.0/kubo_v0.31.0_windows-amd64.zip
Downloaded to: C:\Users\...\Temp\kubo.zip
File size: XXXXXXX bytes
Extracting archive...
Found ipfs.exe at: C:\Users\...\Temp\kubo_extract\kubo\ipfs.exe
Copied to: dist\windows\ipfs.exe
SUCCESS: IPFS downloaded and extracted
[OK] Downloaded: dist\windows\ipfs.exe
```

### Test 2: Full Build
```cmd
REM Clean build
rmdir /s /q dist\windows

REM Build all components
.\build-windows-no-installer.bat
```

Expected binaries in `dist\windows\`:
- pinshare.exe (main application)
- pinsharesvc.exe (Windows service wrapper)
- pinshare-tray.exe (system tray app)
- ipfs.exe (IPFS daemon)
- resources\ (tray app icons)

---

## Known Limitations

1. **UI Build Temporarily Disabled:**
   The React UI (`pinshare-ui`) is being merged from another branch. The build scripts will skip UI building if the directory doesn't exist.

2. **MSI Build Requires .NET:**
   Cannot build the MSI installer without .NET SDK 6.0+. Use the binaries-only build option if .NET is not available.

3. **Cross-Compilation from macOS/Linux:**
   The bash script (`build-windows.sh`) supports cross-compilation, but the WiX MSI build must be done on Windows.

---

## Quick Reference

| Scenario | Command | Requirements |
|----------|---------|--------------|
| Full build with MSI | `.\build-windows.bat` | Go, .NET SDK 6.0+, Git |
| Binaries only | `.\build-windows-no-installer.bat` | Go, Git |
| MSI only (after binaries built) | `cd installer && build-wix6.bat 1.0.0` | .NET SDK 6.0+ |
| Cross-compile from macOS/Linux | `./build-windows.sh` | Go, curl, unzip |

---

## Support

For issues or questions:
- GitHub Issues: https://github.com/Cypherpunk-Labs/PinShare/issues
- Documentation: `docs/windows/BUILD.md`

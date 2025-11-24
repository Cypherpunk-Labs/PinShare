# Windows Installer Quick Start (WiX 6)

## Prerequisites

1. **Install .NET SDK** (6.0 or later):
   ```powershell
   # Windows (via winget)
   winget install Microsoft.DotNet.SDK.8

   # Or download from: https://dotnet.microsoft.com/download
   ```

2. **Install WiX .NET Tool**:
   ```powershell
   dotnet tool install --global wix

   # Verify installation
   wix --version
   ```

## Build the Installer

### Step 1: Build Windows Components

```bash
# From repository root
make -f Makefile.windows windows-all
```

This creates:
- `dist/windows/pinsharesvc.exe`
- `dist/windows/pinshare.exe`
- `dist/windows/pinshare-tray.exe`
- `dist/windows/ipfs.exe`
- `dist/windows/ui/` (React app)

### Step 2: Build the MSI

```bash
cd installer

# Windows:
build-wix6.bat

# Linux/macOS:
./build-wix6.sh
```

**Output**: `installer/bin/Release/PinShare-Setup.msi`

## What Changed from WiX 3?

| WiX 3 (Deprecated) | WiX 6 (Current) |
|-------------------|-----------------|
| Install toolset separately | `dotnet tool install --global wix` |
| `candle.exe` + `light.exe` | `dotnet build` |
| `build.bat` with manual commands | `build-wix6.bat` with MSBuild |
| `Product.wxs` | `Package.wxs` (new syntax) |
| Manual `heat.exe` for files | Auto-harvest in `.wixproj` |

## Key Files

- **`Package.wxs`** - Installer definition (WiX 6 format)
- **`PinShare.wixproj`** - MSBuild project file
- **`build-wix6.bat`** - Build automation script
- **`README-WIX6.md`** - Full documentation

## Testing

```powershell
# Install (with UI)
msiexec /i installer\bin\Release\PinShare-Setup.msi

# Silent install
msiexec /i installer\bin\Release\PinShare-Setup.msi /quiet

# Uninstall
msiexec /x installer\bin\Release\PinShare-Setup.msi
```

## Verify Installation

```powershell
# Check service
sc query PinShareService

# Check files
dir "C:\Program Files\PinShare"
dir "C:\ProgramData\PinShare"

# Open UI
start http://localhost:8888
```

## Troubleshooting

**Error: "dotnet: command not found"**
- Install .NET SDK: https://dotnet.microsoft.com/download

**Error: "wix: command not found"**
```powershell
dotnet tool install --global wix
# Add to PATH: %USERPROFILE%\.dotnet\tools
```

**Error: "binaries not found"**
- Run: `make -f Makefile.windows windows-all` first

**Error: "UI files not found"**
```bash
cd pinshare-ui
npm install
npm run build
```

## CI/CD

```yaml
# GitHub Actions example
- uses: actions/setup-dotnet@v3
  with:
    dotnet-version: '8.0.x'

- run: dotnet tool install --global wix

- run: |
    make -f Makefile.windows windows-all
    cd installer
    dotnet build PinShare.wixproj -c Release
```

## More Info

- Full docs: `installer/README-WIX6.md`
- Build guide: `docs/windows/BUILD.md`
- WiX docs: https://docs.firegiant.com/

---

**Status**: ✅ Complete and tested
**WiX Version**: 6.0.2
**Committed**: branch `claude/windows-service-wrapper-plan-01NFgPq7Z22pinZbjqPcFHVu`

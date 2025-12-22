# PinShare Windows Installer (WiX 6)

This directory contains the WiX 6 configuration for building the PinShare Windows installer.

## Prerequisites

### 1. .NET SDK 6.0 or later

```powershell
# Download from https://dotnet.microsoft.com/download
# Or via winget:
winget install Microsoft.DotNet.SDK.8

# Verify
dotnet --version
```

### 2. WiX .NET Tool

```powershell
# Install globally
dotnet tool install --global wix

# Verify
wix --version

# Update if already installed
dotnet tool update --global wix
```

## Building the Installer

### Quick Start

```bash
# 1. Build all Windows components first
make -f Makefile.windows windows-all

# 2. Build the installer
cd installer

# On Windows:
build-wix6.bat

# On Linux/macOS:
./build-wix6.sh
```

### Manual Build

If you prefer to build manually:

```powershell
# From installer directory
dotnet build PinShare.wixproj -c Release

# Output: bin/Release/PinShare-Setup.msi
```

## Project Structure

```
installer/
├── PinShare.wixproj         # MSBuild SDK-style project file
├── Package.wxs              # Main installer definition (WiX 6 format)
├── build-wix6.bat           # Automated build script (Windows)
├── build-wix6.sh            # Automated build script (Linux/macOS)
├── license.rtf              # License agreement
└── icon.ico                 # Application icon (optional)
```

## Configuration

### Version and Product Info

Edit `Package.wxs`:
```xml
<?define ProductName = "PinShare" ?>
<?define ProductVersion = "1.0.0" ?>
<?define Manufacturer = "PinShare Contributors" ?>
<?define UpgradeCode = "YOUR-GUID-HERE" ?>
```

**Generate new GUID:**
```powershell
[guid]::NewGuid()
```

### Ports and Settings

Edit registry values in `Package.wxs`:
```xml
<RegistryValue Name="UIPort" Type="integer" Value="8888" />
<RegistryValue Name="PinShareAPIPort" Type="integer" Value="9090" />
```

## Testing the Installer

### Install

```powershell
# Normal install (UI)
msiexec /i bin\Release\PinShare-Setup.msi

# With logging
msiexec /i bin\Release\PinShare-Setup.msi /l*v install.log

# Silent install
msiexec /i bin\Release\PinShare-Setup.msi /quiet /qn
```

### Uninstall

```powershell
# Normal uninstall (UI)
msiexec /x bin\Release\PinShare-Setup.msi

# Silent uninstall
msiexec /x bin\Release\PinShare-Setup.msi /quiet /qn
```

### Verify Installation

```powershell
# Check service is installed
sc query PinShareService

# Check registry
reg query "HKLM\SOFTWARE\PinShare"

# Check files
dir "C:\Program Files\PinShare"
dir "C:\ProgramData\PinShare"
```

## What the Installer Does

1. ✅ Installs binaries to `C:\Program Files\PinShare\`
   - pinsharesvc.exe
   - pinshare.exe
   - pinshare-tray.exe
   - ipfs.exe
   - ui/ (React app)

2. ✅ Creates data directories in `C:\ProgramData\PinShare\`
   - logs/
   - ipfs/
   - pinshare/
   - upload/
   - cache/
   - rejected/

3. ✅ Configures registry at `HKLM\SOFTWARE\PinShare`
   - Ports, paths, settings

4. ✅ Installs Windows service
   - Runs `pinsharesvc.exe install`
   - Sets to auto-start

5. ✅ Starts the service
   - Runs `pinsharesvc.exe start`

6. ✅ Adds system tray to startup
   - Creates shortcut in Startup folder

7. ✅ Creates Start Menu shortcuts
   - "Open PinShare UI"
   - "Uninstall PinShare"

## Troubleshooting

### Error: ".NET SDK not found"

Install .NET SDK 6.0 or later:
```powershell
winget install Microsoft.DotNet.SDK.8
```

### Error: "wix: command not found"

Install WiX .NET tool:
```powershell
dotnet tool install --global wix

# If it says already installed but still not found:
# Add to PATH: %USERPROFILE%\.dotnet\tools
```

### Error: "binaries not found"

Build the Windows components first:
```bash
make -f Makefile.windows windows-all
```

### Error: "UI files not found"

Build the React UI:
```bash
cd pinshare-ui
npm install
npm run build
```

### Build succeeds but MSI doesn't work

Check the build log for warnings:
```powershell
dotnet build PinShare.wixproj -v detailed
```

Common issues:
- Missing file references in Package.wxs
- Invalid registry keys
- Custom action failures

## Advanced Usage

### Custom Build Configuration

Edit `PinShare.wixproj`:

```xml
<PropertyGroup>
  <OutputName>PinShare-Setup-v1.0.0</OutputName>
  <ProductVersion>1.0.0</ProductVersion>
  <Platform>x64</Platform>
</PropertyGroup>
```

### Add More Files

For binaries:
```xml
<Component Id="MyNewComponent" Bitness="always64">
  <File Id="MyNewFile"
        Name="mynewfile.exe"
        Source="..\dist\windows\mynewfile.exe" />
</Component>
```

For directories (auto-harvested):
```xml
<ItemGroup>
  <HarvestDirectory Include="..\dist\windows\plugins">
    <ComponentGroupName>PluginComponents</ComponentGroupName>
    <DirectoryRefId>PluginsFolder</DirectoryRefId>
  </HarvestDirectory>
</ItemGroup>
```

### Code Signing

Sign the MSI after building:

```powershell
# Sign with certificate
signtool sign `
  /f certificate.pfx `
  /p password `
  /t http://timestamp.digicert.com `
  bin\Release\PinShare-Setup.msi
```

## CI/CD Integration

### GitHub Actions

```yaml
- name: Install .NET SDK
  uses: actions/setup-dotnet@v3
  with:
    dotnet-version: '8.0.x'

- name: Install WiX
  run: dotnet tool install --global wix

- name: Build Installer
  run: |
    cd installer
    dotnet build PinShare.wixproj -c Release

- name: Upload MSI
  uses: actions/upload-artifact@v3
  with:
    name: installer
    path: installer/bin/Release/*.msi
```

## Resources

- **WiX Documentation**: https://docs.firegiant.com/
- **WiX 6 on NuGet**: https://www.nuget.org/packages/wix
- **.NET Tool**: https://learn.microsoft.com/en-us/dotnet/core/tools/global-tools

## License

Same as PinShare - MIT License.

# Build PinShare for Windows
# This script builds all Windows components and the MSI installer

param(
    [switch]$SkipBinaries,
    [switch]$InstallerOnly
)

$ErrorActionPreference = "Stop"

Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "Building PinShare for Windows" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host ""

$repoRoot = $PSScriptRoot
$distDir = Join-Path $repoRoot "dist\windows"

# Create dist directory
if (-not (Test-Path $distDir)) {
    New-Item -ItemType Directory -Path $distDir -Force | Out-Null
}

if (-not $InstallerOnly) {
    # Build PinShare backend
    Write-Host "Building PinShare backend..." -ForegroundColor Yellow
    $env:CGO_ENABLED = "1"
    $env:GOOS = "windows"
    $env:GOARCH = "amd64"

    go build -ldflags "-s -w" -o "$distDir\pinshare.exe" .
    if ($LASTEXITCODE -ne 0) { throw "Failed to build pinshare.exe" }
    Write-Host "✓ Built: $distDir\pinshare.exe" -ForegroundColor Green
    Write-Host ""

    # Build Windows service wrapper
    Write-Host "Building Windows service wrapper..." -ForegroundColor Yellow
    go build -ldflags "-s -w" -o "$distDir\pinsharesvc.exe" .\cmd\pinsharesvc
    if ($LASTEXITCODE -ne 0) { throw "Failed to build pinsharesvc.exe" }
    Write-Host "✓ Built: $distDir\pinsharesvc.exe" -ForegroundColor Green
    Write-Host ""

    # Build system tray application
    Write-Host "Building system tray application..." -ForegroundColor Yellow
    go build -ldflags "-s -w -H windowsgui" -o "$distDir\pinshare-tray.exe" .\cmd\pinshare-tray
    if ($LASTEXITCODE -ne 0) { throw "Failed to build pinshare-tray.exe" }
    Write-Host "✓ Built: $distDir\pinshare-tray.exe" -ForegroundColor Green

    # Copy tray resources (settings dialog, etc.)
    $trayResourcesSrc = Join-Path $repoRoot "cmd\pinshare-tray\resources"
    $trayResourcesDst = Join-Path $distDir "resources"
    if (Test-Path $trayResourcesSrc) {
        if (Test-Path $trayResourcesDst) {
            Remove-Item $trayResourcesDst -Recurse -Force
        }
        Copy-Item -Path $trayResourcesSrc -Destination $trayResourcesDst -Recurse -Force
        Write-Host "✓ Copied: $distDir\resources\" -ForegroundColor Green
    }
    Write-Host ""

    # Build React UI
    Write-Host "Building React UI..." -ForegroundColor Yellow
    Push-Location .\pinshare-ui
    try {
        if (-not (Test-Path "node_modules")) {
            Write-Host "Installing npm dependencies..." -ForegroundColor Gray
            npm install
            if ($LASTEXITCODE -ne 0) { throw "Failed to install npm dependencies" }
        }

        npm run build
        if ($LASTEXITCODE -ne 0) { throw "Failed to build UI" }

        # Copy UI files
        $uiDest = Join-Path $distDir "ui"
        if (Test-Path $uiDest) {
            Remove-Item $uiDest -Recurse -Force
        }
        Copy-Item -Path "dist\*" -Destination $uiDest -Recurse -Force
        Write-Host "✓ Built: $distDir\ui\" -ForegroundColor Green
    } finally {
        Pop-Location
    }
    Write-Host ""

    # Download IPFS if not present
    $ipfsExe = Join-Path $distDir "ipfs.exe"
    if (-not (Test-Path $ipfsExe)) {
        Write-Host "Downloading IPFS Kubo..." -ForegroundColor Yellow
        $ipfsVersion = "v0.31.0"
        $ipfsUrl = "https://dist.ipfs.tech/kubo/$ipfsVersion/kubo_${ipfsVersion}_windows-amd64.zip"
        $zipPath = Join-Path $env:TEMP "kubo.zip"

        Invoke-WebRequest -Uri $ipfsUrl -OutFile $zipPath
        Expand-Archive -Path $zipPath -DestinationPath $env:TEMP -Force
        Copy-Item (Join-Path $env:TEMP "kubo\ipfs.exe") $ipfsExe
        Remove-Item $zipPath -Force
        Remove-Item (Join-Path $env:TEMP "kubo") -Recurse -Force
        Write-Host "✓ Downloaded: $ipfsExe" -ForegroundColor Green
    } else {
        Write-Host "✓ IPFS already present: $ipfsExe" -ForegroundColor Green
    }
    Write-Host ""

    Write-Host "==========================================" -ForegroundColor Cyan
    Write-Host "All Windows components built successfully!" -ForegroundColor Green
    Write-Host "==========================================" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Binaries:" -ForegroundColor Cyan
    Get-ChildItem $distDir -Filter "*.exe" | ForEach-Object {
        Write-Host "  $($_.Name) - $([math]::Round($_.Length / 1MB, 2)) MB"
    }
    Write-Host ""
}

# Build installer if WiX is available
Write-Host "Checking for WiX .NET tool..." -ForegroundColor Yellow
$wixInstalled = Get-Command wix -ErrorAction SilentlyContinue
if (-not $wixInstalled) {
    Write-Host "WiX tool not found. Would you like to install it? (Y/N)" -ForegroundColor Yellow
    $response = Read-Host
    if ($response -eq "Y" -or $response -eq "y") {
        Write-Host "Installing WiX .NET tool..." -ForegroundColor Yellow
        dotnet tool install --global wix
        if ($LASTEXITCODE -ne 0) {
            Write-Host "Failed to install WiX. You can install it manually with:" -ForegroundColor Red
            Write-Host "  dotnet tool install --global wix" -ForegroundColor Gray
            exit 1
        }
        # Refresh PATH
        $env:PATH = [System.Environment]::GetEnvironmentVariable("PATH", "User") + ";" + [System.Environment]::GetEnvironmentVariable("PATH", "Machine")
    } else {
        Write-Host "Skipping installer build. To build later, run:" -ForegroundColor Yellow
        Write-Host "  cd installer" -ForegroundColor Gray
        Write-Host "  .\build-wix6.bat" -ForegroundColor Gray
        exit 0
    }
}

Write-Host ""
Write-Host "Building MSI installer..." -ForegroundColor Yellow
Push-Location .\installer
try {
    .\build-wix6.bat
    if ($LASTEXITCODE -ne 0) { throw "Failed to build installer" }

    Write-Host ""
    Write-Host "==========================================" -ForegroundColor Cyan
    Write-Host "Build Complete!" -ForegroundColor Green
    Write-Host "==========================================" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Installer: " -NoNewline
    Write-Host "installer\bin\Release\PinShare-Setup.msi" -ForegroundColor Green
    Write-Host ""
    Write-Host "To install, run:" -ForegroundColor Cyan
    Write-Host "  msiexec /i installer\bin\Release\PinShare-Setup.msi" -ForegroundColor Gray
} finally {
    Pop-Location
}

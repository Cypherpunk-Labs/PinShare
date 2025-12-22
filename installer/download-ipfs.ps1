# PowerShell script to download and extract IPFS Kubo
# This script is more robust than inline PowerShell commands

param(
    [string]$DestDir = "..\dist\windows",
    [string]$Version = "v0.31.0"
)

$ErrorActionPreference = "Stop"

$url = "https://dist.ipfs.tech/kubo/${Version}/kubo_${Version}_windows-amd64.zip"
$tempZip = Join-Path $env:TEMP "kubo.zip"
$tempExtract = Join-Path $env:TEMP "kubo_extract"
$destFile = Join-Path $DestDir "ipfs.exe"

Write-Host "Downloading IPFS Kubo ${Version}..."
Write-Host "URL: $url"

# Clean up any previous failed downloads
if (Test-Path $tempZip) {
    Write-Host "Removing previous download attempt..."
    Remove-Item $tempZip -Force -ErrorAction SilentlyContinue
}
if (Test-Path $tempExtract) {
    Write-Host "Removing previous extraction attempt..."
    Remove-Item $tempExtract -Recurse -Force -ErrorAction SilentlyContinue
}

try {
    # Download the file with progress and better error handling
    Write-Host "Starting download..."
    $ProgressPreference = 'SilentlyContinue'  # Faster downloads

    try {
        Invoke-WebRequest -Uri $url -OutFile $tempZip -UseBasicParsing -TimeoutSec 300
    } catch {
        throw "Download failed: $($_.Exception.Message)"
    }

    Write-Host "Downloaded to: $tempZip"

    # Verify the file exists and has content
    if (-not (Test-Path $tempZip)) {
        throw "Download failed - file not found at $tempZip"
    }

    $fileSize = (Get-Item $tempZip).Length
    Write-Host "File size: $fileSize bytes"

    # Check if file is suspiciously small (likely an error page)
    if ($fileSize -lt 1000000) {
        Write-Host "WARNING: File seems too small for IPFS Kubo (expected ~17MB)"
        # Check if it's an HTML error page
        $fileContent = Get-Content $tempZip -Raw -ErrorAction SilentlyContinue
        if ($fileContent -match '<html|<!DOCTYPE') {
            throw "Download failed - received HTML error page instead of zip file. URL may be incorrect."
        }
        throw "Download failed - file too small ($fileSize bytes, expected ~17MB)"
    }

    # Verify it's actually a ZIP file by checking magic bytes
    $bytes = [System.IO.File]::ReadAllBytes($tempZip)
    if ($bytes.Length -lt 4 -or $bytes[0] -ne 0x50 -or $bytes[1] -ne 0x4B) {
        throw "Downloaded file is not a valid ZIP archive (magic bytes check failed)"
    }

    # Clean up any previous extraction
    if (Test-Path $tempExtract) {
        Remove-Item $tempExtract -Recurse -Force
    }

    # Extract the archive
    Write-Host "Extracting archive..."
    try {
        # Use .NET's ZipFile class for more reliable extraction
        Add-Type -AssemblyName System.IO.Compression.FileSystem
        [System.IO.Compression.ZipFile]::ExtractToDirectory($tempZip, $tempExtract)
    } catch {
        Write-Host "Standard extraction failed, trying Expand-Archive..."
        # Fallback to Expand-Archive
        Expand-Archive -Path $tempZip -DestinationPath $tempExtract -Force
    }

    # Find ipfs.exe in the extracted files
    $ipfsExe = Get-ChildItem -Path $tempExtract -Filter "ipfs.exe" -Recurse | Select-Object -First 1

    if (-not $ipfsExe) {
        throw "ipfs.exe not found in archive"
    }

    Write-Host "Found ipfs.exe at: $($ipfsExe.FullName)"

    # Ensure destination directory exists
    $destDir = Split-Path $destFile -Parent
    if (-not (Test-Path $destDir)) {
        New-Item -ItemType Directory -Path $destDir -Force | Out-Null
    }

    # Copy to destination
    Copy-Item $ipfsExe.FullName $destFile -Force
    Write-Host "Copied to: $destFile"

    # Cleanup
    Remove-Item $tempZip -Force
    Remove-Item $tempExtract -Recurse -Force

    Write-Host "SUCCESS: IPFS downloaded and extracted"
    exit 0

} catch {
    Write-Host "ERROR: $_"
    Write-Host $_.Exception.Message

    # Cleanup on error
    if (Test-Path $tempZip) { Remove-Item $tempZip -Force -ErrorAction SilentlyContinue }
    if (Test-Path $tempExtract) { Remove-Item $tempExtract -Recurse -Force -ErrorAction SilentlyContinue }

    exit 1
}
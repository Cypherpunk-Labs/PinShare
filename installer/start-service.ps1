# PowerShell script to reliably start the PinShare service
# This script handles Windows 11's stricter service startup requirements

param(
    [string]$ServiceName = "PinShareService",
    [int]$MaxRetries = 3,
    [int]$RetryDelaySeconds = 2
)

$ErrorActionPreference = "Stop"

Write-Host "Starting service: $ServiceName"

# Function to check if service exists
function Test-ServiceExists {
    param([string]$Name)
    $service = Get-Service -Name $Name -ErrorAction SilentlyContinue
    return $null -ne $service
}

# Function to get service status
function Get-ServiceStatus {
    param([string]$Name)
    $service = Get-Service -Name $Name -ErrorAction SilentlyContinue
    if ($service) {
        return $service.Status
    }
    return $null
}

# Check if service exists
if (-not (Test-ServiceExists -Name $ServiceName)) {
    Write-Host "ERROR: Service $ServiceName not found"
    exit 1
}

# Try to start the service with retries
$attempt = 0
$started = $false

while ($attempt -lt $MaxRetries -and -not $started) {
    $attempt++

    try {
        $status = Get-ServiceStatus -Name $ServiceName

        if ($status -eq "Running") {
            Write-Host "Service is already running"
            $started = $true
            break
        }

        if ($attempt -gt 1) {
            Write-Host "Attempt $attempt of $MaxRetries..."
            Start-Sleep -Seconds $RetryDelaySeconds
        }

        Write-Host "Starting service..."
        Start-Service -Name $ServiceName -ErrorAction Stop

        # Wait for service to reach Running state
        $timeout = 30
        $elapsed = 0
        while ($elapsed -lt $timeout) {
            $status = Get-ServiceStatus -Name $ServiceName
            if ($status -eq "Running") {
                Write-Host "SUCCESS: Service started successfully"
                $started = $true
                break
            }
            Start-Sleep -Milliseconds 500
            $elapsed += 0.5
        }

        if (-not $started) {
            throw "Service did not reach Running state within ${timeout}s"
        }

    } catch {
        Write-Host "Failed to start service: $_"
        if ($attempt -eq $MaxRetries) {
            Write-Host "ERROR: Failed to start service after $MaxRetries attempts"
            Write-Host "You can start it manually later using: Start-Service $ServiceName"
            # Don't exit with error - service is installed, just not started
            # This prevents installation rollback
            exit 0
        }
    }
}

if ($started) {
    exit 0
} else {
    Write-Host "Service installation completed, but service is not running"
    Write-Host "You can start it manually using: Start-Service $ServiceName"
    exit 0
}

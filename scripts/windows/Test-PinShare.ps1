# Test-PinShare.ps1
# Comprehensive testing script for PinShare Windows service
# Provides rapid feedback for development and testing

#Requires -Version 5.1
#Requires -RunAsAdministrator

[CmdletBinding()]
param(
    [Parameter()]
    [ValidateSet('All', 'Build', 'Service', 'Health', 'API', 'UI', 'Integration', 'Cleanup')]
    [string]$TestSuite = 'All',

    [Parameter()]
    [switch]$SkipBuild,

    [Parameter()]
    [switch]$Verbose,

    [Parameter()]
    [switch]$KeepData,

    [Parameter()]
    [string]$LogPath = "$PSScriptRoot\..\..\test-results"
)

$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'

# Configuration
$script:Config = @{
    ServiceName = 'PinShareService'
    InstallDir = 'C:\Program Files\PinShare'
    DataDir = 'C:\ProgramData\PinShare'
    APIPort = 9090
    UIPort = 8888
    IPFSAPIPort = 5001
    IPFSSwarmPort = 4001
    LogPath = $LogPath
    TestTimeout = 300  # 5 minutes
}

# Test results tracking
$script:TestResults = @{
    Passed = 0
    Failed = 0
    Skipped = 0
    Tests = @()
}

#region Utility Functions

function Write-TestHeader {
    param([string]$Message)
    Write-Host "`n$('='*80)" -ForegroundColor Cyan
    Write-Host "  $Message" -ForegroundColor Cyan
    Write-Host "$('='*80)`n" -ForegroundColor Cyan
}

function Write-TestResult {
    param(
        [string]$TestName,
        [bool]$Passed,
        [string]$Message = '',
        [object]$Details = $null
    )

    $result = @{
        Name = $TestName
        Passed = $Passed
        Message = $Message
        Details = $Details
        Timestamp = Get-Date
    }

    $script:TestResults.Tests += $result

    if ($Passed) {
        $script:TestResults.Passed++
        Write-Host "✓ PASS: $TestName" -ForegroundColor Green
        if ($Message) { Write-Host "  $Message" -ForegroundColor Gray }
    } else {
        $script:TestResults.Failed++
        Write-Host "✗ FAIL: $TestName" -ForegroundColor Red
        if ($Message) { Write-Host "  $Message" -ForegroundColor Yellow }
        if ($Details) { Write-Host "  Details: $($Details | ConvertTo-Json -Depth 3)" -ForegroundColor Gray }
    }
}

function Test-Port {
    param(
        [int]$Port,
        [string]$Host = 'localhost',
        [int]$TimeoutMs = 1000
    )

    try {
        $tcpClient = New-Object System.Net.Sockets.TcpClient
        $asyncResult = $tcpClient.BeginConnect($Host, $Port, $null, $null)
        $wait = $asyncResult.AsyncWaitHandle.WaitOne($TimeoutMs)

        if ($wait) {
            $tcpClient.EndConnect($asyncResult)
            $tcpClient.Close()
            return $true
        } else {
            $tcpClient.Close()
            return $false
        }
    } catch {
        return $false
    }
}

function Wait-ForPort {
    param(
        [int]$Port,
        [int]$TimeoutSeconds = 30,
        [string]$Description = "Port $Port"
    )

    Write-Host "Waiting for $Description..." -NoNewline
    $stopwatch = [System.Diagnostics.Stopwatch]::StartNew()

    while ($stopwatch.Elapsed.TotalSeconds -lt $TimeoutSeconds) {
        if (Test-Port -Port $Port) {
            Write-Host " Ready!" -ForegroundColor Green
            return $true
        }
        Start-Sleep -Milliseconds 500
        Write-Host "." -NoNewline
    }

    Write-Host " Timeout!" -ForegroundColor Red
    return $false
}

function Invoke-HTTPRequest {
    param(
        [string]$Uri,
        [string]$Method = 'GET',
        [hashtable]$Headers = @{},
        [object]$Body = $null,
        [int]$TimeoutSec = 10
    )

    try {
        $params = @{
            Uri = $Uri
            Method = $Method
            Headers = $Headers
            TimeoutSec = $TimeoutSec
            UseBasicParsing = $true
        }

        if ($Body) {
            $params.Body = ($Body | ConvertTo-Json -Depth 10)
            $params.ContentType = 'application/json'
        }

        $response = Invoke-WebRequest @params
        return @{
            Success = $true
            StatusCode = $response.StatusCode
            Content = $response.Content
            Headers = $response.Headers
        }
    } catch {
        return @{
            Success = $false
            Error = $_.Exception.Message
            StatusCode = $_.Exception.Response.StatusCode.value__
        }
    }
}

#endregion

#region Build Tests

function Test-BuildEnvironment {
    Write-TestHeader "Testing Build Environment"

    # Check Go installation
    try {
        $goVersion = & go version 2>&1
        Write-TestResult -TestName "Go Installation" -Passed $true -Message $goVersion
    } catch {
        Write-TestResult -TestName "Go Installation" -Passed $false -Message "Go not found in PATH"
        return $false
    }

    # Check required Go version (1.21+)
    if ($goVersion -match 'go(\d+\.\d+)') {
        $version = [version]$matches[1]
        $required = [version]"1.21"
        $versionOk = $version -ge $required
        Write-TestResult -TestName "Go Version >= 1.21" -Passed $versionOk -Message "Found: $version, Required: $required"
    }

    # Check GCC for CGo (optional but recommended)
    try {
        $gccVersion = & gcc --version 2>&1 | Select-Object -First 1
        Write-TestResult -TestName "GCC Installation (Optional)" -Passed $true -Message $gccVersion
    } catch {
        Write-TestResult -TestName "GCC Installation (Optional)" -Passed $true -Message "GCC not found (OK - will use CGO_ENABLED=0)"
    }

    # Check source files
    $requiredFiles = @(
        "go.mod",
        "main.go",
        "cmd\pinsharesvc\main.go",
        "cmd\pinshare-tray\main.go",
        "Makefile.windows"
    )

    foreach ($file in $requiredFiles) {
        $exists = Test-Path (Join-Path $PSScriptRoot "..\..\$file")
        Write-TestResult -TestName "Source File: $file" -Passed $exists -Message (if ($exists) { "Found" } else { "Missing" })
    }

    return $true
}

function Test-BuildService {
    Write-TestHeader "Building PinShare Service"

    Push-Location (Join-Path $PSScriptRoot "..\..")

    try {
        # Build service executable
        Write-Host "Building pinsharesvc.exe..." -ForegroundColor Cyan
        $env:CGO_ENABLED = "0"
        $env:GOOS = "windows"
        $env:GOARCH = "amd64"

        $buildOutput = & go build -v -o "build\pinsharesvc.exe" ".\cmd\pinsharesvc" 2>&1
        $serviceBuilt = $LASTEXITCODE -eq 0
        Write-TestResult -TestName "Build pinsharesvc.exe" -Passed $serviceBuilt -Details $buildOutput

        # Build tray application
        Write-Host "Building pinshare-tray.exe..." -ForegroundColor Cyan
        $buildOutput = & go build -v -ldflags "-H=windowsgui" -o "build\pinshare-tray.exe" ".\cmd\pinshare-tray" 2>&1
        $trayBuilt = $LASTEXITCODE -eq 0
        Write-TestResult -TestName "Build pinshare-tray.exe" -Passed $trayBuilt -Details $buildOutput

        # Verify binaries
        if ($serviceBuilt) {
            $serviceExe = Join-Path $PWD "build\pinsharesvc.exe"
            $fileInfo = Get-Item $serviceExe
            Write-TestResult -TestName "Verify pinsharesvc.exe" -Passed $true -Message "Size: $($fileInfo.Length) bytes"
        }

        if ($trayBuilt) {
            $trayExe = Join-Path $PWD "build\pinshare-tray.exe"
            $fileInfo = Get-Item $trayExe
            Write-TestResult -TestName "Verify pinshare-tray.exe" -Passed $true -Message "Size: $($fileInfo.Length) bytes"
        }

        return ($serviceBuilt -and $trayBuilt)
    } finally {
        Pop-Location
        Remove-Item Env:\CGO_ENABLED -ErrorAction SilentlyContinue
        Remove-Item Env:\GOOS -ErrorAction SilentlyContinue
        Remove-Item Env:\GOARCH -ErrorAction SilentlyContinue
    }
}

#endregion

#region Service Tests

function Test-ServiceInstallation {
    Write-TestHeader "Testing Service Installation"

    $serviceExe = Join-Path $PSScriptRoot "..\..\build\pinsharesvc.exe"

    if (-not (Test-Path $serviceExe)) {
        Write-TestResult -TestName "Service Executable Exists" -Passed $false -Message "Build required first"
        return $false
    }

    # Uninstall if already installed
    $existing = Get-Service -Name $script:Config.ServiceName -ErrorAction SilentlyContinue
    if ($existing) {
        Write-Host "Removing existing service..." -ForegroundColor Yellow
        & $serviceExe uninstall
        Start-Sleep -Seconds 2
    }

    # Install service
    Write-Host "Installing service..." -ForegroundColor Cyan
    $installOutput = & $serviceExe install 2>&1
    $installed = $LASTEXITCODE -eq 0
    Write-TestResult -TestName "Install Service" -Passed $installed -Details $installOutput

    if (-not $installed) {
        return $false
    }

    # Verify service exists
    Start-Sleep -Seconds 1
    $service = Get-Service -Name $script:Config.ServiceName -ErrorAction SilentlyContinue
    $serviceExists = $null -ne $service
    Write-TestResult -TestName "Service Registered" -Passed $serviceExists

    if ($serviceExists) {
        Write-TestResult -TestName "Service Status" -Passed $true -Message "Status: $($service.Status), StartType: $($service.StartType)"
    }

    return $serviceExists
}

function Test-ServiceStart {
    Write-TestHeader "Testing Service Start"

    $service = Get-Service -Name $script:Config.ServiceName -ErrorAction SilentlyContinue
    if (-not $service) {
        Write-TestResult -TestName "Service Exists" -Passed $false
        return $false
    }

    if ($service.Status -eq 'Running') {
        Write-Host "Service already running, stopping first..." -ForegroundColor Yellow
        Stop-Service -Name $script:Config.ServiceName -Force
        Start-Sleep -Seconds 3
    }

    # Start service
    Write-Host "Starting service..." -ForegroundColor Cyan
    Start-Service -Name $script:Config.ServiceName

    # Wait for service to start
    $timeout = 30
    $stopwatch = [System.Diagnostics.Stopwatch]::StartNew()

    while ($stopwatch.Elapsed.TotalSeconds -lt $timeout) {
        $service = Get-Service -Name $script:Config.ServiceName
        if ($service.Status -eq 'Running') {
            Write-TestResult -TestName "Service Started" -Passed $true -Message "Start time: $($stopwatch.Elapsed.TotalSeconds)s"
            return $true
        }
        Start-Sleep -Milliseconds 500
    }

    Write-TestResult -TestName "Service Started" -Passed $false -Message "Timeout after ${timeout}s"
    return $false
}

function Test-ServiceStop {
    Write-TestHeader "Testing Service Stop"

    $service = Get-Service -Name $script:Config.ServiceName -ErrorAction SilentlyContinue
    if (-not $service -or $service.Status -ne 'Running') {
        Write-TestResult -TestName "Service Running" -Passed $false
        return $false
    }

    # Stop service
    Write-Host "Stopping service..." -ForegroundColor Cyan
    Stop-Service -Name $script:Config.ServiceName -Force

    # Wait for service to stop
    $timeout = 30
    $stopwatch = [System.Diagnostics.Stopwatch]::StartNew()

    while ($stopwatch.Elapsed.TotalSeconds -lt $timeout) {
        $service = Get-Service -Name $script:Config.ServiceName
        if ($service.Status -eq 'Stopped') {
            Write-TestResult -TestName "Service Stopped" -Passed $true -Message "Stop time: $($stopwatch.Elapsed.TotalSeconds)s"
            return $true
        }
        Start-Sleep -Milliseconds 500
    }

    Write-TestResult -TestName "Service Stopped" -Passed $false -Message "Timeout after ${timeout}s"
    return $false
}

#endregion

#region Health Tests

function Test-IPFSHealth {
    Write-TestHeader "Testing IPFS Health"

    # Wait for IPFS to be ready
    if (-not (Wait-ForPort -Port $script:Config.IPFSAPIPort -TimeoutSeconds 60 -Description "IPFS API")) {
        Write-TestResult -TestName "IPFS Port Listening" -Passed $false
        return $false
    }

    Write-TestResult -TestName "IPFS Port Listening" -Passed $true

    # Test IPFS version endpoint
    $versionResponse = Invoke-HTTPRequest -Uri "http://localhost:$($script:Config.IPFSAPIPort)/api/v0/version"
    Write-TestResult -TestName "IPFS Version API" -Passed $versionResponse.Success -Details $versionResponse

    if ($versionResponse.Success) {
        $version = ($versionResponse.Content | ConvertFrom-Json).Version
        Write-Host "  IPFS Version: $version" -ForegroundColor Gray
    }

    # Test IPFS ID
    $idResponse = Invoke-HTTPRequest -Uri "http://localhost:$($script:Config.IPFSAPIPort)/api/v0/id" -Method POST
    Write-TestResult -TestName "IPFS ID API" -Passed $idResponse.Success -Details $idResponse

    if ($idResponse.Success) {
        $id = ($idResponse.Content | ConvertFrom-Json).ID
        Write-Host "  Peer ID: $id" -ForegroundColor Gray
    }

    return ($versionResponse.Success -and $idResponse.Success)
}

function Test-PinShareHealth {
    Write-TestHeader "Testing PinShare API Health"

    # Wait for PinShare API to be ready
    if (-not (Wait-ForPort -Port $script:Config.APIPort -TimeoutSeconds 60 -Description "PinShare API")) {
        Write-TestResult -TestName "PinShare Port Listening" -Passed $false
        return $false
    }

    Write-TestResult -TestName "PinShare Port Listening" -Passed $true

    # Test health endpoint
    $healthResponse = Invoke-HTTPRequest -Uri "http://localhost:$($script:Config.APIPort)/health"
    Write-TestResult -TestName "PinShare Health Endpoint" -Passed $healthResponse.Success -Details $healthResponse

    # Test files endpoint
    $filesResponse = Invoke-HTTPRequest -Uri "http://localhost:$($script:Config.APIPort)/api/v1/files"
    Write-TestResult -TestName "PinShare Files API" -Passed $filesResponse.Success -Details $filesResponse

    if ($filesResponse.Success) {
        try {
            $files = $filesResponse.Content | ConvertFrom-Json
            Write-Host "  Files count: $($files.Count)" -ForegroundColor Gray
        } catch {
            Write-Host "  Could not parse response" -ForegroundColor Yellow
        }
    }

    return ($healthResponse.Success -or $filesResponse.Success)
}

function Test-UIServerHealth {
    Write-TestHeader "Testing UI Server Health"

    # Wait for UI server to be ready
    if (-not (Wait-ForPort -Port $script:Config.UIPort -TimeoutSeconds 30 -Description "UI Server")) {
        Write-TestResult -TestName "UI Port Listening" -Passed $false
        return $false
    }

    Write-TestResult -TestName "UI Port Listening" -Passed $true

    # Test UI root
    $uiResponse = Invoke-HTTPRequest -Uri "http://localhost:$($script:Config.UIPort)/"
    Write-TestResult -TestName "UI Root Endpoint" -Passed $uiResponse.Success -Details $uiResponse

    if ($uiResponse.Success -and $uiResponse.Content -match 'PinShare') {
        Write-Host "  UI appears to be serving correctly" -ForegroundColor Gray
    }

    return $uiResponse.Success
}

#endregion

#region Integration Tests

function Test-FileUpload {
    Write-TestHeader "Testing File Upload"

    # Create a test file
    $testFile = Join-Path $env:TEMP "pinshare-test-$(Get-Random).txt"
    "Test content $(Get-Date)" | Out-File -FilePath $testFile -Encoding UTF8

    try {
        # TODO: Implement multipart file upload test
        # This requires proper multipart/form-data implementation
        Write-TestResult -TestName "File Upload" -Passed $false -Message "Not yet implemented"

        return $false
    } finally {
        Remove-Item $testFile -ErrorAction SilentlyContinue
    }
}

function Test-ServiceRecovery {
    Write-TestHeader "Testing Service Recovery"

    # This test verifies that the service restarts child processes if they crash
    # For now, we'll skip this complex test
    Write-TestResult -TestName "Service Recovery" -Passed $false -Message "Complex test - manual verification required"

    return $false
}

#endregion

#region Cleanup

function Invoke-Cleanup {
    param([bool]$KeepData = $false)

    Write-TestHeader "Cleanup"

    # Stop service
    Write-Host "Stopping service..." -ForegroundColor Cyan
    $service = Get-Service -Name $script:Config.ServiceName -ErrorAction SilentlyContinue
    if ($service -and $service.Status -eq 'Running') {
        Stop-Service -Name $script:Config.ServiceName -Force
        Start-Sleep -Seconds 3
    }

    # Uninstall service
    Write-Host "Uninstalling service..." -ForegroundColor Cyan
    $serviceExe = Join-Path $PSScriptRoot "..\..\build\pinsharesvc.exe"
    if (Test-Path $serviceExe) {
        & $serviceExe uninstall 2>&1 | Out-Null
        Start-Sleep -Seconds 2
    }

    # Clean up data if requested
    if (-not $KeepData) {
        Write-Host "Removing data directory..." -ForegroundColor Cyan
        if (Test-Path $script:Config.DataDir) {
            Remove-Item $script:Config.DataDir -Recurse -Force -ErrorAction SilentlyContinue
        }
    }

    Write-Host "Cleanup complete" -ForegroundColor Green
}

#endregion

#region Main Execution

function Invoke-TestSuite {
    param([string]$Suite)

    $timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
    $logFile = Join-Path $script:Config.LogPath "test-run-$timestamp.log"

    # Ensure log directory exists
    if (-not (Test-Path $script:Config.LogPath)) {
        New-Item -ItemType Directory -Path $script:Config.LogPath -Force | Out-Null
    }

    # Start transcript
    Start-Transcript -Path $logFile -Append

    try {
        Write-Host "`n" -NoNewline
        Write-Host "╔═══════════════════════════════════════════════════════════════╗" -ForegroundColor Magenta
        Write-Host "║                                                               ║" -ForegroundColor Magenta
        Write-Host "║              PinShare Windows Testing Suite                   ║" -ForegroundColor Magenta
        Write-Host "║                                                               ║" -ForegroundColor Magenta
        Write-Host "╚═══════════════════════════════════════════════════════════════╝" -ForegroundColor Magenta
        Write-Host ""

        Write-Host "Test Suite: $Suite" -ForegroundColor Cyan
        Write-Host "Log File: $logFile" -ForegroundColor Cyan
        Write-Host "Started: $(Get-Date)" -ForegroundColor Cyan
        Write-Host ""

        # Run tests based on suite
        $suiteTests = @{
            'Build' = @('BuildEnvironment', 'BuildService')
            'Service' = @('ServiceInstallation', 'ServiceStart', 'ServiceStop')
            'Health' = @('IPFSHealth', 'PinShareHealth', 'UIServerHealth')
            'API' = @('PinShareHealth')
            'UI' = @('UIServerHealth')
            'Integration' = @('FileUpload', 'ServiceRecovery')
            'All' = @('BuildEnvironment', 'BuildService', 'ServiceInstallation', 'ServiceStart',
                      'IPFSHealth', 'PinShareHealth', 'UIServerHealth', 'ServiceStop')
        }

        $testsToRun = $suiteTests[$Suite]

        foreach ($testName in $testsToRun) {
            & "Test-$testName"
        }

        # Print summary
        Write-TestHeader "Test Summary"

        $total = $script:TestResults.Passed + $script:TestResults.Failed + $script:TestResults.Skipped
        $passRate = if ($total -gt 0) { [math]::Round(($script:TestResults.Passed / $total) * 100, 2) } else { 0 }

        Write-Host "Total Tests: $total" -ForegroundColor Cyan
        Write-Host "Passed: $($script:TestResults.Passed)" -ForegroundColor Green
        Write-Host "Failed: $($script:TestResults.Failed)" -ForegroundColor Red
        Write-Host "Skipped: $($script:TestResults.Skipped)" -ForegroundColor Yellow
        Write-Host "Pass Rate: $passRate%" -ForegroundColor $(if ($passRate -ge 80) { 'Green' } elseif ($passRate -ge 50) { 'Yellow' } else { 'Red' })
        Write-Host ""

        # Export results
        $resultsFile = Join-Path $script:Config.LogPath "test-results-$timestamp.json"
        $script:TestResults | ConvertTo-Json -Depth 5 | Out-File $resultsFile
        Write-Host "Results exported to: $resultsFile" -ForegroundColor Cyan

        # Cleanup if requested
        if ($Suite -in @('All', 'Cleanup')) {
            Invoke-Cleanup -KeepData $KeepData
        }

        return ($script:TestResults.Failed -eq 0)

    } finally {
        Stop-Transcript
        Write-Host "`nLog saved to: $logFile" -ForegroundColor Cyan
    }
}

# Execute
$success = Invoke-TestSuite -Suite $TestSuite
exit $(if ($success) { 0 } else { 1 })

#endregion

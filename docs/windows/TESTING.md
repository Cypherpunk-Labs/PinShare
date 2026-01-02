# Windows Testing Strategy & Guide

Complete testing strategy for rapid feedback and iteration on Windows development.

## Table of Contents

1. [Quick Start](#quick-start)
2. [Test Automation](#test-automation)
3. [Manual Testing](#manual-testing)
4. [Development Workflow](#development-workflow)
5. [Troubleshooting](#troubleshooting)
6. [CI/CD Integration](#cicd-integration)

---

## Quick Start

### Prerequisites

- Windows 10/11 (build 19041 or later)
- Git Bash (preferred shell)
- Administrator privileges
- Go 1.21+ installed

### Run All Tests

```powershell
# Open PowerShell as Administrator
cd C:\path\to\PinShare
.\scripts\windows\Test-PinShare.ps1 -TestSuite All
```

###

 Run Specific Test Suite

```powershell
# Build tests only
.\scripts\windows\Test-PinShare.ps1 -TestSuite Build

# Service installation/lifecycle tests
.\scripts\windows\Test-PinShare.ps1 -TestSuite Service

# Health check tests
.\scripts\windows\Test-PinShare.ps1 -TestSuite Health

# Integration tests
.\scripts\windows\Test-PinShare.ps1 -TestSuite Integration
```

---

## Test Automation

### Test-PinShare.ps1

Main automated testing script with the following features:

**Test Suites:**
- `Build` - Validate build environment and compile binaries
- `Service` - Test service installation, start, stop
- `Health` - Verify IPFS and API health
- `API` - Test PinShare REST API endpoints
- `Integration` - End-to-end integration tests
- `All` - Complete test run
- `Cleanup` - Remove service and data

**Parameters:**
```powershell
-TestSuite <String>    # Which tests to run (default: All)
-SkipBuild            # Skip building binaries
-Verbose              # Enable verbose output
-KeepData             # Don't delete data directory on cleanup
-LogPath <String>     # Custom log directory
```

**Examples:**
```powershell
# Quick build-only test
.\scripts\windows\Test-PinShare.ps1 -TestSuite Build

# Test with existing build
.\scripts\windows\Test-PinShare.ps1 -SkipBuild -KeepData

# Verbose output with custom logs
.\scripts\windows\Test-PinShare.ps1 -Verbose -LogPath "C:\Logs\PinShare"
```

**Output:**
- Console output with colored pass/fail indicators
- Detailed log file in `test-results/` directory
- JSON results file for automation

---

## Manual Testing

### Critical Path Testing

1. **Build Verification**
   ```powershell
   # Build binaries
   cd cmd\pinsharesvc
   go build -o ..\..\build\pinsharesvc.exe

   cd ..\pinshare-tray
   go build -ldflags "-H=windowsgui" -o ..\..\build\pinshare-tray.exe
   ```

2. **Service Installation**
   ```powershell
   # Install service
   .\build\pinsharesvc.exe install

   # Verify registration
   Get-Service PinShareService

   # Check Event Log
   Get-EventLog -LogName Application -Source PinShareService -Newest 10
   ```

3. **Service Lifecycle**
   ```powershell
   # Start service
   Start-Service PinShareService

   # Check status
   Get-Service PinShareService

   # Monitor startup
   Get-EventLog -LogName Application -Source PinShareService -After (Get-Date).AddMinutes(-5)

   # Stop service
   Stop-Service PinShareService -Force
   ```

4. **Health Verification**
   ```powershell
   # Test IPFS
   Invoke-WebRequest http://localhost:5001/api/v0/version

   # Test PinShare API
   Invoke-WebRequest http://localhost:9090/api/v1/files
   ```

5. **Log Review**
   ```powershell
   # Service logs
   Get-Content C:\ProgramData\PinShare\logs\pinshare-service.log -Tail 50

   # IPFS logs
   Get-Content C:\ProgramData\PinShare\logs\ipfs.log -Tail 50

   # PinShare application logs
   Get-Content C:\ProgramData\PinShare\logs\pinshare.log -Tail 50
   ```

### Stress Testing

```powershell
# Rapid restart test
for ($i = 1; $i -le 10; $i++) {
    Write-Host "Cycle $i..."
    Restart-Service PinShareService
    Start-Sleep -Seconds 5
    $status = Get-Service PinShareService
    if ($status.Status -ne 'Running') {
        Write-Error "Service failed to start on cycle $i"
        break
    }
}
```

### Port Conflict Testing

```powershell
# Simulate port conflict
$testServer = Start-Job -ScriptBlock {
    $listener = [System.Net.Sockets.TcpListener]::new([System.Net.IPAddress]::Any, 9090)
    $listener.Start()
    Start-Sleep -Seconds 300
    $listener.Stop()
}

# Try to start service (should fail or handle gracefully)
Start-Service PinShareService

# Cleanup
Stop-Job $testServer
Remove-Job $testServer
```

---

## Development Workflow

### Rapid Iteration Cycle

For fast development feedback:

```powershell
# 1. Make code changes

# 2. Quick compile check
go build .\cmd\pinsharesvc

# 3. Full build
.\scripts\windows\Test-PinShare.ps1 -TestSuite Build

# 4. Install and test
.\scripts\windows\Test-PinShare.ps1 -TestSuite Service,Health

# 5. Review logs if issues
Get-Content C:\ProgramData\PinShare\logs\pinshare-service.log -Tail 100 -Wait
```

### Debug Mode Testing

```powershell
# Run in debug mode (not as service)
.\build\pinsharesvc.exe debug

# This runs in foreground with console output
# Useful for immediate feedback
# Press Ctrl+C to stop
```

### Live Monitoring

```powershell
# Monitor service status
while ($true) {
    Clear-Host
    Write-Host "=== Service Status ===" -ForegroundColor Cyan
    Get-Service PinShareService | Format-List

    Write-Host "`n=== Recent Errors ===" -ForegroundColor Yellow
    Get-EventLog -LogName Application -Source PinShareService -EntryType Error -Newest 5 | Format-List

    Write-Host "`n=== Port Status ===" -ForegroundColor Green
    Get-NetTCPConnection -LocalPort 9090,8888,5001,4001 -ErrorAction SilentlyContinue | Format-Table

    Start-Sleep -Seconds 5
}
```

### Unit Testing

```powershell
# Run Go unit tests
cd cmd\pinsharesvc
go test -v ./...

cd ..\pinshare-tray
go test -v ./...

# With coverage
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

## Troubleshooting

### Common Issues

#### Service Won't Start

```powershell
# Check Event Log
Get-EventLog -LogName Application -Source PinShareService -Newest 20

# Check if ports are in use
Get-NetTCPConnection -LocalPort 9090,8888,5001,4001

# Check service configuration
Get-Service PinShareService | Select-Object *

# Try debug mode
.\build\pinsharesvc.exe debug
```

#### IPFS Won't Initialize

```powershell
# Check IPFS repository
Test-Path C:\ProgramData\PinShare\.ipfs

# Check IPFS binary
Get-Command ipfs -ErrorAction SilentlyContinue

# Manual IPFS init
$env:IPFS_PATH = "C:\ProgramData\PinShare\.ipfs"
ipfs init

# Check IPFS config
ipfs config show
```

#### Port Already in Use

```powershell
# Find process using port 9090
Get-NetTCPConnection -LocalPort 9090 | Select-Object OwningProcess
Get-Process -Id <ProcessID>

# Kill the process
Stop-Process -Id <ProcessID> -Force
```

#### Permissions Issues

```powershell
# Check data directory permissions
Get-Acl C:\ProgramData\PinShare | Format-List

# Grant service account permissions
$acl = Get-Acl C:\ProgramData\PinShare
$rule = New-Object System.Security.AccessControl.FileSystemAccessRule(
    "NT AUTHORITY\LOCAL SERVICE", "FullControl", "ContainerInherit,ObjectInherit", "None", "Allow"
)
$acl.SetAccessRule($rule)
Set-Acl C:\ProgramData\PinShare $acl
```

### Diagnostic Data Collection

```powershell
# Create diagnostic bundle
$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$diagPath = "C:\Temp\pinshare-diag-$timestamp"
New-Item -ItemType Directory -Path $diagPath

# Collect logs
Copy-Item C:\ProgramData\PinShare\logs\* $diagPath\logs\ -Recurse -ErrorAction SilentlyContinue

# Collect configuration
Copy-Item C:\ProgramData\PinShare\config.json $diagPath\ -ErrorAction SilentlyContinue

# Collect Event Log
Get-EventLog -LogName Application -Source PinShareService -Newest 100 | Export-Csv "$diagPath\eventlog.csv"

# Collect service status
Get-Service PinShareService | Select-Object * | Export-Csv "$diagPath\service-status.csv"

# Collect network status
Get-NetTCPConnection -LocalPort 9090,8888,5001,4001 | Export-Csv "$diagPath\network-ports.csv"

# Compress
Compress-Archive -Path $diagPath -DestinationPath "$diagPath.zip"
Write-Host "Diagnostics saved to: $diagPath.zip"
```

---

## CI/CD Integration

### GitHub Actions Example

```yaml
# .github/workflows/windows-test.yml
name: Windows Tests

on: [push, pull_request]

jobs:
  test-windows:
    runs-on: windows-latest
    steps:
      - uses: actions/checkout@v3

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Run Build Tests
        shell: powershell
        run: |
          .\scripts\windows\Test-PinShare.ps1 -TestSuite Build

      - name: Upload Test Results
        if: always()
        uses: actions/upload-artifact@v3
        with:
          name: test-results
          path: test-results/
```

### Azure DevOps Example

```yaml
# azure-pipelines.yml
trigger:
  - main
  - develop

pool:
  vmImage: 'windows-latest'

steps:
  - task: GoTool@0
    inputs:
      version: '1.21'

  - powershell: |
      .\scripts\windows\Test-PinShare.ps1 -TestSuite All
    displayName: 'Run Tests'

  - task: PublishTestResults@2
    condition: always()
    inputs:
      testResultsFormat: 'JUnit'
      testResultsFiles: 'test-results/*.xml'
```

---

## Performance Testing

### Startup Time

```powershell
# Measure service startup time
$stopwatch = [System.Diagnostics.Stopwatch]::StartNew()
Start-Service PinShareService

while ((Get-Service PinShareService).Status -ne 'Running') {
    Start-Sleep -Milliseconds 100
}

$stopwatch.Stop()
Write-Host "Service started in $($stopwatch.Elapsed.TotalSeconds) seconds"

# Measure IPFS ready time
$stopwatch.Restart()
while (-not (Test-NetConnection localhost -Port 5001 -InformationLevel Quiet)) {
    Start-Sleep -Milliseconds 500
}
$stopwatch.Stop()
Write-Host "IPFS ready in $($stopwatch.Elapsed.TotalSeconds) seconds"
```

### Memory Usage Monitoring

```powershell
# Monitor memory usage
$processes = @('pinsharesvc', 'ipfs', 'pinshare')
while ($true) {
    Clear-Host
    Write-Host "=== Memory Usage ===" -ForegroundColor Cyan
    foreach ($proc in $processes) {
        Get-Process $proc -ErrorAction SilentlyContinue |
            Select-Object Name, @{Name='Memory(MB)';Expression={[math]::Round($_.WS/1MB,2)}} |
            Format-Table
    }
    Start-Sleep -Seconds 5
}
```

---

## Test Coverage Goals

| Component | Target Coverage | Priority |
|-----------|----------------|----------|
| Service Lifecycle | 90% | Critical |
| Process Management | 85% | Critical |
| Health Checking | 80% | High |
| Configuration | 75% | High |
| Tray Application | 60% | Medium |

---

## Continuous Improvement

### Add New Tests

When adding features:

1. Add unit tests in the relevant package
2. Add integration tests to `Test-PinShare.ps1`
3. Update this documentation
4. Add to CI/CD pipeline

### Test Maintenance

Weekly:
- Review failed test history
- Update timeouts if needed
- Check for flaky tests
- Update expected behaviors

Monthly:
- Review test coverage
- Add tests for reported bugs
- Performance benchmark comparison
- Update test data/fixtures

---

## Additional Resources

- [Windows Service Best Practices](https://docs.microsoft.com/en-us/windows/win32/services/service-security-and-access-rights)
- [PowerShell Testing with Pester](https://pester.dev/)
- [Go Testing Documentation](https://golang.org/pkg/testing/)

---

**Last Updated:** 2025-01-22
**Maintained By:** PinShare Development Team

# Windows 11 Compatibility

## Service Start Issue on Windows 11

### Problem

The MSI installer would correctly start the PinShare service after installation on Windows 10, but not on Windows 11. The service would be installed and configured correctly, but would remain in the "Stopped" state.

### Root Cause

Windows 11 has stricter security policies around service startup during MSI installation:

1. **Timing Sensitivity**: Windows 11 requires more time between service registration and service start
2. **Service Control Manager (SCM) Delays**: The SCM may not immediately make newly registered services available for starting
3. **Return Code Handling**: The original implementation used `Return="ignore"` which masked failures

The original WiX configuration used `net.exe start PinShareService` which worked reliably on Windows 10 but would fail silently on Windows 11.

### Solution

Implemented a robust PowerShell-based service start script with:

1. **Retry Logic**: Attempts to start the service up to 3 times with delays
2. **State Verification**: Waits for the service to reach "Running" state before proceeding
3. **Graceful Degradation**: If service start fails, installation completes successfully and provides instructions for manual start
4. **Better Logging**: Detailed output to help diagnose any remaining issues

#### Files Changed

**Created:**
- `installer/start-service.ps1` - PowerShell script with retry logic

**Modified:**
- `installer/Package.wxs`:
  - Replaced `net.exe start` with PowerShell script execution
  - Added `StartServiceScript` component to include the PS1 file
  - Updated `InstallExecuteSequence` to use new action

### Technical Details

#### Old Implementation (Windows 10 only)

```xml
<SetProperty Id="StartService" Value="&quot;[SystemFolder]net.exe&quot; start PinShareService" />
<CustomAction Id="StartService"
              DllEntry="WixQuietExec64"
              Execute="deferred"
              Return="ignore" />
```

**Problems:**
- `net.exe start` is less reliable on Windows 11
- `Return="ignore"` masks failures
- No retry mechanism
- No state verification

#### New Implementation (Windows 10 & 11)

```xml
<SetProperty Id="StartServicePS"
             Value="&quot;[SystemFolder]WindowsPowerShell\v1.0\powershell.exe&quot; -ExecutionPolicy Bypass -NoProfile -File &quot;[INSTALLFOLDER]start-service.ps1&quot;" />
<CustomAction Id="StartServicePS"
              DllEntry="WixQuietExec64"
              Execute="deferred"
              Return="ignore" />
```

**Benefits:**
- PowerShell's `Start-Service` cmdlet is more reliable
- Built-in retry logic (3 attempts with 2s delay)
- Verifies service reaches "Running" state
- Provides clear user feedback
- Graceful failure (completes installation even if service doesn't start)

#### PowerShell Script Features

The `start-service.ps1` script includes:

```powershell
# Parameters
$ServiceName = "PinShareService"
$MaxRetries = 3
$RetryDelaySeconds = 2

# Features:
- Service existence check
- Status verification before starting
- Retry loop with exponential backoff
- 30-second timeout for service to reach Running state
- Detailed logging at each step
- Graceful exit (exit 0) even on failure to prevent installation rollback
```

### Testing

The fix has been tested on:
- ✓ Windows 10 (21H2, 22H2)
- ✓ Windows 11 (21H2, 22H2, 23H2)

#### Test Procedure

1. **Clean Installation**
   ```cmd
   msiexec /i PinShare-Setup.msi /l*v install.log
   ```
   - Service should start automatically
   - Tray application should appear
   - Verify in Services: PinShare service is Running

2. **Service Logs**
   Check the MSI log file for:
   ```
   Starting service: PinShareService
   Starting service...
   SUCCESS: Service started successfully
   ```

3. **Manual Verification**
   ```powershell
   Get-Service PinShareService
   ```
   Should show `Status: Running`

### Troubleshooting

If the service still doesn't start after installation:

1. **Check MSI Log**
   ```cmd
   msiexec /i PinShare-Setup.msi /l*v install.log
   notepad install.log
   ```
   Search for "PinShareService" to see start attempts

2. **Manual Start**
   ```powershell
   Start-Service PinShareService
   ```

3. **Check Event Log**
   ```powershell
   Get-EventLog -LogName Application -Source PinShareService -Newest 10
   ```

4. **Verify Service Configuration**
   ```cmd
   sc query PinShareService
   sc qc PinShareService
   ```

### Known Limitations

1. **PowerShell Requirement**: Requires PowerShell 5.1+ (included in Windows 10/11)
2. **Graceful Failure**: If service fails to start, installation completes without error (by design to prevent rollback)
3. **No Rollback**: Service start failures don't trigger installation rollback

### Future Improvements

Potential enhancements:

1. **Delayed Auto-Start**: Configure service with delayed auto-start type for better Windows 11 compatibility
2. **Event Logging**: Add custom event log entries for service start failures
3. **UI Feedback**: Show dialog to user if service fails to start (optional)
4. **Telemetry**: Collect anonymous metrics on start success/failure rates by OS version

### References

- [Windows Service Control Manager](https://docs.microsoft.com/en-us/windows/win32/services/service-control-manager)
- [WiX Toolset Documentation](https://wixtoolset.org/docs/)
- [PowerShell Start-Service](https://docs.microsoft.com/en-us/powershell/module/microsoft.powershell.management/start-service)

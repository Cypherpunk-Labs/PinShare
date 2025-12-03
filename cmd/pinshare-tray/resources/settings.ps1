# PinShare Settings Dialog
# Uses WinForms for native Windows UI
# Handles UAC elevation internally for registry writes

param(
    [switch]$Save,
    [string]$ConfigJson
)

Add-Type -AssemblyName System.Windows.Forms
Add-Type -AssemblyName System.Drawing

$registryPath = "HKLM:\SOFTWARE\PinShare"

# Check if running elevated
function Test-Elevated {
    $identity = [Security.Principal.WindowsIdentity]::GetCurrent()
    $principal = New-Object Security.Principal.WindowsPrincipal($identity)
    return $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

# If called with -Save and elevated, write config and exit
if ($Save -and (Test-Elevated)) {
    try {
        $config = $ConfigJson | ConvertFrom-Json

        # Ensure registry key exists
        if (-not (Test-Path $registryPath)) {
            New-Item -Path $registryPath -Force | Out-Null
        }

        # Write string values
        Set-ItemProperty -Path $registryPath -Name "OrgName" -Value $config.OrgName
        Set-ItemProperty -Path $registryPath -Name "GroupName" -Value $config.GroupName
        Set-ItemProperty -Path $registryPath -Name "VirusTotalToken" -Value $config.VirusTotalToken
        Set-ItemProperty -Path $registryPath -Name "LogLevel" -Value $config.LogLevel

        # Write integer values (ports)
        Set-ItemProperty -Path $registryPath -Name "IPFSAPIPort" -Value $config.IPFSAPIPort -Type DWord
        Set-ItemProperty -Path $registryPath -Name "IPFSGatewayPort" -Value $config.IPFSGatewayPort -Type DWord
        Set-ItemProperty -Path $registryPath -Name "IPFSSwarmPort" -Value $config.IPFSSwarmPort -Type DWord
        Set-ItemProperty -Path $registryPath -Name "PinShareAPIPort" -Value $config.PinShareAPIPort -Type DWord
        Set-ItemProperty -Path $registryPath -Name "PinShareP2PPort" -Value $config.PinShareP2PPort -Type DWord
        Set-ItemProperty -Path $registryPath -Name "UIPort" -Value $config.UIPort -Type DWord

        # Write boolean values as DWord (0/1)
        Set-ItemProperty -Path $registryPath -Name "SkipVirusTotal" -Value ([int]$config.SkipVirusTotal) -Type DWord
        Set-ItemProperty -Path $registryPath -Name "EnableCache" -Value ([int]$config.EnableCache) -Type DWord
        Set-ItemProperty -Path $registryPath -Name "ArchiveNode" -Value ([int]$config.ArchiveNode) -Type DWord

        exit 0
    } catch {
        [System.Windows.Forms.MessageBox]::Show(
            "Failed to save settings: $($_.Exception.Message)",
            "Error",
            [System.Windows.Forms.MessageBoxButtons]::OK,
            [System.Windows.Forms.MessageBoxIcon]::Error)
        exit 2
    }
}

# Read current config from registry
function Read-Config {
    $config = @{
        IPFSAPIPort = 5001
        IPFSGatewayPort = 8080
        IPFSSwarmPort = 4001
        PinShareAPIPort = 9090
        PinShareP2PPort = 50001
        UIPort = 8888
        OrgName = "MyOrganization"
        GroupName = "MyGroup"
        SkipVirusTotal = $false
        EnableCache = $true
        ArchiveNode = $false
        VirusTotalToken = ""
        LogLevel = "info"
        InstallDirectory = "C:\Program Files\PinShare"
        DataDirectory = "C:\ProgramData\PinShare"
    }

    if (Test-Path $registryPath) {
        try {
            $key = Get-Item $registryPath -ErrorAction SilentlyContinue
            if ($key) {
                # Read integer values (ports)
                foreach ($prop in @("IPFSAPIPort","IPFSGatewayPort","IPFSSwarmPort",
                                   "PinShareAPIPort","PinShareP2PPort","UIPort")) {
                    $val = $key.GetValue($prop)
                    if ($null -ne $val) { $config[$prop] = [int]$val }
                }

                # Read string values
                foreach ($prop in @("OrgName","GroupName","VirusTotalToken","LogLevel",
                                   "InstallDirectory","DataDirectory")) {
                    $val = $key.GetValue($prop)
                    if ($null -ne $val -and $val -ne "") { $config[$prop] = $val }
                }

                # Read boolean values (stored as DWord 0/1)
                foreach ($prop in @("SkipVirusTotal","EnableCache","ArchiveNode")) {
                    $val = $key.GetValue($prop)
                    if ($null -ne $val) { $config[$prop] = [bool][int]$val }
                }
            }
        } catch {
            # Silently use defaults if registry read fails
        }
    }

    return $config
}

# Load current config
$config = Read-Config

# Create main form
$form = New-Object System.Windows.Forms.Form
$form.Text = "PinShare Settings"
$form.Size = New-Object System.Drawing.Size(500, 480)
$form.StartPosition = "CenterScreen"
$form.FormBorderStyle = "FixedDialog"
$form.MaximizeBox = $false
$form.MinimizeBox = $false

# Create TabControl
$tabControl = New-Object System.Windows.Forms.TabControl
$tabControl.Location = New-Object System.Drawing.Point(10, 10)
$tabControl.Size = New-Object System.Drawing.Size(465, 380)

# ==================== Tab 1: Network Ports ====================
$tabPorts = New-Object System.Windows.Forms.TabPage
$tabPorts.Text = "Network Ports"
$tabPorts.Padding = New-Object System.Windows.Forms.Padding(10)

$y = 20
$portFields = @{}

$portDefs = @(
    @{Name="IPFSAPIPort"; Label="IPFS API Port:"; Hint="(default: 5001)"},
    @{Name="IPFSGatewayPort"; Label="IPFS Gateway Port:"; Hint="(default: 8080)"},
    @{Name="IPFSSwarmPort"; Label="IPFS Swarm Port:"; Hint="(default: 4001)"},
    @{Name="PinShareAPIPort"; Label="PinShare API Port:"; Hint="(default: 9090)"},
    @{Name="PinShareP2PPort"; Label="PinShare P2P Port:"; Hint="(default: 50001)"},
    @{Name="UIPort"; Label="UI Port:"; Hint="(default: 8888)"}
)

foreach ($port in $portDefs) {
    $lbl = New-Object System.Windows.Forms.Label
    $lbl.Location = New-Object System.Drawing.Point(10, $y)
    $lbl.Size = New-Object System.Drawing.Size(140, 20)
    $lbl.Text = $port.Label
    $tabPorts.Controls.Add($lbl)

    $txt = New-Object System.Windows.Forms.TextBox
    $txt.Location = New-Object System.Drawing.Point(160, ($y - 3))
    $txt.Size = New-Object System.Drawing.Size(80, 20)
    $txt.Text = $config[$port.Name]
    $txt.MaxLength = 5
    $tabPorts.Controls.Add($txt)
    $portFields[$port.Name] = $txt

    $hint = New-Object System.Windows.Forms.Label
    $hint.Location = New-Object System.Drawing.Point(250, $y)
    $hint.Size = New-Object System.Drawing.Size(150, 20)
    $hint.Text = $port.Hint
    $hint.ForeColor = [System.Drawing.Color]::Gray
    $tabPorts.Controls.Add($hint)

    $y += 40
}

# Warning label
$lblWarning = New-Object System.Windows.Forms.Label
$lblWarning.Location = New-Object System.Drawing.Point(10, 270)
$lblWarning.Size = New-Object System.Drawing.Size(420, 40)
$lblWarning.Text = "Warning: Changing ports requires a service restart. Ensure the new ports are not in use by other applications."
$lblWarning.ForeColor = [System.Drawing.Color]::DarkOrange
$tabPorts.Controls.Add($lblWarning)

$tabControl.Controls.Add($tabPorts)

# ==================== Tab 2: Organization ====================
$tabOrg = New-Object System.Windows.Forms.TabPage
$tabOrg.Text = "Organization"

$lblOrg = New-Object System.Windows.Forms.Label
$lblOrg.Location = New-Object System.Drawing.Point(10, 20)
$lblOrg.Size = New-Object System.Drawing.Size(130, 20)
$lblOrg.Text = "Organization Name:"
$tabOrg.Controls.Add($lblOrg)

$txtOrgName = New-Object System.Windows.Forms.TextBox
$txtOrgName.Location = New-Object System.Drawing.Point(150, 17)
$txtOrgName.Size = New-Object System.Drawing.Size(280, 20)
$txtOrgName.Text = $config.OrgName
$tabOrg.Controls.Add($txtOrgName)

$lblGroup = New-Object System.Windows.Forms.Label
$lblGroup.Location = New-Object System.Drawing.Point(10, 55)
$lblGroup.Size = New-Object System.Drawing.Size(130, 20)
$lblGroup.Text = "Group Name:"
$tabOrg.Controls.Add($lblGroup)

$txtGroupName = New-Object System.Windows.Forms.TextBox
$txtGroupName.Location = New-Object System.Drawing.Point(150, 52)
$txtGroupName.Size = New-Object System.Drawing.Size(280, 20)
$txtGroupName.Text = $config.GroupName
$tabOrg.Controls.Add($txtGroupName)

$lblNote = New-Object System.Windows.Forms.Label
$lblNote.Location = New-Object System.Drawing.Point(10, 100)
$lblNote.Size = New-Object System.Drawing.Size(420, 60)
$lblNote.Text = "Organization and Group form the gossip topic for peer discovery.`n`nAll PinShare nodes with the same Organization and Group will automatically discover and connect to each other."
$tabOrg.Controls.Add($lblNote)

$tabControl.Controls.Add($tabOrg)

# ==================== Tab 3: Features ====================
$tabFeatures = New-Object System.Windows.Forms.TabPage
$tabFeatures.Text = "Features"

$chkSkipVT = New-Object System.Windows.Forms.CheckBox
$chkSkipVT.Location = New-Object System.Drawing.Point(10, 20)
$chkSkipVT.Size = New-Object System.Drawing.Size(420, 20)
$chkSkipVT.Text = "Skip VirusTotal scanning"
$chkSkipVT.Checked = $config.SkipVirusTotal
$tabFeatures.Controls.Add($chkSkipVT)

$lblSkipVTDesc = New-Object System.Windows.Forms.Label
$lblSkipVTDesc.Location = New-Object System.Drawing.Point(30, 42)
$lblSkipVTDesc.Size = New-Object System.Drawing.Size(400, 20)
$lblSkipVTDesc.Text = "Disable virus scanning for uploaded files"
$lblSkipVTDesc.ForeColor = [System.Drawing.Color]::Gray
$tabFeatures.Controls.Add($lblSkipVTDesc)

$chkCache = New-Object System.Windows.Forms.CheckBox
$chkCache.Location = New-Object System.Drawing.Point(10, 75)
$chkCache.Size = New-Object System.Drawing.Size(420, 20)
$chkCache.Text = "Enable file caching"
$chkCache.Checked = $config.EnableCache
$tabFeatures.Controls.Add($chkCache)

$lblCacheDesc = New-Object System.Windows.Forms.Label
$lblCacheDesc.Location = New-Object System.Drawing.Point(30, 97)
$lblCacheDesc.Size = New-Object System.Drawing.Size(400, 20)
$lblCacheDesc.Text = "Cache downloaded files locally for faster access"
$lblCacheDesc.ForeColor = [System.Drawing.Color]::Gray
$tabFeatures.Controls.Add($lblCacheDesc)

$chkArchive = New-Object System.Windows.Forms.CheckBox
$chkArchive.Location = New-Object System.Drawing.Point(10, 130)
$chkArchive.Size = New-Object System.Drawing.Size(420, 20)
$chkArchive.Text = "Archive node mode"
$chkArchive.Checked = $config.ArchiveNode
$tabFeatures.Controls.Add($chkArchive)

$lblArchiveDesc = New-Object System.Windows.Forms.Label
$lblArchiveDesc.Location = New-Object System.Drawing.Point(30, 152)
$lblArchiveDesc.Size = New-Object System.Drawing.Size(400, 35)
$lblArchiveDesc.Text = "Pin all content shared by network peers (requires significant disk space)"
$lblArchiveDesc.ForeColor = [System.Drawing.Color]::Gray
$tabFeatures.Controls.Add($lblArchiveDesc)

$tabControl.Controls.Add($tabFeatures)

# ==================== Tab 4: Security ====================
$tabSecurity = New-Object System.Windows.Forms.TabPage
$tabSecurity.Text = "Security"

$lblToken = New-Object System.Windows.Forms.Label
$lblToken.Location = New-Object System.Drawing.Point(10, 20)
$lblToken.Size = New-Object System.Drawing.Size(130, 20)
$lblToken.Text = "VirusTotal API Token:"
$tabSecurity.Controls.Add($lblToken)

$txtToken = New-Object System.Windows.Forms.TextBox
$txtToken.Location = New-Object System.Drawing.Point(150, 17)
$txtToken.Size = New-Object System.Drawing.Size(280, 20)
$txtToken.Text = $config.VirusTotalToken
$txtToken.UseSystemPasswordChar = $true
$tabSecurity.Controls.Add($txtToken)

$lblTokenDesc = New-Object System.Windows.Forms.Label
$lblTokenDesc.Location = New-Object System.Drawing.Point(10, 45)
$lblTokenDesc.Size = New-Object System.Drawing.Size(420, 20)
$lblTokenDesc.Text = "Get a free API key from virustotal.com"
$lblTokenDesc.ForeColor = [System.Drawing.Color]::Gray
$tabSecurity.Controls.Add($lblTokenDesc)

$lblLogLevel = New-Object System.Windows.Forms.Label
$lblLogLevel.Location = New-Object System.Drawing.Point(10, 85)
$lblLogLevel.Size = New-Object System.Drawing.Size(130, 20)
$lblLogLevel.Text = "Log Level:"
$tabSecurity.Controls.Add($lblLogLevel)

$cmbLogLevel = New-Object System.Windows.Forms.ComboBox
$cmbLogLevel.Location = New-Object System.Drawing.Point(150, 82)
$cmbLogLevel.Size = New-Object System.Drawing.Size(120, 20)
$cmbLogLevel.DropDownStyle = "DropDownList"
$cmbLogLevel.Items.AddRange(@("debug", "info", "warn", "error"))
$idx = $cmbLogLevel.Items.IndexOf($config.LogLevel)
if ($idx -ge 0) { $cmbLogLevel.SelectedIndex = $idx } else { $cmbLogLevel.SelectedIndex = 1 }
$tabSecurity.Controls.Add($cmbLogLevel)

$lblLogDesc = New-Object System.Windows.Forms.Label
$lblLogDesc.Location = New-Object System.Drawing.Point(10, 115)
$lblLogDesc.Size = New-Object System.Drawing.Size(420, 20)
$lblLogDesc.Text = "debug = verbose, info = normal, warn/error = minimal"
$lblLogDesc.ForeColor = [System.Drawing.Color]::Gray
$tabSecurity.Controls.Add($lblLogDesc)

$tabControl.Controls.Add($tabSecurity)

# ==================== Tab 5: Info (Read-Only) ====================
$tabInfo = New-Object System.Windows.Forms.TabPage
$tabInfo.Text = "Info"

$lblInstallDir = New-Object System.Windows.Forms.Label
$lblInstallDir.Location = New-Object System.Drawing.Point(10, 20)
$lblInstallDir.Size = New-Object System.Drawing.Size(110, 20)
$lblInstallDir.Text = "Install Directory:"
$tabInfo.Controls.Add($lblInstallDir)

$txtInstallDir = New-Object System.Windows.Forms.TextBox
$txtInstallDir.Location = New-Object System.Drawing.Point(130, 17)
$txtInstallDir.Size = New-Object System.Drawing.Size(300, 20)
$txtInstallDir.Text = $config.InstallDirectory
$txtInstallDir.ReadOnly = $true
$txtInstallDir.BackColor = [System.Drawing.SystemColors]::Control
$tabInfo.Controls.Add($txtInstallDir)

$lblDataDir = New-Object System.Windows.Forms.Label
$lblDataDir.Location = New-Object System.Drawing.Point(10, 55)
$lblDataDir.Size = New-Object System.Drawing.Size(110, 20)
$lblDataDir.Text = "Data Directory:"
$tabInfo.Controls.Add($lblDataDir)

$txtDataDir = New-Object System.Windows.Forms.TextBox
$txtDataDir.Location = New-Object System.Drawing.Point(130, 52)
$txtDataDir.Size = New-Object System.Drawing.Size(300, 20)
$txtDataDir.Text = $config.DataDirectory
$txtDataDir.ReadOnly = $true
$txtDataDir.BackColor = [System.Drawing.SystemColors]::Control
$tabInfo.Controls.Add($txtDataDir)

$lblInfoNote = New-Object System.Windows.Forms.Label
$lblInfoNote.Location = New-Object System.Drawing.Point(10, 100)
$lblInfoNote.Size = New-Object System.Drawing.Size(420, 40)
$lblInfoNote.Text = "These paths are set during installation and cannot be changed here."
$lblInfoNote.ForeColor = [System.Drawing.Color]::Gray
$tabInfo.Controls.Add($lblInfoNote)

$tabControl.Controls.Add($tabInfo)

$form.Controls.Add($tabControl)

# ==================== Buttons ====================
$btnSave = New-Object System.Windows.Forms.Button
$btnSave.Location = New-Object System.Drawing.Point(290, 405)
$btnSave.Size = New-Object System.Drawing.Size(85, 28)
$btnSave.Text = "Save"

$btnCancel = New-Object System.Windows.Forms.Button
$btnCancel.Location = New-Object System.Drawing.Point(385, 405)
$btnCancel.Size = New-Object System.Drawing.Size(85, 28)
$btnCancel.Text = "Cancel"
$btnCancel.Add_Click({ $form.Close() })

$form.Controls.Add($btnSave)
$form.Controls.Add($btnCancel)
$form.AcceptButton = $btnSave
$form.CancelButton = $btnCancel

# Validation function
function Validate-Port($value, $name) {
    $port = 0
    if (-not [int]::TryParse($value, [ref]$port)) {
        return "$name must be a number"
    }
    if ($port -lt 1 -or $port -gt 65535) {
        return "$name must be between 1 and 65535"
    }
    return $null
}

# Track if settings were saved
$script:settingsChanged = $false

# Save button handler
$btnSave.Add_Click({
    # Validate all ports
    foreach ($port in @("IPFSAPIPort","IPFSGatewayPort","IPFSSwarmPort",
                        "PinShareAPIPort","PinShareP2PPort","UIPort")) {
        $err = Validate-Port $portFields[$port].Text $port
        if ($err) {
            [System.Windows.Forms.MessageBox]::Show($err, "Validation Error",
                [System.Windows.Forms.MessageBoxButtons]::OK,
                [System.Windows.Forms.MessageBoxIcon]::Warning)
            $tabControl.SelectedTab = $tabPorts
            $portFields[$port].Focus()
            return
        }
    }

    # Validate org/group names are not empty
    if ([string]::IsNullOrWhiteSpace($txtOrgName.Text)) {
        [System.Windows.Forms.MessageBox]::Show("Organization Name cannot be empty", "Validation Error",
            [System.Windows.Forms.MessageBoxButtons]::OK,
            [System.Windows.Forms.MessageBoxIcon]::Warning)
        $tabControl.SelectedTab = $tabOrg
        $txtOrgName.Focus()
        return
    }

    if ([string]::IsNullOrWhiteSpace($txtGroupName.Text)) {
        [System.Windows.Forms.MessageBox]::Show("Group Name cannot be empty", "Validation Error",
            [System.Windows.Forms.MessageBoxButtons]::OK,
            [System.Windows.Forms.MessageBoxIcon]::Warning)
        $tabControl.SelectedTab = $tabOrg
        $txtGroupName.Focus()
        return
    }

    # Collect all settings
    $newConfig = @{
        IPFSAPIPort = [int]$portFields["IPFSAPIPort"].Text
        IPFSGatewayPort = [int]$portFields["IPFSGatewayPort"].Text
        IPFSSwarmPort = [int]$portFields["IPFSSwarmPort"].Text
        PinShareAPIPort = [int]$portFields["PinShareAPIPort"].Text
        PinShareP2PPort = [int]$portFields["PinShareP2PPort"].Text
        UIPort = [int]$portFields["UIPort"].Text
        OrgName = $txtOrgName.Text.Trim()
        GroupName = $txtGroupName.Text.Trim()
        SkipVirusTotal = $chkSkipVT.Checked
        EnableCache = $chkCache.Checked
        ArchiveNode = $chkArchive.Checked
        VirusTotalToken = $txtToken.Text
        LogLevel = $cmbLogLevel.SelectedItem.ToString()
    }

    # Convert to JSON with proper escaping
    $json = ($newConfig | ConvertTo-Json -Compress) -replace "'", "''"

    # Get path to this script
    $scriptPath = $MyInvocation.MyCommand.Definition
    if (-not $scriptPath) {
        $scriptPath = $PSCommandPath
    }

    # Re-launch elevated to save
    $psi = New-Object System.Diagnostics.ProcessStartInfo
    $psi.FileName = "powershell.exe"
    $psi.Arguments = "-ExecutionPolicy Bypass -NoProfile -WindowStyle Hidden -File `"$scriptPath`" -Save -ConfigJson '$json'"
    $psi.Verb = "runas"
    $psi.UseShellExecute = $true

    try {
        $proc = [System.Diagnostics.Process]::Start($psi)
        $proc.WaitForExit()

        if ($proc.ExitCode -eq 0) {
            $script:settingsChanged = $true
            $form.Close()
        } elseif ($proc.ExitCode -eq 2) {
            # Error was already shown by the elevated process
        }
    } catch [System.ComponentModel.Win32Exception] {
        # User cancelled UAC prompt
        [System.Windows.Forms.MessageBox]::Show(
            "Settings were not saved. Administrator privileges are required to modify system settings.",
            "Cancelled",
            [System.Windows.Forms.MessageBoxButtons]::OK,
            [System.Windows.Forms.MessageBoxIcon]::Information)
    } catch {
        [System.Windows.Forms.MessageBox]::Show(
            "Failed to save settings: $($_.Exception.Message)",
            "Error",
            [System.Windows.Forms.MessageBoxButtons]::OK,
            [System.Windows.Forms.MessageBoxIcon]::Error)
    }
})

# Show the form
[void]$form.ShowDialog()

# Exit with appropriate code
if ($script:settingsChanged) {
    exit 0  # Settings saved successfully
} else {
    exit 1  # User cancelled
}

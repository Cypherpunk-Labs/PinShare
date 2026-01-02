package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"pinshare/internal/winservice"

	"golang.org/x/sys/windows/svc/debug"
)

// Process management constants
const (
	// orphanCleanupDelay is the time to wait after killing orphaned processes
	// before attempting to remove lock files. Windows can take time to release handles.
	orphanCleanupDelay = 2 * time.Second

	// lockFileRemovalRetryDelay is the delay between lock file removal attempts
	lockFileRemovalRetryDelay = 1 * time.Second

	// lockFileRemovalMaxAttempts is the maximum number of attempts to remove a lock file
	lockFileRemovalMaxAttempts = 3

	// Binary names for process management
	ipfsBinaryName     = "ipfs.exe"
	pinShareBinaryName = "pinshare.exe"

	// Lock file name
	ipfsLockFileName = "repo.lock"
)

type ProcessManager struct {
	config   *ServiceConfig
	eventLog debug.Log

	// processMu protects the process state fields below
	processMu       sync.Mutex
	ipfsCmd         *exec.Cmd
	pinshareCmd     *exec.Cmd
	ipfsLogFile     *os.File
	pinshareLogFile *os.File
	ipfsExited      chan struct{} // closed when IPFS process exits
	pinshareExited  chan struct{} // closed when PinShare process exits
}

func NewProcessManager(config *ServiceConfig, eventLog debug.Log) *ProcessManager {
	return &ProcessManager{
		config:   config,
		eventLog: eventLog,
	}
}

// CleanupOrphanedProcesses kills any orphaned IPFS or PinShare processes and removes stale lock files.
// This should be called before starting the service to ensure a clean state.
func (pm *ProcessManager) CleanupOrphanedProcesses() {
	pm.logInfo("Checking for orphaned processes...")

	// Kill any orphaned processes
	pm.killOrphanedProcess(ipfsBinaryName, "IPFS")
	pm.killOrphanedProcess(pinShareBinaryName, winservice.AppName)

	// Wait for processes to fully terminate and file handles to be released
	time.Sleep(orphanCleanupDelay)

	// Remove stale IPFS lock file if it exists
	pm.removeStaleLockFile()

	pm.logInfo("Orphaned process cleanup complete")
}

// killProcessByPID kills a process and its children using taskkill.
// If taskkill fails, it falls back to process.Kill().
func (pm *ProcessManager) killProcessByPID(pid int, name string) {
	killCmd := exec.Command("taskkill", "/F", "/T", "/PID", fmt.Sprintf("%d", pid))
	if output, err := killCmd.CombinedOutput(); err != nil {
		pm.logError(fmt.Sprintf("taskkill failed for %s (PID %d): %s", name, pid, string(output)), err)
	}
}

// removeStaleLockFile attempts to remove the IPFS lock file with retries.
func (pm *ProcessManager) removeStaleLockFile() {
	ipfsLockFile := filepath.Join(pm.config.GetIPFSDataPath(), ipfsLockFileName)
	if _, err := os.Stat(ipfsLockFile); err != nil {
		return // Lock file doesn't exist
	}

	pm.logInfo(fmt.Sprintf("Removing stale IPFS lock file: %s", ipfsLockFile))

	for attempt := 1; attempt <= lockFileRemovalMaxAttempts; attempt++ {
		if err := os.Remove(ipfsLockFile); err != nil {
			if attempt < lockFileRemovalMaxAttempts {
				pm.logInfo(fmt.Sprintf("Lock file removal attempt %d failed, retrying...", attempt))
				time.Sleep(lockFileRemovalRetryDelay)
			} else {
				pm.logError(fmt.Sprintf("Failed to remove IPFS lock file after %d attempts", lockFileRemovalMaxAttempts), err)
			}
		} else {
			pm.logInfo("IPFS lock file removed successfully")
			return
		}
	}
}

// killOrphanedProcess finds and kills any running instances of a process by name
func (pm *ProcessManager) killOrphanedProcess(processName, displayName string) {
	// Use tasklist to check if the process is running
	checkCmd := exec.Command("tasklist", "/FI", fmt.Sprintf("IMAGENAME eq %s", processName), "/NH", "/FO", "CSV")
	output, err := checkCmd.Output()
	if err != nil {
		// tasklist failed, skip this check
		return
	}

	// Check if the process is in the output (CSV format: "process.exe","PID",...)
	outputStr := string(output)
	if !strings.Contains(outputStr, processName) {
		// Process not running
		return
	}

	pm.logInfo(fmt.Sprintf("Found orphaned %s process, terminating...", displayName))

	// Kill all instances of the process using taskkill
	// /F = Force, /T = Tree (kill child processes), /IM = Image name
	killCmd := exec.Command("taskkill", "/F", "/T", "/IM", processName)
	if output, err := killCmd.CombinedOutput(); err != nil {
		pm.logError(fmt.Sprintf("Failed to kill orphaned %s: %s", displayName, string(output)), err)
	} else {
		pm.logInfo(fmt.Sprintf("Orphaned %s process terminated", displayName))
	}
}


// StartIPFS starts the IPFS daemon
func (pm *ProcessManager) StartIPFS(ctx context.Context) error {
	pm.processMu.Lock()
	defer pm.processMu.Unlock()

	// Check if IPFS binary exists
	if _, err := os.Stat(pm.config.IPFSBinary); os.IsNotExist(err) {
		return fmt.Errorf("IPFS binary not found at %s", pm.config.IPFSBinary)
	}

	// Initialize IPFS data directory if it doesn't exist
	ipfsDataPath := pm.config.GetIPFSDataPath()
	if _, err := os.Stat(filepath.Join(ipfsDataPath, "config")); os.IsNotExist(err) {
		pm.logInfo("Initializing IPFS...")
		if err := pm.initializeIPFS(); err != nil {
			return fmt.Errorf("failed to initialize IPFS: %w", err)
		}
	} else {
		// IPFS already initialized - ensure ports match config.json
		pm.logInfo("Syncing IPFS configuration with service config...")
		if err := pm.configureIPFS(); err != nil {
			pm.logError("Failed to sync IPFS config", err)
			// Continue anyway - IPFS may still work with old ports
		}
	}

	// Open log file
	logPath := filepath.Join(pm.config.DataDirectory, winservice.DirLogs, "ipfs.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open IPFS log file: %w", err)
	}
	pm.ipfsLogFile = logFile

	// Create command
	pm.ipfsCmd = exec.CommandContext(ctx, pm.config.IPFSBinary, "daemon")
	pm.ipfsCmd.Env = append(os.Environ(),
		fmt.Sprintf("IPFS_PATH=%s", ipfsDataPath),
	)
	pm.ipfsCmd.Stdout = logFile
	pm.ipfsCmd.Stderr = logFile

	// Set process group for proper cleanup on Windows
	pm.ipfsCmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}

	// Start the process
	if err := pm.ipfsCmd.Start(); err != nil {
		logFile.Close()
		return fmt.Errorf("failed to start IPFS daemon: %w", err)
	}

	pm.logInfo(fmt.Sprintf("IPFS daemon started with PID %d", pm.ipfsCmd.Process.Pid))

	// Create exit channel and monitor process in background
	pm.ipfsExited = make(chan struct{})
	go pm.monitorProcess(ctx, pm.ipfsCmd, "IPFS", pm.ipfsExited)

	return nil
}

// initializeIPFS initializes IPFS in the data directory
func (pm *ProcessManager) initializeIPFS() error {
	ipfsDataPath := pm.config.GetIPFSDataPath()

	cmd := exec.Command(pm.config.IPFSBinary, "init")
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("IPFS_PATH=%s", ipfsDataPath),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ipfs init failed: %w\nOutput: %s", err, string(output))
	}

	pm.logInfo("IPFS initialized successfully")

	// Configure IPFS settings
	if err := pm.configureIPFS(); err != nil {
		return fmt.Errorf("failed to configure IPFS: %w", err)
	}

	return nil
}

// configureIPFS configures IPFS settings from PinShare config.json.
// This MAY ONLY be called when IPFS is NOT running (no repo.lock held).
//
// TODO: Add support for configuring which network interface/IP version to bind to.
// See: https://github.com/Episk-pos/PinShare/issues/10
func (pm *ProcessManager) configureIPFS() error {
	ipfsDataPath := pm.config.GetIPFSDataPath()
	env := append(os.Environ(), fmt.Sprintf("IPFS_PATH=%s", ipfsDataPath))

	pm.logInfo(fmt.Sprintf("Configuring IPFS ports: API=%d, Gateway=%d, Swarm=%d",
		pm.config.IPFSAPIPort, pm.config.IPFSGatewayPort, pm.config.IPFSSwarmPort))

	// Set API port
	if err := pm.runIPFSConfig(env, "Addresses.API", fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", pm.config.IPFSAPIPort)); err != nil {
		return err
	}

	// Set Gateway port
	if err := pm.runIPFSConfig(env, "Addresses.Gateway", fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", pm.config.IPFSGatewayPort)); err != nil {
		return err
	}

	// Set Swarm port - include all transport protocols (TCP, UDP/QUIC, WebRTC, WebTransport)
	// Must use --json flag since this is a JSON array
	port := pm.config.IPFSSwarmPort
	swarmAddrs := fmt.Sprintf(`["/ip4/0.0.0.0/tcp/%d", "/ip6/::/tcp/%d", "/ip4/0.0.0.0/udp/%d/webrtc-direct", "/ip4/0.0.0.0/udp/%d/quic-v1", "/ip4/0.0.0.0/udp/%d/quic-v1/webtransport", "/ip6/::/udp/%d/webrtc-direct", "/ip6/::/udp/%d/quic-v1", "/ip6/::/udp/%d/quic-v1/webtransport"]`,
		port, port, port, port, port, port, port, port)
	if err := pm.runIPFSConfigJSON(env, "Addresses.Swarm", swarmAddrs); err != nil {
		return err
	}

	pm.logInfo("IPFS configuration updated successfully")
	return nil
}

// runIPFSConfig runs an IPFS config command
func (pm *ProcessManager) runIPFSConfig(env []string, key, value string) error {
	cmd := exec.Command(pm.config.IPFSBinary, "config", key, value)
	cmd.Env = env

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ipfs config %s failed: %w\nOutput: %s", key, err, string(output))
	}

	return nil
}

// runIPFSConfigJSON runs an IPFS config command with --json flag for array/object values
func (pm *ProcessManager) runIPFSConfigJSON(env []string, key, jsonValue string) error {
	cmd := exec.Command(pm.config.IPFSBinary, "config", "--json", key, jsonValue)
	cmd.Env = env

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ipfs config --json %s failed: %w\nOutput: %s", key, err, string(output))
	}

	return nil
}

// StartPinShare starts the PinShare backend
func (pm *ProcessManager) StartPinShare(ctx context.Context) error {
	pm.processMu.Lock()
	defer pm.processMu.Unlock()

	// Check if PinShare binary exists
	if _, err := os.Stat(pm.config.PinShareBinary); os.IsNotExist(err) {
		return fmt.Errorf("PinShare binary not found at %s", pm.config.PinShareBinary)
	}

	// Open log file
	logPath := filepath.Join(pm.config.DataDirectory, winservice.DirLogs, "pinshare.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open PinShare log file: %w", err)
	}
	pm.pinshareLogFile = logFile

	// Create environment variables for PinShare
	dataPath := pm.config.GetPinShareDataPath()

	// Get current PATH and prepend install directory so 'ipfs' command is found
	currentPath := os.Getenv("PATH")
	newPath := pm.config.InstallDirectory
	if currentPath != "" {
		newPath = pm.config.InstallDirectory + ";" + currentPath
	}

	env := append(os.Environ(),
		fmt.Sprintf("PATH=%s", newPath),
		fmt.Sprintf("IPFS_API=http://localhost:%d", pm.config.IPFSAPIPort),
		fmt.Sprintf("PS_ORGNAME=%s", pm.config.OrgName),
		fmt.Sprintf("PS_GROUPNAME=%s", pm.config.GroupName),
		fmt.Sprintf("PS_LIBP2P_PORT=%d", pm.config.PinShareP2PPort),
		fmt.Sprintf("PS_UPLOAD_FOLDER=%s", filepath.Join(pm.config.DataDirectory, winservice.DirUpload)),
		fmt.Sprintf("PS_CACHE_FOLDER=%s", filepath.Join(pm.config.DataDirectory, winservice.DirCache)),
		fmt.Sprintf("PS_REJECT_FOLDER=%s", filepath.Join(pm.config.DataDirectory, winservice.DirRejected)),
		fmt.Sprintf("PS_METADATA_FILE=%s", filepath.Join(dataPath, "metadata.json")),
		fmt.Sprintf("PS_IDENTITY_KEY_FILE=%s", filepath.Join(dataPath, "identity.key")),
		fmt.Sprintf("PS_ENCRYPTION_KEY=%s", pm.config.EncryptionKey),
		fmt.Sprintf("PORT=%d", pm.config.PinShareAPIPort),
	)

	// Add feature flags
	// When running as a Windows service, chromedp cannot work (no desktop in Session 0).
	// If no real VT_TOKEN is provided, we MUST skip VirusTotal scanning to avoid chromedp deadlock.
	// Users who want VirusTotal scanning must provide a valid VT_TOKEN in the config.
	if pm.config.VirusTotalToken != "" {
		// Real VT token provided - use VirusTotal API scanning
		env = append(env, fmt.Sprintf("VT_TOKEN=%s", pm.config.VirusTotalToken))
		if pm.config.SkipVirusTotal {
			env = append(env, "PS_FF_SKIP_VT=true")
		}
	} else {
		// No VT token - must skip VT to avoid chromedp deadlock in service context
		env = append(env, "PS_FF_SKIP_VT=true")
		pm.logInfo("No VirusTotal token configured - virus scanning disabled (chromedp cannot run in service context)")
	}
	if pm.config.EnableCache {
		env = append(env, "PS_FF_CACHE=true")
	}
	if pm.config.ArchiveNode {
		env = append(env, "PS_FF_ARCHIVE_NODE=true")
	}

	// Create command
	pm.pinshareCmd = exec.CommandContext(ctx, pm.config.PinShareBinary)
	pm.pinshareCmd.Env = env
	pm.pinshareCmd.Stdout = logFile
	pm.pinshareCmd.Stderr = logFile
	pm.pinshareCmd.Dir = dataPath

	// Set process group for proper cleanup on Windows
	pm.pinshareCmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}

	// Start the process
	if err := pm.pinshareCmd.Start(); err != nil {
		logFile.Close()
		return fmt.Errorf("failed to start PinShare: %w", err)
	}

	pm.logInfo(fmt.Sprintf("PinShare backend started with PID %d", pm.pinshareCmd.Process.Pid))

	// Create exit channel and monitor process in background
	pm.pinshareExited = make(chan struct{})
	go pm.monitorProcess(ctx, pm.pinshareCmd, winservice.AppName, pm.pinshareExited)

	return nil
}

// monitorProcess monitors a process and logs when it exits.
// The exited channel is closed when the process exits, allowing Stop functions to wait.
func (pm *ProcessManager) monitorProcess(ctx context.Context, cmd *exec.Cmd, name string, exited chan struct{}) {
	defer close(exited)

	err := cmd.Wait()

	select {
	case <-ctx.Done():
		// Context cancelled, expected shutdown
		pm.logInfo(fmt.Sprintf("%s process exited (shutdown requested)", name))
	default:
		// Unexpected exit
		if err != nil {
			pm.logError(fmt.Sprintf("%s process exited unexpectedly", name), err)
		} else {
			pm.logInfo(fmt.Sprintf("%s process exited", name))
		}
	}
}

// StopIPFS stops the IPFS daemon
func (pm *ProcessManager) StopIPFS() error {
	pm.processMu.Lock()
	defer pm.processMu.Unlock()

	if pm.ipfsCmd == nil || pm.ipfsCmd.Process == nil {
		return nil
	}

	pm.logInfo("Stopping IPFS daemon...")
	pid := pm.ipfsCmd.Process.Pid

	// Kill the process tree using taskkill
	pm.killProcessByPID(pid, "IPFS")

	// If taskkill failed, try direct kill as fallback
	if pm.ipfsCmd.Process != nil {
		_ = pm.ipfsCmd.Process.Kill()
	}

	// Wait for process to exit via the monitor goroutine (with timeout)
	if pm.ipfsExited != nil {
		select {
		case <-time.After(winservice.ProcessShutdownTimeout):
			pm.logError("IPFS shutdown timeout", nil)
		case <-pm.ipfsExited:
			// Process exited, monitor goroutine has called Wait()
		}
	}

	// Close log file
	if pm.ipfsLogFile != nil {
		pm.ipfsLogFile.Close()
		pm.ipfsLogFile = nil
	}

	pm.ipfsCmd = nil
	pm.logInfo("IPFS daemon stopped")
	return nil
}

// StopPinShare stops the PinShare backend
func (pm *ProcessManager) StopPinShare() error {
	pm.processMu.Lock()
	defer pm.processMu.Unlock()

	if pm.pinshareCmd == nil || pm.pinshareCmd.Process == nil {
		return nil
	}

	pm.logInfo("Stopping PinShare backend...")
	pid := pm.pinshareCmd.Process.Pid

	// Kill the process tree using taskkill
	pm.killProcessByPID(pid, winservice.AppName)

	// If taskkill failed, try direct kill as fallback
	if pm.pinshareCmd.Process != nil {
		_ = pm.pinshareCmd.Process.Kill()
	}

	// Wait for process to exit via the monitor goroutine (with timeout)
	if pm.pinshareExited != nil {
		select {
		case <-time.After(winservice.ProcessShutdownTimeout):
			pm.logError("PinShare shutdown timeout", nil)
		case <-pm.pinshareExited:
			// Process exited, monitor goroutine has called Wait()
		}
	}

	// Close log file
	if pm.pinshareLogFile != nil {
		pm.pinshareLogFile.Close()
		pm.pinshareLogFile = nil
	}

	pm.pinshareCmd = nil
	pm.logInfo("PinShare backend stopped")
	return nil
}

// StopAll stops all managed processes (PinShare first, then IPFS)
func (pm *ProcessManager) StopAll() {
	// Stop PinShare first since it depends on IPFS
	if err := pm.StopPinShare(); err != nil {
		pm.logError("Error stopping PinShare during cleanup", err)
	}

	// Then stop IPFS
	if err := pm.StopIPFS(); err != nil {
		pm.logError("Error stopping IPFS during cleanup", err)
	}
}

// RestartIPFS restarts the IPFS daemon
func (pm *ProcessManager) RestartIPFS(ctx context.Context) error {
	if err := pm.StopIPFS(); err != nil {
		return err
	}
	time.Sleep(winservice.ServiceRestartDelay)
	return pm.StartIPFS(ctx)
}

// RestartPinShare restarts the PinShare backend
func (pm *ProcessManager) RestartPinShare(ctx context.Context) error {
	if err := pm.StopPinShare(); err != nil {
		return err
	}
	time.Sleep(winservice.ServiceRestartDelay)
	return pm.StartPinShare(ctx)
}

// Logging helpers
func (pm *ProcessManager) logInfo(msg string) {
	if pm.eventLog != nil {
		pm.eventLog.Info(1, msg)
	}
}

func (pm *ProcessManager) logError(msg string, err error) {
	errMsg := msg
	if err != nil {
		errMsg = fmt.Sprintf("%s: %v", msg, err)
	}
	if pm.eventLog != nil {
		pm.eventLog.Error(1, errMsg)
	}
}

package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"golang.org/x/sys/windows/svc/debug"
)

type ProcessManager struct {
	config        *ServiceConfig
	eventLog      debug.Log
	ipfsCmd       *exec.Cmd
	pinshareCmd   *exec.Cmd
	ipfsLogFile   *os.File
	pinshareLogFile *os.File
	mu            sync.Mutex
}

func NewProcessManager(config *ServiceConfig, eventLog debug.Log) *ProcessManager {
	return &ProcessManager{
		config:   config,
		eventLog: eventLog,
	}
}

// StartIPFS starts the IPFS daemon
func (pm *ProcessManager) StartIPFS(ctx context.Context) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Check if IPFS binary exists
	if _, err := os.Stat(pm.config.IPFSBinary); os.IsNotExist(err) {
		return fmt.Errorf("IPFS binary not found at %s", pm.config.IPFSBinary)
	}

	// Initialize IPFS repo if it doesn't exist
	repoPath := pm.config.GetIPFSRepoPath()
	if _, err := os.Stat(filepath.Join(repoPath, "config")); os.IsNotExist(err) {
		pm.logInfo("Initializing IPFS repository...")
		if err := pm.initializeIPFS(); err != nil {
			return fmt.Errorf("failed to initialize IPFS: %w", err)
		}
	}

	// Open log file
	logPath := filepath.Join(pm.config.DataDirectory, "logs", "ipfs.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open IPFS log file: %w", err)
	}
	pm.ipfsLogFile = logFile

	// Create command
	pm.ipfsCmd = exec.CommandContext(ctx, pm.config.IPFSBinary, "daemon")
	pm.ipfsCmd.Env = append(os.Environ(),
		fmt.Sprintf("IPFS_PATH=%s", repoPath),
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

	// Monitor process in background
	go pm.monitorProcess(ctx, pm.ipfsCmd, "IPFS")

	return nil
}

// initializeIPFS initializes a new IPFS repository
func (pm *ProcessManager) initializeIPFS() error {
	repoPath := pm.config.GetIPFSRepoPath()

	cmd := exec.Command(pm.config.IPFSBinary, "init")
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("IPFS_PATH=%s", repoPath),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ipfs init failed: %w\nOutput: %s", err, string(output))
	}

	pm.logInfo("IPFS repository initialized successfully")

	// Configure IPFS settings
	if err := pm.configureIPFS(); err != nil {
		return fmt.Errorf("failed to configure IPFS: %w", err)
	}

	return nil
}

// configureIPFS configures IPFS settings
func (pm *ProcessManager) configureIPFS() error {
	repoPath := pm.config.GetIPFSRepoPath()
	env := append(os.Environ(), fmt.Sprintf("IPFS_PATH=%s", repoPath))

	// Set API port
	if err := pm.runIPFSConfig(env, "Addresses.API", fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", pm.config.IPFSAPIPort)); err != nil {
		return err
	}

	// Set Gateway port
	if err := pm.runIPFSConfig(env, "Addresses.Gateway", fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", pm.config.IPFSGatewayPort)); err != nil {
		return err
	}

	// Set Swarm port
	if err := pm.runIPFSConfig(env, "Addresses.Swarm", fmt.Sprintf("[\"/ip4/0.0.0.0/tcp/%d\", \"/ip6/::/tcp/%d\"]", pm.config.IPFSSwarmPort, pm.config.IPFSSwarmPort)); err != nil {
		return err
	}

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

// StartPinShare starts the PinShare backend
func (pm *ProcessManager) StartPinShare(ctx context.Context) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Check if PinShare binary exists
	if _, err := os.Stat(pm.config.PinShareBinary); os.IsNotExist(err) {
		return fmt.Errorf("PinShare binary not found at %s", pm.config.PinShareBinary)
	}

	// Open log file
	logPath := filepath.Join(pm.config.DataDirectory, "logs", "pinshare.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open PinShare log file: %w", err)
	}
	pm.pinshareLogFile = logFile

	// Create environment variables for PinShare
	dataPath := pm.config.GetPinShareDataPath()
	env := append(os.Environ(),
		fmt.Sprintf("IPFS_API=http://localhost:%d", pm.config.IPFSAPIPort),
		fmt.Sprintf("PS_ORGNAME=%s", pm.config.OrgName),
		fmt.Sprintf("PS_GROUPNAME=%s", pm.config.GroupName),
		fmt.Sprintf("PS_LIBP2P_PORT=%d", pm.config.PinShareP2PPort),
		fmt.Sprintf("PS_UPLOAD_FOLDER=%s", filepath.Join(pm.config.DataDirectory, "upload")),
		fmt.Sprintf("PS_CACHE_FOLDER=%s", filepath.Join(pm.config.DataDirectory, "cache")),
		fmt.Sprintf("PS_REJECT_FOLDER=%s", filepath.Join(pm.config.DataDirectory, "rejected")),
		fmt.Sprintf("PS_METADATA_FILE=%s", filepath.Join(dataPath, "metadata.json")),
		fmt.Sprintf("PS_IDENTITY_KEY_FILE=%s", filepath.Join(dataPath, "identity.key")),
		fmt.Sprintf("PS_DATABASE_FILE=%s", filepath.Join(dataPath, "pinshare.db")),
		fmt.Sprintf("PS_ENCRYPTION_KEY=%s", pm.config.EncryptionKey),
		fmt.Sprintf("PORT=%d", pm.config.PinShareAPIPort),
	)

	// Add feature flags
	if pm.config.SkipVirusTotal {
		env = append(env, "PS_FF_SKIP_VT=true")
		// Set dummy VT_TOKEN to bypass chromedp test and use VirusTotal path
		// The application will use Security Capability 2 (VirusTotal API)
		if pm.config.VirusTotalToken == "" {
			env = append(env, "VT_TOKEN=SKIP_VT_FOR_SERVICE")
		}
	}
	if pm.config.EnableCache {
		env = append(env, "PS_FF_CACHE=true")
	}
	if pm.config.ArchiveNode {
		env = append(env, "PS_FF_ARCHIVE_NODE=true")
	}
	if pm.config.VirusTotalToken != "" {
		env = append(env, fmt.Sprintf("VT_TOKEN=%s", pm.config.VirusTotalToken))
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

	// Monitor process in background
	go pm.monitorProcess(ctx, pm.pinshareCmd, "PinShare")

	return nil
}

// monitorProcess monitors a process and logs when it exits
func (pm *ProcessManager) monitorProcess(ctx context.Context, cmd *exec.Cmd, name string) {
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
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.ipfsCmd == nil || pm.ipfsCmd.Process == nil {
		return nil
	}

	pm.logInfo("Stopping IPFS daemon...")

	// Send interrupt signal
	if err := pm.ipfsCmd.Process.Signal(os.Interrupt); err != nil {
		// If interrupt fails, try kill
		pm.logError("Failed to send interrupt to IPFS, forcing kill", err)
		if err := pm.ipfsCmd.Process.Kill(); err != nil {
			return fmt.Errorf("failed to kill IPFS process: %w", err)
		}
	}

	// Wait for process to exit (with timeout)
	done := make(chan error, 1)
	go func() {
		_, err := pm.ipfsCmd.Process.Wait()
		done <- err
	}()

	select {
	case <-time.After(10 * time.Second):
		pm.logError("IPFS shutdown timeout, forcing kill", nil)
		_ = pm.ipfsCmd.Process.Kill()
	case err := <-done:
		if err != nil {
			pm.logError("IPFS process wait error", err)
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
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.pinshareCmd == nil || pm.pinshareCmd.Process == nil {
		return nil
	}

	pm.logInfo("Stopping PinShare backend...")

	// Send interrupt signal
	if err := pm.pinshareCmd.Process.Signal(os.Interrupt); err != nil {
		// If interrupt fails, try kill
		pm.logError("Failed to send interrupt to PinShare, forcing kill", err)
		if err := pm.pinshareCmd.Process.Kill(); err != nil {
			return fmt.Errorf("failed to kill PinShare process: %w", err)
		}
	}

	// Wait for process to exit (with timeout)
	done := make(chan error, 1)
	go func() {
		_, err := pm.pinshareCmd.Process.Wait()
		done <- err
	}()

	select {
	case <-time.After(10 * time.Second):
		pm.logError("PinShare shutdown timeout, forcing kill", nil)
		_ = pm.pinshareCmd.Process.Kill()
	case err := <-done:
		if err != nil {
			pm.logError("PinShare process wait error", err)
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

// RestartIPFS restarts the IPFS daemon
func (pm *ProcessManager) RestartIPFS(ctx context.Context) error {
	if err := pm.StopIPFS(); err != nil {
		return err
	}
	time.Sleep(2 * time.Second)
	return pm.StartIPFS(ctx)
}

// RestartPinShare restarts the PinShare backend
func (pm *ProcessManager) RestartPinShare(ctx context.Context) error {
	if err := pm.StopPinShare(); err != nil {
		return err
	}
	time.Sleep(2 * time.Second)
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

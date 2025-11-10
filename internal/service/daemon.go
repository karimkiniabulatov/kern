package service

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/karimkiniabulatov/kern/internal/config"
)

type DaemonConfig struct {
	Enabled    bool   `json:"enabled"`
	Port       int    `json:"port"`
	AutoStart  bool   `json:"auto_start"`
	LogFile    string `json:"log_file"`
	PIDFile    string `json:"pid_file"`
}

type DaemonManager struct {
	config     *DaemonConfig
	appConfig  *config.Config
	httpServer *http.Server
}

func NewDaemonManager(appConfig *config.Config) *DaemonManager {
	return &DaemonManager{
		config:    loadDaemonConfig(),
		appConfig: appConfig,
	}
}

func loadDaemonConfig() *DaemonConfig {
	configPath, err := getDaemonConfigPath()
	if err != nil {
		return getDefaultDaemonConfig()
	}

	configFile := filepath.Join(configPath, "daemon.json")
	data, err := os.ReadFile(configFile)
	if err != nil {
		return getDefaultDaemonConfig()
	}

	var daemonConfig DaemonConfig
	if err := json.Unmarshal(data, &daemonConfig); err != nil {
		return getDefaultDaemonConfig()
	}

	return &daemonConfig
}

func (dm *DaemonManager) SaveConfig() error {
	configPath, err := getDaemonConfigPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(configPath, 0755); err != nil {
		return err
	}

	configFile := filepath.Join(configPath, "daemon.json")
	data, err := json.MarshalIndent(dm.config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configFile, data, 0644)
}

func getDaemonConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".config", "kern"), nil
}

func getDefaultDaemonConfig() *DaemonConfig {
	return &DaemonConfig{
		Enabled:   false,
		Port:      28126,
		AutoStart: false,
		LogFile:   "/var/log/kern.log",
		PIDFile:   "/var/run/kern.pid",
	}
}

// StartDaemon starts the API server in background mode
func (dm *DaemonManager) StartDaemon() error {
	if !dm.config.Enabled {
		return fmt.Errorf("daemon is not enabled")
	}

	// Check if already running
	if dm.IsRunning() {
		return fmt.Errorf("kern daemon is already running")
	}

	// Start in background
	cmd := exec.Command("kern", "--remote", strconv.Itoa(dm.config.Port))
	
	// Redirect output to log file
	if dm.config.LogFile != "" {
		logFile, err := os.OpenFile(dm.config.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("failed to open log file: %v", err)
		}
		cmd.Stdout = logFile
		cmd.Stderr = logFile
	}

	// Start process
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start daemon: %v", err)
	}

	// Save PID
	if err := dm.savePID(cmd.Process.Pid); err != nil {
		cmd.Process.Kill()
		return fmt.Errorf("failed to save PID: %v", err)
	}

	log.Printf("✅ kern daemon started on port %d (PID: %d)", dm.config.Port, cmd.Process.Pid)
	return nil
}

// StopDaemon stops the running daemon
func (dm *DaemonManager) StopDaemon() error {
	pid, err := dm.getPID()
	if err != nil {
		return fmt.Errorf("daemon is not running or PID file not found")
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		dm.removePID()
		return fmt.Errorf("process not found: %v", err)
	}

	// Send SIGTERM
	if err := process.Signal(os.Interrupt); err != nil {
		// Try force kill
		if err := process.Kill(); err != nil {
			return fmt.Errorf("failed to stop daemon: %v", err)
		}
	}

	dm.removePID()
	log.Printf("✅ kern daemon stopped (PID: %d)", pid)
	return nil
}

// RestartDaemon restarts the daemon
func (dm *DaemonManager) RestartDaemon() error {
	if err := dm.StopDaemon(); err != nil {
		log.Printf("⚠️ Could not stop daemon: %v", err)
	}

	time.Sleep(2 * time.Second)
	return dm.StartDaemon()
}

// IsRunning checks if daemon is running
func (dm *DaemonManager) IsRunning() bool {
	pid, err := dm.getPID()
	if err != nil {
		return false
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// Check if process is still alive
	return process.Signal(os.Signal(nil)) == nil
}

// Status returns daemon status information
func (dm *DaemonManager) Status() map[string]interface{} {
	status := map[string]interface{}{
		"enabled":    dm.config.Enabled,
		"auto_start": dm.config.AutoStart,
		"port":       dm.config.Port,
		"running":    dm.IsRunning(),
	}

	if dm.IsRunning() {
		pid, _ := dm.getPID()
		status["pid"] = pid
	}

	return status
}

// EnableAutoStart configures auto-start on system boot
func (dm *DaemonManager) EnableAutoStart() error {
	dm.config.AutoStart = true
	dm.config.Enabled = true
	
	if err := dm.SaveConfig(); err != nil {
		return fmt.Errorf("failed to save auto-start config: %v", err)
	}

	// Create systemd service file
	if err := dm.createSystemdService(); err != nil {
		return fmt.Errorf("failed to create systemd service: %v", err)
	}

	log.Printf("✅ Auto-start enabled for kern daemon")
	return nil
}

// DisableAutoStart disables auto-start
func (dm *DaemonManager) DisableAutoStart() error {
	dm.config.AutoStart = false
	
	if err := dm.SaveConfig(); err != nil {
		return fmt.Errorf("failed to save config: %v", err)
	}

	log.Printf("✅ Auto-start disabled for kern daemon")
	return nil
}

func (dm *DaemonManager) savePID(pid int) error {
	if dm.config.PIDFile == "" {
		return nil
	}
	return os.WriteFile(dm.config.PIDFile, []byte(strconv.Itoa(pid)), 0644)
}

func (dm *DaemonManager) getPID() (int, error) {
	if dm.config.PIDFile == "" {
		return 0, fmt.Errorf("PID file not configured")
	}

	data, err := os.ReadFile(dm.config.PIDFile)
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(string(data))
}

func (dm *DaemonManager) removePID() {
	if dm.config.PIDFile != "" {
		os.Remove(dm.config.PIDFile)
	}
}

func (dm *DaemonManager) createSystemdService() error {
	serviceContent := `[Unit]
Description=kern System Monitoring Daemon
After=network.target

[Service]
Type=simple
User=%s
ExecStart=/usr/local/bin/kern --daemon
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
`

	// Get current user
	user := os.Getenv("USER")
	if user == "" {
		user = "root"
	}

	serviceFile := fmt.Sprintf("/etc/systemd/system/kern.service", user)
	
	// Check if we have write permissions
	if _, err := os.Stat("/etc/systemd/system"); os.IsNotExist(err) {
		log.Printf("⚠️ Systemd directory not found, skipping service creation")
		return nil
	}

	// Try to create service file (requires sudo)
	cmd := exec.Command("sudo", "sh", "-c", 
		fmt.Sprintf("echo '%s' > %s", fmt.Sprintf(serviceContent, user), serviceFile))
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create systemd service (need sudo): %v", err)
	}

	// Enable the service
	enableCmd := exec.Command("sudo", "systemctl", "enable", "kern.service")
	if err := enableCmd.Run(); err != nil {
		return fmt.Errorf("failed to enable service: %v", err)
	}

	return nil
}

func (dm *DaemonManager) GetConfig() *DaemonConfig {
	return dm.config
}

func (dm *DaemonManager) UpdateConfig(newConfig *DaemonConfig) error {
	dm.config = newConfig
	return dm.SaveConfig()
}
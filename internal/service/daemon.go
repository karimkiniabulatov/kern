package service

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"time"
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
	httpServer *http.Server
}

func NewDaemonManager() *DaemonManager {
	return &DaemonManager{
		config: loadDaemonConfig(),
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
		// Если файла нет, создаем дефолтный и сразу сохраняем
		defaultConfig := getDefaultDaemonConfig()
		defaultConfig.Enabled = true // ВКЛЮЧАЕМ по умолчанию!
		defaultConfig.AutoStart = true // АВТОСТАРТ по умолчанию!
		saveDaemonConfig(defaultConfig)
		return defaultConfig
	}

	var daemonConfig DaemonConfig
	if err := json.Unmarshal(data, &daemonConfig); err != nil {
		return getDefaultDaemonConfig()
	}

	return &daemonConfig
}

func saveDaemonConfig(config *DaemonConfig) error {
	configPath, err := getDaemonConfigPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(configPath, 0755); err != nil {
		return err
	}

	configFile := filepath.Join(configPath, "daemon.json")
	data, err := json.MarshalIndent(config, "", "  ")
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
	// Кроссплатформенные пути по умолчанию
	logFile := "/var/log/kern-daemon.log"
	pidFile := "/tmp/kern-daemon.pid"
	
	if runtime.GOOS == "windows" {
		logFile = filepath.Join(os.TempDir(), "kern-daemon.log")
		pidFile = filepath.Join(os.TempDir(), "kern-daemon.pid")
	}

	return &DaemonConfig{
		Enabled:   true,  // ВКЛЮЧЕНО по умолчанию!
		Port:      28126,
		AutoStart: true,  // АВТОСТАРТ по умолчанию!
		LogFile:   logFile,
		PIDFile:   pidFile,
	}
}

// StartDaemon starts the API server in background mode
func (dm *DaemonManager) StartDaemon() error {
	// Если уже запущен, останавливаем сначала
	if dm.IsRunning() {
		log.Println("Daemon already running, restarting...")
		dm.StopDaemon()
		time.Sleep(2 * time.Second)
	}

	// Собираем команду
	cmdArgs := []string{"--remote", strconv.Itoa(dm.config.Port)}
	
	// Если есть лог файл, добавляем логирование
	if dm.config.LogFile != "" {
		// Убедимся, что директория для логов существует
		logDir := filepath.Dir(dm.config.LogFile)
		os.MkdirAll(logDir, 0755)
	}

	cmd := exec.Command("kern", cmdArgs...)
	
	// Направляем вывод в лог файл или в /dev/null (nul на Windows)
	if dm.config.LogFile != "" {
		logFile, err := os.OpenFile(dm.config.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			log.Printf("Cannot open log file: %v, using stdout", err)
		} else {
			cmd.Stdout = logFile
			cmd.Stderr = logFile
		}
	} else {
		// Если лог файл не указан, направляем в nul (/dev/null на Unix)
		nullPath := "/dev/null"
		if runtime.GOOS == "windows" {
			nullPath = "nul"
		}
		nullFile, _ := os.OpenFile(nullPath, os.O_WRONLY, 0644)
		cmd.Stdout = nullFile
		cmd.Stderr = nullFile
	}

	// Запускаем процесс в фоне
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start daemon: %v", err)
	}

	// Сохраняем PID
	if err := dm.savePID(cmd.Process.Pid); err != nil {
		log.Printf("Failed to save PID: %v", err)
	}

	// Ждем немного чтобы сервер успел запуститься
	time.Sleep(1 * time.Second)

	// Проверяем, что сервер действительно запустился
	if !dm.checkServerRunning() {
		// Если не запустился, пытаемся убить процесс и возвращаем ошибку
		cmd.Process.Kill()
		return fmt.Errorf("daemon started but API server is not responding")
	}

	log.Printf("kern daemon started successfully on port %d (PID: %d)", dm.config.Port, cmd.Process.Pid)
	return nil
}

// checkServerRunning проверяет, что API сервер действительно работает
func (dm *DaemonManager) checkServerRunning() bool {
	url := fmt.Sprintf("http://localhost:%d/health", dm.config.Port)
	
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	
	return resp.StatusCode == http.StatusOK
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

	// Пытаемся корректно завершить
	if err := process.Signal(os.Interrupt); err != nil {
		// Если не получается, убиваем форсированно
		if err := process.Kill(); err != nil {
			dm.removePID()
			return fmt.Errorf("failed to stop daemon: %v", err)
		}
	}

	// Ждем завершения
	time.Sleep(1 * time.Second)
	dm.removePID()
	
	log.Printf("kern daemon stopped (PID: %d)", pid)
	return nil
}

// RestartDaemon restarts the daemon
func (dm *DaemonManager) RestartDaemon() error {
	log.Println("Restarting kern daemon...")
	
	if err := dm.StopDaemon(); err != nil {
		log.Printf("Could not stop daemon: %v", err)
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

	// Проверяем что процесс жив И сервер отвечает
	if process.Signal(os.Signal(nil)) != nil {
		return false
	}

	return dm.checkServerRunning()
}

// Status returns daemon status information
func (dm *DaemonManager) Status() map[string]interface{} {
	isRunning := dm.IsRunning()
	status := map[string]interface{}{
		"enabled":    dm.config.Enabled,
		"auto_start": dm.config.AutoStart,
		"port":       dm.config.Port,
		"running":    isRunning,
		"api_url":    fmt.Sprintf("http://localhost:%d", dm.config.Port),
	}

	if isRunning {
		pid, _ := dm.getPID()
		status["pid"] = pid
		
		// Проверяем доступность API
		if dm.checkServerRunning() {
			status["api_status"] = "healthy"
		} else {
			status["api_status"] = "unresponsive"
		}
	}

	return status
}

// EnableAutoStart configures auto-start on system boot
func (dm *DaemonManager) EnableAutoStart() error {
	dm.config.AutoStart = true
	dm.config.Enabled = true
	
	if err := saveDaemonConfig(dm.config); err != nil {
		return fmt.Errorf("failed to save auto-start config: %v", err)
	}

	// Сразу запускаем демона если он не запущен
	if !dm.IsRunning() {
		if err := dm.StartDaemon(); err != nil {
			return fmt.Errorf("failed to start daemon: %v", err)
		}
	}

	log.Printf("Auto-start enabled and daemon started on port %d", dm.config.Port)
	return nil
}

// DisableAutoStart disables auto-start
func (dm *DaemonManager) DisableAutoStart() error {
	dm.config.AutoStart = false
	
	if err := saveDaemonConfig(dm.config); err != nil {
		return fmt.Errorf("failed to save config: %v", err)
	}

	log.Printf("Auto-start disabled for kern daemon")
	return nil
}

// EnsureRunning гарантирует что демон запущен
func (dm *DaemonManager) EnsureRunning() error {
	if dm.IsRunning() {
		return nil
	}

	if !dm.config.Enabled {
		return fmt.Errorf("daemon is disabled in configuration")
	}

	log.Println("Starting kern daemon (ensure running)...")
	return dm.StartDaemon()
}

// GetConfig returns current daemon configuration
func (dm *DaemonManager) GetConfig() *DaemonConfig {
	return dm.config
}

// UpdateConfig updates and saves daemon configuration
func (dm *DaemonManager) UpdateConfig(newConfig *DaemonConfig) error {
	dm.config = newConfig
	return saveDaemonConfig(newConfig)
}

func (dm *DaemonManager) savePID(pid int) error {
	if dm.config.PIDFile == "" {
		// Кроссплатформенный путь по умолчанию
		if runtime.GOOS == "windows" {
			dm.config.PIDFile = filepath.Join(os.TempDir(), "kern-daemon.pid")
		} else {
			dm.config.PIDFile = "/tmp/kern-daemon.pid"
		}
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
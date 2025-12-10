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
	"syscall"
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
		defaultConfig.Enabled = true
		defaultConfig.AutoStart = true
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
	
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(homeDir, "AppData", "Local", "kern"), nil
	case "darwin": // macOS
		return filepath.Join(homeDir, "Library", "Application Support", "kern"), nil
	default: // Linux и другие Unix-системы
		return filepath.Join(homeDir, ".config", "kern"), nil
	}
}

func getDefaultDaemonConfig() *DaemonConfig {
	configPath, err := getDaemonConfigPath()
	if err != nil {
		// Fallback на временную директорию если не удалось получить config path
		configPath = os.TempDir()
	}

	logFile := filepath.Join(configPath, "kern-daemon.log")
	pidFile := filepath.Join(configPath, "kern-daemon.pid")

	return &DaemonConfig{
		Enabled:   true,
		Port:      28126,
		AutoStart: true,
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

	// Создаем директории для логов и PID файлов если нужно
	if err := os.MkdirAll(filepath.Dir(dm.config.LogFile), 0755); err != nil {
		log.Printf("Warning: cannot create log directory: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(dm.config.PIDFile), 0755); err != nil {
		log.Printf("Warning: cannot create PID directory: %v", err)
	}

	// Собираем команду
	cmdArgs := []string{"--remote", strconv.Itoa(dm.config.Port)}
	
	cmd := exec.Command("kern", cmdArgs...)
	
	// Направляем вывод в лог файл или в nul
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
		nullFile, err := os.OpenFile(nullPath, os.O_WRONLY, 0644)
		if err == nil {
			cmd.Stdout = nullFile
			cmd.Stderr = nullFile
		}
	}

	// Устанавливаем флаги для корректной работы в фоне в зависимости от ОС
	if runtime.GOOS != "windows" {
		// На Unix-системах устанавливаем Setsid чтобы процесс стал лидером сессии
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setsid: true,
		}
	}

	// Запускаем процесс в фоне
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start daemon: %v", err)
	}

	// Сохраняем PID
	if err := dm.savePID(cmd.Process.Pid); err != nil {
		log.Printf("Failed to save PID: %v", err)
	}

	// Отсоединяем процесс от родительского (для Unix-систем)
	if runtime.GOOS != "windows" {
		// На Unix мы можем сразу завершить родительский процесс, оставив дочерний работать
		go func() {
			time.Sleep(100 * time.Millisecond)
			cmd.Process.Release()
		}()
	}

	// Ждем немного чтобы сервер успел запуститься
	time.Sleep(2 * time.Second)

	// Проверяем, что сервер действительно запустился
	if !dm.checkServerRunning() {
		// Если не запустился, пытаемся убить процесс и возвращаем ошибку
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		dm.removePID()
		return fmt.Errorf("daemon started but API server is not responding")
	}

	log.Printf("kern daemon started successfully on port %d (PID: %d)", dm.config.Port, cmd.Process.Pid)
	return nil
}

// checkServerRunning проверяет, что API сервер действительно работает
func (dm *DaemonManager) checkServerRunning() bool {
	url := fmt.Sprintf("http://localhost:%d/health", dm.config.Port)
	
	client := &http.Client{Timeout: 3 * time.Second}
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

	// Ждем завершения и проверяем что процесс завершился
	for i := 0; i < 10; i++ {
		if !dm.isProcessRunning(pid) {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	dm.removePID()
	
	log.Printf("kern daemon stopped (PID: %d)", pid)
	return nil
}

// isProcessRunning проверяет работает ли процесс
func (dm *DaemonManager) isProcessRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	
	// На Windows всегда возвращает nil, поэтому используем другой метод
	if runtime.GOOS == "windows" {
		// На Windows посылаем сигнал 0 чтобы проверить процесс
		err := process.Signal(os.Signal(nil))
		return err == nil
	}
	
	// На Unix системах используем kill -0
	err = process.Signal(syscall.Signal(0))
	return err == nil
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

	if !dm.isProcessRunning(pid) {
		dm.removePID()
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
		"platform":   runtime.GOOS,
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

	// Платформо-специфичная настройка автозапуска
	if err := dm.setupAutoStart(); err != nil {
		log.Printf("Warning: failed to setup platform-specific auto-start: %v", err)
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

	// Удаляем платформо-специфичную настройку автозапуска
	if err := dm.removeAutoStart(); err != nil {
		log.Printf("Warning: failed to remove platform-specific auto-start: %v", err)
	}

	log.Printf("Auto-start disabled for kern daemon")
	return nil
}

// setupAutoStart настраивает автозапуск в зависимости от платформы
func (dm *DaemonManager) setupAutoStart() error {
	switch runtime.GOOS {
	case "windows":
		return dm.setupWindowsAutoStart()
	case "darwin":
		return dm.setupMacAutoStart()
	case "linux":
		return dm.setupLinuxAutoStart()
	default:
		log.Printf("Auto-start not implemented for platform: %s", runtime.GOOS)
		return nil
	}
}

// removeAutoStart удаляет настройки автозапуска
func (dm *DaemonManager) removeAutoStart() error {
	switch runtime.GOOS {
	case "windows":
		return dm.removeWindowsAutoStart()
	case "darwin":
		return dm.removeMacAutoStart()
	case "linux":
		return dm.removeLinuxAutoStart()
	default:
		return nil
	}
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
		configPath, err := getDaemonConfigPath()
		if err != nil {
			configPath = os.TempDir()
		}
		dm.config.PIDFile = filepath.Join(configPath, "kern-daemon.pid")
	}
	
	// Убедимся что директория существует
	if err := os.MkdirAll(filepath.Dir(dm.config.PIDFile), 0755); err != nil {
		return err
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

// Методы для платформо-специфичного автозапуска (заглушки - требуют реализации)
func (dm *DaemonManager) setupWindowsAutoStart() error {
	// Реализация через реестр или папку автозагрузки
	log.Printf("Windows auto-start setup would be implemented here")
	return nil
}

func (dm *DaemonManager) removeWindowsAutoStart() error {
	log.Printf("Windows auto-start removal would be implemented here")
	return nil
}

func (dm *DaemonManager) setupMacAutoStart() error {
	// Реализация через launchd/LaunchAgents
	log.Printf("macOS auto-start setup would be implemented here")
	return nil
}

func (dm *DaemonManager) removeMacAutoStart() error {
	log.Printf("macOS auto-start removal would be implemented here")
	return nil
}

func (dm *DaemonManager) setupLinuxAutoStart() error {
	// Реализация через systemd, init.d, или desktop-файлы в зависимости от дистрибутива
	log.Printf("Linux auto-start setup would be implemented here")
	return nil
}

func (dm *DaemonManager) removeLinuxAutoStart() error {
	log.Printf("Linux auto-start removal would be implemented here")
	return nil
}

// AppManagement управляет самим приложением kern
func (dm *DaemonManager) AppManagement() map[string]func() error {
    return map[string]func() error{
        "pause":  dm.pauseApp,
        "resume": dm.resumeApp,
        "stop":   dm.stopApp,
        "restart": dm.restartApp,
    }
}

func (dm *DaemonManager) pauseApp() error {
    // Реализация приостановки приложения
    return nil
}

func (dm *DaemonManager) resumeApp() error {
    // Реализация возобновления приложения  
    return nil
}

func (dm *DaemonManager) stopApp() error {
    // Реализация остановки приложения
    return nil
}

func (dm *DaemonManager) restartApp() error {
    // Реализация перезапуска приложения
    return nil
}
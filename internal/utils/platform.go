package utils

import (
	"os"
    "os/exec"
    "runtime"
    "strings"
)

// IsCommandAvailable проверяет доступность команды в системе
func IsCommandAvailable(name string) bool {
    _, err := exec.LookPath(name)
    return err == nil
}

// GetOSVersion возвращает детальную информацию об ОС
/*
func GetOSVersion() string {
    switch runtime.GOOS {
    case "linux":
        if data, err := os.ReadFile("/etc/os-release"); err == nil {
            return parseOSRelease(string(data))
        }
    case "windows":
        cmd := exec.Command("systeminfo")
        if output, err := cmd.Output(); err == nil {
            return parseWindowsSystemInfo(string(output))
        }
    case "darwin":
        cmd := exec.Command("sw_vers")
        if output, err := cmd.Output(); err == nil {
            return parseMacOSVersion(string(output))
        }
    }
    return runtime.GOOS
}
*/
func GetOSVersion() string {
    return runtime.GOOS
}

// ExecuteCommand выполняет команду с обработкой ошибок
func ExecuteCommand(name string, args ...string) (string, error) {
    cmd := exec.Command(name, args...)
    output, err := cmd.Output()
    if err != nil {
        return "", err
    }
    return strings.TrimSpace(string(output)), nil
}
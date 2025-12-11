package utils

import (
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
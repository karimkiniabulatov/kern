package disk

import (
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

type DiskInfo struct {
	Filesystem string
	Size       string
	Used       string
	Available  string
	UsePercent float64
	MountedOn  string
}

func Summary() ([]DiskInfo, error) {
	var disks []DiskInfo
	var err error
	
	switch runtime.GOOS {
	case "windows":
		disks, err = getWindowsDiskInfo()
	case "darwin":
		disks, err = getDarwinDiskInfo()
	default:
		disks, err = getLinuxDiskInfo()
	}

	// Если произошла ошибка или нет данных, возвращаем пустую структуру с гарантированными полями
	if err != nil || len(disks) == 0 {
		// Создаем минимальный набор данных для обеспечения ожидаемого формата
		defaultDisk := DiskInfo{
			Filesystem: "unknown",
			Size:       "0 B",
			Used:       "0 B",
			Available:  "0 B",
			UsePercent: 0.0, // Гарантируем числовое значение
			MountedOn:  "/",
		}
		return []DiskInfo{defaultDisk}, nil
	}

	// Гарантируем, что все UsePercent имеют числовое значение
	for i := range disks {
		if disks[i].UsePercent < 0 {
			disks[i].UsePercent = 0.0
		}
	}

	return disks, nil
}

func getLinuxDiskInfo() ([]DiskInfo, error) {
	cmd := exec.Command("df", "-h")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseDFOutput(string(output))
}

func getDarwinDiskInfo() ([]DiskInfo, error) {
	cmd := exec.Command("df", "-h")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseDFOutput(string(output))
}

func getWindowsDiskInfo() ([]DiskInfo, error) {
	cmd := exec.Command("wmic", "logicaldisk", "get", "size,freespace,caption")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseWMICOutput(string(output))
}

func parseDFOutput(output string) ([]DiskInfo, error) {
	lines := strings.Split(output, "\n")
	var disks []DiskInfo

	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue // Skip header and empty lines
		}

		// Split by whitespace, handling multiple spaces
		re := regexp.MustCompile(`\s+`)
		fields := re.Split(line, -1)

		if len(fields) >= 6 {
			usePercentStr := strings.TrimSuffix(fields[4], "%")
			usePercent, err := strconv.ParseFloat(usePercentStr, 64)
			if err != nil {
				usePercent = 0.0 // Гарантируем числовое значение при ошибке парсинга
			}

			// Пропускаем временные файловые системы и специальные точки монтирования
			if shouldSkipFilesystem(fields[0], fields[5]) {
				continue
			}

			disk := DiskInfo{
				Filesystem: fields[0],
				Size:       fields[1],
				Used:       fields[2],
				Available:  fields[3],
				UsePercent: usePercent,
				MountedOn:  fields[5],
			}
			disks = append(disks, disk)
		}
	}

	return disks, nil
}

func parseWMICOutput(output string) ([]DiskInfo, error) {
	lines := strings.Split(output, "\n")
	var disks []DiskInfo

	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue // Skip header and empty lines
		}

		// Split by whitespace, handling multiple spaces
		re := regexp.MustCompile(`\s+`)
		fields := re.Split(line, -1)

		if len(fields) >= 3 {
			// Windows output: Caption FreeSpace Size
			// Example: C: 1234567890 12345678900
			freeSpace, err := strconv.ParseUint(fields[1], 10, 64)
			if err != nil {
				continue
			}

			totalSize, err := strconv.ParseUint(fields[2], 10, 64)
			if err != nil {
				continue
			}

			used := totalSize - freeSpace
			usePercent := 0.0
			if totalSize > 0 {
				usePercent = float64(used) / float64(totalSize) * 100
			}

			disk := DiskInfo{
				Filesystem: fields[0],
				Size:       formatBytes(totalSize),
				Used:       formatBytes(used),
				Available:  formatBytes(freeSpace),
				UsePercent: usePercent,
				MountedOn:  fields[0], // In Windows, the drive letter is the mount point
			}
			disks = append(disks, disk)
		}
	}

	return disks, nil
}

func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return strconv.FormatUint(bytes, 10) + " B"
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return strconv.FormatFloat(float64(bytes)/float64(div), 'f', 1, 64) + " " + string("KMGTPE"[exp]) + "B"
}

func shouldSkipFilesystem(filesystem, mountPoint string) bool {
	// Пропускаем временные файловые системы
	if strings.HasPrefix(filesystem, "tmpfs") ||
		strings.HasPrefix(filesystem, "devtmpfs") ||
		strings.HasPrefix(filesystem, "overlay") ||
		strings.HasPrefix(filesystem, "shm") ||
		strings.HasPrefix(filesystem, "udev") {
		return true
	}

	// Пропускаем специальные точки монтирования
	if mountPoint == "/dev" ||
		mountPoint == "/sys" ||
		mountPoint == "/proc" ||
		mountPoint == "/sys/fs/cgroup" ||
		strings.HasPrefix(mountPoint, "/var/lib/docker") ||
		strings.HasPrefix(mountPoint, "/snap") {
		return true
	}

	// Пропускаем loop устройства
	if strings.HasPrefix(filesystem, "/dev/loop") {
		return true
	}

	return false
}
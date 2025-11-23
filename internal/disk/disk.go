package disk

import (
	"fmt"
	"os"
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
	Physical   bool   // true = физический, false = логический
	DiskType   string // "SSD", "HDD", "NVMe", "RAID", "Network", "Unknown"
	Model      string // Модель диска
	Serial     string // Серийный номер
	SMARTStatus string // SMART-статус: "PASSED", "FAILED", "UNKNOWN", "Unavailable"
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
			UsePercent: 0.0,
			MountedOn:  "/",
			Physical:   false,
			DiskType:   "Unknown",
			Model:      "Unknown",
			Serial:     "Unknown",
			SMARTStatus: "UNKNOWN",
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
				usePercent = 0.0
			}

			// Пропускаем временные файловые системы и специальные точки монтирования
			if shouldSkipFilesystem(fields[0], fields[5]) {
				continue
			}

			// Определяем тип устройства и физические характеристики
			physical, diskType, model, serial, smartStatus := detectDiskProperties(fields[0])

			disk := DiskInfo{
				Filesystem: fields[0],
				Size:       fields[1],
				Used:       fields[2],
				Available:  fields[3],
				UsePercent: usePercent,
				MountedOn:  fields[5],
				Physical:   physical,
				DiskType:   diskType,
				Model:      model,
				Serial:     serial,
				SMARTStatus: smartStatus,
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
			continue
		}

		re := regexp.MustCompile(`\s+`)
		fields := re.Split(line, -1)

		if len(fields) >= 3 {
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

			// Для Windows определяем тип устройства
			physical, diskType, model, serial, smartStatus := detectWindowsDiskProperties(fields[0])

			disk := DiskInfo{
				Filesystem: fields[0],
				Size:       formatBytes(totalSize),
				Used:       formatBytes(used),
				Available:  formatBytes(freeSpace),
				UsePercent: usePercent,
				MountedOn:  fields[0],
				Physical:   physical,
				DiskType:   diskType,
				Model:      model,
				Serial:     serial,
				SMARTStatus: smartStatus,
			}
			disks = append(disks, disk)
		}
	}

	return disks, nil
}

// detectDiskProperties определяет свойства диска для Linux/Unix систем
func detectDiskProperties(filesystem string) (bool, string, string, string, string) {
	physical := true
	diskType := "Unknown"
	model := "Unknown"
	serial := "Unknown"
	smartStatus := "UNKNOWN"

	// Извлекаем имя устройства из пути (например, /dev/sda1 -> sda)
	device := strings.TrimPrefix(filesystem, "/dev/")
	if device == "" {
		return physical, diskType, model, serial, smartStatus
	}

	// Получаем базовое устройство (без номера раздела)
	baseDevice := getBaseDevice(device)
	
	// Определяем базовый тип по имени устройства
	if strings.Contains(filesystem, "nvme") {
		diskType = "NVMe"
	} else if strings.Contains(filesystem, "ssd") {
		diskType = "SSD"
	} else if strings.Contains(filesystem, "sd") || strings.Contains(filesystem, "hd") {
		diskType = "HDD"
	} else if strings.Contains(filesystem, "md") {
		diskType = "RAID"
	} else if strings.Contains(filesystem, "dm-") {
		diskType = "LVM"
		physical = false
	} else if strings.Contains(filesystem, "loop") {
		diskType = "Loopback"
		physical = false
	} else if strings.HasPrefix(filesystem, "//") || strings.Contains(filesystem, "nfs") {
		diskType = "Network"
		physical = false
	} else if strings.Contains(filesystem, "tmpfs") {
		diskType = "Temporary"
		physical = false
	}

	// Пытаемся получить дополнительную информацию через lsblk и smartctl (для Linux)
	if runtime.GOOS != "windows" && diskType != "Network" && diskType != "Temporary" && physical {
		// Уточняем тип диска через rotational флаг
		rotationalPath := fmt.Sprintf("/sys/block/%s/queue/rotational", baseDevice)
		if data, err := os.ReadFile(rotationalPath); err == nil {
			if strings.TrimSpace(string(data)) == "0" {
				diskType = "SSD"
			} else if strings.TrimSpace(string(data)) == "1" {
				diskType = "HDD"
			}
		}

		// Получаем информацию через lsblk
		if info, err := getDiskInfoWithLsblk(filesystem); err == nil {
			if info.diskType != "" && info.diskType != "Partition" {
				diskType = info.diskType
			}
			if info.model != "" {
				model = info.model
			}
			if info.serial != "" {
				serial = info.serial
			}
		}

		// Получаем SMART-статус
		smartStatus = getSMARTStatus(baseDevice)

		// Если это LVM, RAID или другие виртуальные устройства, помечаем как логические
		if diskType == "LVM" || strings.Contains(diskType, "RAID") {
			physical = false
		}
	}

	return physical, diskType, model, serial, smartStatus
}

// getBaseDevice возвращает базовое имя устройства без номера раздела
func getBaseDevice(device string) string {
	// Убираем цифры в конце для SATA/SCSI устройств (sda1 -> sda)
	re := regexp.MustCompile(`^([a-z]+)(\d+)$`)
	if matches := re.FindStringSubmatch(device); matches != nil {
		return matches[1]
	}
	
	// Убираем часть после 'p' для NVMe устройств (nvme0n1p1 -> nvme0n1)
	re = regexp.MustCompile(`^(nvme\d+n\d+)p\d+$`)
	if matches := re.FindStringSubmatch(device); matches != nil {
		return matches[1]
	}
	
	// Убираем часть после 'p' для MMC устройств (mmcblk0p1 -> mmcblk0)
	re = regexp.MustCompile(`^(mmcblk\d+)p\d+$`)
	if matches := re.FindStringSubmatch(device); matches != nil {
		return matches[1]
	}
	
	return device
}

// getSMARTStatus получает SMART-статус диска
func getSMARTStatus(device string) string {
	// Пытаемся использовать smartctl для получения SMART-статуса
	cmd := exec.Command("smartctl", "-H", "/dev/"+device)
	output, err := cmd.Output()
	if err != nil {
		return "Unavailable"
	}
	
	outputStr := string(output)
	if strings.Contains(outputStr, "PASSED") || strings.Contains(outputStr, "OK") {
		return "PASSED"
	} else if strings.Contains(outputStr, "FAILED") {
		return "FAILED"
	}
	
	return "UNKNOWN"
}

// Структура для хранения информации из lsblk
type lsblkInfo struct {
	diskType string
	model    string
	serial   string
}

// getDiskInfoWithLsblk получает информацию о диске через lsblk
func getDiskInfoWithLsblk(filesystem string) (lsblkInfo, error) {
	info := lsblkInfo{}
	
	// Извлекаем имя устройства из пути (например, /dev/sda1 -> sda)
	device := strings.TrimPrefix(filesystem, "/dev/")
	if device == "" {
		return info, exec.ErrNotFound
	}

	// Выполняем lsblk для получения информации об устройстве
	cmd := exec.Command("lsblk", "-o", "TYPE,MODEL,SERIAL", "-n", "-d", device)
	output, err := cmd.Output()
	if err != nil {
		return info, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) > 0 {
		fields := strings.Fields(lines[0])
		if len(fields) >= 1 {
			// Определяем тип устройства
			switch fields[0] {
			case "disk":
				info.diskType = "HDD" // Будет уточнено через rotational флаг
			case "part":
				info.diskType = "Partition"
				return info, nil // Для разделов не получаем модель/серийник
			case "rom":
				info.diskType = "ROM"
			case "lvm":
				info.diskType = "LVM"
			case "raid":
				info.diskType = "RAID"
			}

			// Получаем модель и серийный номер
			if len(fields) >= 2 {
				info.model = fields[1]
			}
			if len(fields) >= 3 {
				info.serial = fields[2]
			}
		}
	}

	return info, nil
}

// detectWindowsDiskProperties определяет свойства диска для Windows
func detectWindowsDiskProperties(drive string) (bool, string, string, string, string) {
	physical := true
	diskType := "Unknown"
	model := "Unknown"
	serial := "Unknown"
	smartStatus := "UNKNOWN"

	// Для Windows пытаемся определить тип через wmic
	cmd := exec.Command("wmic", "diskdrive", "get", "model,serialnumber,mediatype")
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				model = fields[0]
				serial = fields[1]
				mediaType := fields[2]
				
				switch mediaType {
				case "Fixed hard disk media":
					diskType = "HDD"
				case "SSD":
					diskType = "SSD"
				case "NVMe":
					diskType = "NVMe"
				}
				break
			}
		}
	}

	// Сетевые диски и CD-ROM помечаем как логические
	if strings.HasPrefix(drive, "\\\\") {
		diskType = "Network"
		physical = false
	} else if drive == "CD-ROM" {
		diskType = "CD-ROM"
		physical = false
	}

	return physical, diskType, model, serial, smartStatus
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
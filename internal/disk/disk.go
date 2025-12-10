package disk

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
)

var (
	previousDisks   []DiskInfo
	diskCacheMutex sync.RWMutex
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
    Vendor     string // Производитель (вендор)
    SMARTStatus string // SMART-статус: "PASSED", "FAILED", "UNKNOWN", "Unavailable"
}

// GroupDisksByPhysicalDevice группирует диски по физическим устройствам (экспортируемая версия)
func GroupDisksByPhysicalDevice(disks []DiskInfo) map[string][]DiskInfo {
    return groupDisksByPhysicalDevice(disks)
}

// groupDisksByPhysicalDevice группирует диски по физическим устройствам
func groupDisksByPhysicalDevice(disks []DiskInfo) map[string][]DiskInfo {
    groups := make(map[string][]DiskInfo)
    
    for _, disk := range disks {
        // Создаем ключ для группировки
        var key string
        
        if disk.Physical && disk.Model != "Unknown" && disk.Serial != "Unknown" {
            // Для физических дисков используем модель+серийник
            key = fmt.Sprintf("PHYSICAL:%s:%s", disk.Model, disk.Serial)
        } else if strings.HasPrefix(disk.Filesystem, "/dev/") {
            // Для логических разделов определяем базовое устройство
            baseDevice := getBaseDevice(strings.TrimPrefix(disk.Filesystem, "/dev/"))
            key = fmt.Sprintf("LOGICAL:%s", baseDevice)
        } else if strings.Contains(disk.DiskType, "RAID") {
            // Для RAID массивов
            key = fmt.Sprintf("RAID:%s", disk.Model)
        } else {
            // Остальные
            key = fmt.Sprintf("OTHER:%s", disk.Filesystem)
        }
        
        groups[key] = append(groups[key], disk)
    }
    
    return groups
}

// sortDisks стабильно сортирует диски для отображения
func sortDisks(disks []DiskInfo) []DiskInfo {
    // Группируем по физическим устройствам
    groups := GroupDisksByPhysicalDevice(disks)
    
    var sortedDisks []DiskInfo
    
    // Сначала физические диски
    for key, group := range groups {
        if strings.HasPrefix(key, "PHYSICAL:") {
            // Сортируем разделы внутри физического диска
            sort.SliceStable(group, func(i, j int) bool {
                // Сначала корневые разделы, затем остальные
                if group[i].MountedOn == "/" && group[j].MountedOn != "/" {
                    return true
                }
                if group[i].MountedOn != "/" && group[j].MountedOn == "/" {
                    return false
                }
                return group[i].Filesystem < group[j].Filesystem
            })
            sortedDisks = append(sortedDisks, group...)
        }
    }
    
    // Затем RAID массивы
    for key, group := range groups {
        if strings.HasPrefix(key, "RAID:") {
            sortedDisks = append(sortedDisks, group...)
        }
    }
    
    // Затем логические разделы
    for key, group := range groups {
        if strings.HasPrefix(key, "LOGICAL:") {
            sortedDisks = append(sortedDisks, group...)
        }
    }
    
    // Остальные
    for key, group := range groups {
        if strings.HasPrefix(key, "OTHER:") {
            sortedDisks = append(sortedDisks, group...)
        }
    }
    
    return sortedDisks
}

// filterRemovedDisks удаляет из предыдущих данных диски, которых нет в текущих
func filterRemovedDisks(current, previous []DiskInfo) []DiskInfo {
    // Создаем карту текущих дисков для быстрого поиска
    currentMap := make(map[string]bool)
    for _, disk := range current {
        key := fmt.Sprintf("%s:%s", disk.Filesystem, disk.MountedOn)
        currentMap[key] = true
    }
    
    // Удаляем из предыдущих те, которых нет в текущих
    var filtered []DiskInfo
    for _, disk := range previous {
        key := fmt.Sprintf("%s:%s", disk.Filesystem, disk.MountedOn)
        if currentMap[key] {
            filtered = append(filtered, disk)
        }
    }
    
    return filtered
}

// stableSortDisks стабильно сортирует диски для отображения без скачков
func stableSortDisks(disks []DiskInfo) []DiskInfo {
    // Вместо map[string][]DiskInfo используем упорядоченную структуру
    type deviceGroup struct {
        baseDevice string
        disks      []DiskInfo
    }
    
    var groups []deviceGroup
    deviceMap := make(map[string]int)
    
    for _, disk := range disks {
        baseDevice := getBaseDevice(strings.TrimPrefix(disk.Filesystem, "/dev/"))
        if baseDevice == "" {
            baseDevice = "unknown"
        }
        
        idx, exists := deviceMap[baseDevice]
        if !exists {
            idx = len(groups)
            deviceMap[baseDevice] = idx
            groups = append(groups, deviceGroup{
                baseDevice: baseDevice,
                disks:      []DiskInfo{},
            })
        }
        groups[idx].disks = append(groups[idx].disks, disk)
    }
    
    // Стабильная сортировка групп по имени базового устройства
    sort.SliceStable(groups, func(i, j int) bool {
        return groups[i].baseDevice < groups[j].baseDevice
    })
    
    var sortedDisks []DiskInfo
    for _, group := range groups {
        // Стабильная сортировка дисков внутри группы
        sort.SliceStable(group.disks, func(i, j int) bool {
            // Сначала корневой раздел
            if group.disks[i].MountedOn == "/" && group.disks[j].MountedOn != "/" {
                return true
            }
            if group.disks[i].MountedOn != "/" && group.disks[j].MountedOn == "/" {
                return false
            }
            // Затем по алфавиту точек монтирования
            return group.disks[i].MountedOn < group.disks[j].MountedOn
        })
        sortedDisks = append(sortedDisks, group.disks...)
    }
    
    return sortedDisks
}

// Изменить сигнатуру функции для поддержки детального режима
func Summary(detailed bool) ([]DiskInfo, error) {
    diskCacheMutex.Lock()
    defer diskCacheMutex.Unlock()
    
    var disks []DiskInfo
    var err error
    
    switch runtime.GOOS {
    case "windows":
        disks, err = getWindowsDiskInfo(detailed)
    case "darwin":
        disks, err = getDarwinDiskInfo(detailed)
    default:
        disks, err = getLinuxDiskInfo(detailed)
    }

    // Если произошла ошибка или нет данных
    if err != nil || len(disks) == 0 {
        // Возвращаем предыдущие данные или заглушку
        if len(previousDisks) > 0 {
            return previousDisks, nil
        }
        return getFallbackDiskInfo(), nil
    }

    
    // Обновляем кэш
    previousDisks = disks

    // Применить фильтрацию в зависимости от режима
    if !detailed {
        disks = filterPrimaryDisks(disks)
    }

    // Гарантируем корректные значения UsePercent
    for i := range disks {
        if disks[i].UsePercent < 0 {
            disks[i].UsePercent = 0.0
        }
    }

    // СТАБИЛЬНАЯ СОРТИРОВКА для предотвращения скачков
    disks = stableSortDisks(disks)

    return disks, nil
}

// Обновленные функции с поддержкой детального режима

func getLinuxDiskInfo(detailed bool) ([]DiskInfo, error) {
    var disks []DiskInfo
    var err error

    // Базовое получение информации через df
    cmd := exec.Command("df", "-h")
    output, err := cmd.Output()
    if err != nil {
        return getFallbackDiskInfo(), nil
    }
    
    // Парсим вывод df
    parsedDisks, err := parseDFOutput(string(output), detailed)
    if err != nil {
        return nil, err
    }
    
    // В детальном режиме добавляем аппаратную информацию
    if detailed {
        // Получаем аппаратную информацию
        hardwareDevices, err := detectAllStorageDevices(detailed)
        if err == nil && len(hardwareDevices) > 0 {
            // Объединяем информацию из df и аппаратную информацию
            disks = mergeDiskInfo(parsedDisks, hardwareDevices)
        } else {
            // Если аппаратное обнаружение не удалось, используем только df
            disks = parsedDisks
        }
    } else {
        // В обычном режиме используем только отфильтрованные данные df
        disks = filterPrimaryDisks(parsedDisks)
    }
    
    // ДОБАВЛЕНО: Для детального режима определяем RAID-диски и их использование
    if detailed {
        disks = enhanceRaidDiskInfo(disks)
    }
    
    return disks, nil
}

func getWindowsDiskInfo(detailed bool) ([]DiskInfo, error) {
    var disks []DiskInfo
    
    // Получаем информацию о логических дисках через wmic
    cmd := exec.Command("wmic", "logicaldisk", "get", "size,freespace,caption,drivetype,volumename")
    output, err := cmd.Output()
    if err != nil {
        // Если wmic не работает, пытаемся получить информацию аппаратно
        if detailed {
            hardwareDevices, err := detectAllStorageDevices(detailed)
            if err == nil && len(hardwareDevices) > 0 {
                return hardwareDevices, nil
            }
            return getFallbackDiskInfo(), nil
        }
        return nil, err
    }
    
    // Парсим вывод wmic
    parsedDisks, err := parseWMICOutput(string(output), detailed)
    if err != nil {
        return nil, err
    }
    
    // В детальном режиме добавляем аппаратную информацию
    if detailed {
        // Получаем аппаратную информацию
        hardwareDevices, err := detectAllStorageDevices(detailed)
        if err == nil && len(hardwareDevices) > 0 {
            // Объединяем информацию
            disks = mergeDiskInfo(parsedDisks, hardwareDevices)
        } else {
            disks = parsedDisks
        }
    } else {
        // В обычном режиме используем только основные диски
        disks = filterPrimaryDisks(parsedDisks)
    }
    
    return disks, nil
}

func getDarwinDiskInfo(detailed bool) ([]DiskInfo, error) {
    var disks []DiskInfo
    
    // Базовое получение информации через df
    cmd := exec.Command("df", "-h")
    output, err := cmd.Output()
    if err != nil {
        // Если df не работает, пытаемся получить информацию аппаратно
        if detailed {
            hardwareDevices, err := detectAllStorageDevices(detailed)
            if err == nil && len(hardwareDevices) > 0 {
                return hardwareDevices, nil
            }
            return getFallbackDiskInfo(), nil
        }
        return nil, err
    }
    
    // Парсим вывод df
    parsedDisks, err := parseDFOutput(string(output), detailed)
    if err != nil {
        return nil, err
    }
    
    // В детальном режиме добавляем аппаратную информацию
    if detailed {
        // Получаем аппаратную информацию
        hardwareDevices, err := detectAllStorageDevices(detailed)
        if err == nil && len(hardwareDevices) > 0 {
            // Объединяем информацию
            disks = mergeDiskInfo(parsedDisks, hardwareDevices)
        } else {
            disks = parsedDisks
        }
    } else {
        // В обычном режиме используем только основные диски
        disks = filterPrimaryDisks(parsedDisks)
    }
    
    return disks, nil
}

// detectAllStorageDevices определяет все устройства хранения в зависимости от платформы
func detectAllStorageDevices(detailed bool) ([]DiskInfo, error) {
    if !detailed {
        // В недетальном режиме возвращаем пустой список
        return []DiskInfo{}, nil
    }
    
    switch runtime.GOOS {
    case "linux":
        // В hardware_linux.go есть функция detectAllStorageDevicesLinux
        return detectAllStorageDevicesLinux(detailed)
    case "windows":
        // В hardware_windows.go функция называется detectAllStorageDevicesWindows
        return detectAllStorageDevicesWindows(detailed)
    case "darwin":
        // В hardware_darwin.go функция называется detectAllStorageDevicesDarwin  
        return detectAllStorageDevicesDarwin(detailed)
    default:
        return []DiskInfo{}, fmt.Errorf("storage device detection not supported on %s", runtime.GOOS)
    }
}

// Новая функция для объединения информации из DF и аппаратных устройств
func mergeDiskInfo(dfDisks, hardwareDisks []DiskInfo) []DiskInfo {
    var mergedDisks []DiskInfo
    
    // Создаем карту аппаратных устройств для быстрого поиска
    hardwareMap := make(map[string]DiskInfo)
    for _, hw := range hardwareDisks {
        // Используем файловую систему или модель как ключ
        key := hw.Filesystem
        if key == "" || key == "Unknown" {
            key = hw.Model
        }
        if key != "" && key != "Unknown" {
            hardwareMap[key] = hw
        }
    }
    
    // Объединяем информацию из DF с аппаратными данными
    for _, dfDisk := range dfDisks {
        mergedDisk := dfDisk
        
        // Ищем соответствующее аппаратное устройство
        var hwKey string
        
        if strings.HasPrefix(dfDisk.Filesystem, "/dev/") {
            // Для Linux: извлекаем базовое устройство (sda1 -> sda)
            baseDevice := getBaseDevice(strings.TrimPrefix(dfDisk.Filesystem, "/dev/"))
            if baseDevice != "" {
                hwKey = "/dev/" + baseDevice
            }
        } else if runtime.GOOS == "windows" {
            // Для Windows ищем по модели
            hwKey = dfDisk.Model
        } else {
            // Для macOS используем файловую систему
            hwKey = dfDisk.Filesystem
        }
        
        if hwDisk, exists := hardwareMap[hwKey]; exists {
            // Обновляем поля аппаратной информации
            if mergedDisk.Model == "" || mergedDisk.Model == "Unknown" {
                if hwDisk.Model != "" && hwDisk.Model != "Unknown" {
                    mergedDisk.Model = hwDisk.Model
                }
            }
            
            if mergedDisk.Serial == "" || mergedDisk.Serial == "Unknown" {
                if hwDisk.Serial != "" && hwDisk.Serial != "Unknown" {
                    mergedDisk.Serial = hwDisk.Serial
                }
            }
            
            if mergedDisk.DiskType == "" || mergedDisk.DiskType == "Unknown" {
                if hwDisk.DiskType != "" && hwDisk.DiskType != "Unknown" {
                    mergedDisk.DiskType = hwDisk.DiskType
                }
            }
            
            if mergedDisk.SMARTStatus == "" || mergedDisk.SMARTStatus == "UNKNOWN" {
                if hwDisk.SMARTStatus != "" && hwDisk.SMARTStatus != "UNKNOWN" {
                    mergedDisk.SMARTStatus = hwDisk.SMARTStatus
                }
            }
            
            if hwDisk.Vendor != "" && hwDisk.Vendor != "Unknown" {
                mergedDisk.Vendor = hwDisk.Vendor
            }
            
            if hwDisk.Physical {
                mergedDisk.Physical = true
            }
            
            // Удаляем устройство из карты, чтобы не дублировать
            delete(hardwareMap, hwKey)
        }
        
        mergedDisks = append(mergedDisks, mergedDisk)
    }
    
    // Добавляем оставшиеся аппаратные устройства (несмонтированные)
    for _, hwDisk := range hardwareMap {
        // Для несмонтированных устройств создаем минимальную информацию
        if hwDisk.Filesystem == "" {
            hwDisk.Filesystem = "(unmounted)"
        }
        if hwDisk.MountedOn == "" {
            hwDisk.MountedOn = "(unmounted)"
        }
        if hwDisk.Size == "" || hwDisk.Size == "0 B" {
            hwDisk.Size = "Unknown"
        }
        
        // Устанавливаем флаг несмонтированного устройства
        hwDisk.UsePercent = 0.0
        hwDisk.Used = "0 B"
        hwDisk.Available = hwDisk.Size
        
        mergedDisks = append(mergedDisks, hwDisk)
    }
    
    return mergedDisks
}

// filterPrimaryDisks возвращает только основные диски
func filterPrimaryDisks(disks []DiskInfo) []DiskInfo {
    var primaryDisks []DiskInfo
    
    for _, disk := range disks {
        if isPrimaryDisk(disk) {
            primaryDisks = append(primaryDisks, disk)
        }
    }
    
    // Если не найдено основных дисков, возвращаем первые 3
    if len(primaryDisks) == 0 && len(disks) > 0 {
        maxDisks := len(disks)
        if maxDisks > 3 {
            maxDisks = 3
        }
        return disks[:maxDisks]
    }
    
    return primaryDisks
}

// isPrimaryDisk определяет, является ли диск основным
func isPrimaryDisk(disk DiskInfo) bool {
    // Критерии для разных ОС
    switch runtime.GOOS {
    case "linux":
        // Основные точки монтирования
        primaryMounts := []string{"/", "/home", "/boot", "/var", "/usr"}
        for _, mount := range primaryMounts {
            if disk.MountedOn == mount {
                return true
            }
        }
        // Исключаем временные файловые системы, НО включаем флешки в /media/ и /mnt/
        if strings.HasPrefix(disk.Filesystem, "tmpfs") || 
           strings.HasPrefix(disk.Filesystem, "devtmpfs") ||
           strings.HasPrefix(disk.Filesystem, "overlay") {
            return false
        }
        // Физические диски считаем основными (включая флешки)
        return disk.Physical
        
    case "windows":
        // Основные системные диски
        if disk.MountedOn == "C:" || disk.MountedOn == "D:" {
            return true
        }
        // Исключаем сетевые диски, НО включаем съемные
        if disk.DiskType == "Network" || strings.HasPrefix(disk.Filesystem, "\\\\") {
            return false
        }
        // Физические диски и съемные
        return disk.Physical
        
    case "darwin":
        // Основные точки монтирования macOS
        primaryMounts := []string{"/", "/System", "/Users", "/Volumes/Macintosh HD"}
        for _, mount := range primaryMounts {
            if disk.MountedOn == mount {
                return true
            }
        }
        // Исключаем временные и сетевые
        if strings.Contains(disk.MountedOn, "/Volumes/") && !disk.Physical {
            return false
        }
        return disk.Physical
        
    default:
        // По умолчанию возвращаем физические диски
        return disk.Physical
    }
}

func getFallbackDiskInfo() []DiskInfo {
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
    return []DiskInfo{defaultDisk}
}

// Остальные существующие функции остаются без изменений...

func parseDFOutput(output string, detailed bool) ([]DiskInfo, error) {
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

            // В детальном режиме не пропускаем файловые системы
            if !detailed && shouldSkipFilesystem(fields[0], fields[5]) {
                continue
            }

            // Определяем тип устройства и физические характеристики
            physical, diskType, model, serial, smartStatus, vendor := detectDiskProperties(fields[0])

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
                Vendor:     vendor, // ДОБАВЛЕНО: поле Vendor
            }
            disks = append(disks, disk)
        }
    }

    return disks, nil
}

func parseWMICOutput(output string, detailed bool) ([]DiskInfo, error) {
    lines := strings.Split(output, "\n")
    var disks []DiskInfo

    for i, line := range lines {
        if i == 0 || strings.TrimSpace(line) == "" {
            continue
        }

        re := regexp.MustCompile(`\s+`)
        fields := re.Split(line, -1)

        if len(fields) >= 4 {
            // Поле drivetype: 2=съемный, 3=локальный HDD, 4=сеть, 5=CD-ROM
            driveType := 3
            if len(fields) >= 4 {
                if dt, err := strconv.Atoi(fields[3]); err == nil {
                    driveType = dt
                }
            }

            // В недетальном режиме пропускаем съемные и сетевые диски
            if !detailed && (driveType == 2 || driveType == 4 || driveType == 5) {
                continue
            }

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

            // Для Windows определяем тип устройства - исправлено: добавлена переменная vendor
            physical, diskType, model, serial, smartStatus, vendor := detectWindowsDiskProperties(fields[0], driveType)

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
                Vendor:     vendor, // Исправлено: используем полученное значение vendor
            }
            disks = append(disks, disk)
        }
    }

    return disks, nil
}

// detectDiskProperties определяет свойства диска для Linux/Unix систем
func detectDiskProperties(filesystem string) (bool, string, string, string, string, string) {
    physical := true
    diskType := "Unknown"
    model := "Unknown"
    serial := "Unknown"
    smartStatus := "UNKNOWN"
    vendor := "Unknown"

    // Улучшенное определение RAID
    if strings.HasPrefix(filesystem, "/dev/md") {
        diskType = "RAID"
        physical = false
        
        // Получаем детальную информацию о RAID
        if raidInfo := getRaidDetails(filesystem); raidInfo != "" {
            model = raidInfo
        } else {
            model = "RAID Array"
        }
    }
    
    // Извлекаем имя устройства из пути (например, /dev/sda1 -> sda)
    device := strings.TrimPrefix(filesystem, "/dev/")
    if device == "" {
        return physical, diskType, model, serial, smartStatus, vendor
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
        devicePath := filepath.Join("/sys/block", baseDevice)
        
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

        // Получаем модель из sys (резервный метод)
        if model == "" || model == "Unknown" {
            modelPath := filepath.Join(devicePath, "device", "model")
            if data, err := os.ReadFile(modelPath); err == nil {
                model = strings.TrimSpace(string(data))
            }
        }

        // Получаем SMART-статус
        smartStatus = getSMARTStatus(baseDevice)

        // Если это LVM, RAID или другие виртуальные устройства, помечаем как логические
        if diskType == "LVM" || strings.Contains(diskType, "RAID") {
            physical = false
        }
    }

    // Получаем вендора по модели
    if model != "" && model != "Unknown" {
        vendor = getDiskVendor(model)
    }

    return physical, diskType, model, serial, smartStatus, vendor
}

// getRaidDetails получает детальную информацию о RAID массиве
func getRaidDetails(device string) string {
    deviceName := strings.TrimPrefix(device, "/dev/")
    
    if runtime.GOOS == "linux" {
        // Получаем уровень RAID и компоненты
        cmd := exec.Command("mdadm", "--detail", "--brief", "/dev/"+deviceName)
        if output, err := cmd.Output(); err == nil {
            lines := strings.Split(string(output), "\n")
            var level, components string
            
            for _, line := range lines {
                if strings.Contains(line, "level=") {
                    parts := strings.Split(line, "level=")
                    if len(parts) > 1 {
                        level = strings.Fields(parts[1])[0]
                    }
                }
                if strings.Contains(line, "devices=") {
                    parts := strings.Split(line, "devices=")
                    if len(parts) > 1 {
                        components = strings.Fields(parts[1])[0]
                    }
                }
            }
            
            if level != "" {
                if components != "" {
                    return fmt.Sprintf("RAID %s (%s drives)", level, components)
                }
                return fmt.Sprintf("RAID %s", level)
            }
        }
        
        // Альтернативно: проверяем через sysfs
        levelPath := fmt.Sprintf("/sys/block/%s/md/level", deviceName)
        if data, err := os.ReadFile(levelPath); err == nil {
            level := strings.TrimSpace(string(data))
            return fmt.Sprintf("RAID %s", level)
        }
    }
    
    return "RAID Array"
}

// detectWindowsDiskProperties определяет свойства диска для Windows систем
func detectWindowsDiskProperties(drive string, driveType int) (bool, string, string, string, string, string) {
    physical := true
    diskType := "Unknown"
    model := "Unknown"
    serial := "Unknown"
    smartStatus := "UNKNOWN"
    vendor := "Unknown" // ДОБАВЛЕНО: переменная vendor

    // Определяем тип по driveType
    switch driveType {
    case 2:
        diskType = "Removable"
        physical = false
    case 3:
        diskType = "Local Disk"
        physical = true
    case 4:
        diskType = "Network"
        physical = false
    case 5:
        diskType = "CD-ROM"
        physical = false
    }

    // Получаем дополнительную информацию через wmic
    if diskType == "Local Disk" || diskType == "Removable" {
        cmd := exec.Command("wmic", "diskdrive", "get", "Model,SerialNumber,MediaType,InterfaceType,Manufacturer", "/format:list")
        output, err := cmd.Output()
        if err == nil {
            lines := strings.Split(string(output), "\n")
            var currentModel, currentSerial, currentManufacturer string
            
            for _, line := range lines {
                line = strings.TrimSpace(line)
                if strings.HasPrefix(line, "Model=") {
                    currentModel = strings.TrimPrefix(line, "Model=")
                } else if strings.HasPrefix(line, "SerialNumber=") {
                    currentSerial = strings.TrimPrefix(line, "SerialNumber=")
                } else if strings.HasPrefix(line, "MediaType=") {
                    mediaType := strings.TrimPrefix(line, "MediaType=")
                    switch mediaType {
                    case "Fixed hard disk media":
                        diskType = "HDD"
                    case "SSD":
                        diskType = "SSD"
                    case "External hard disk media":
                        diskType = "External"
                    }
                } else if strings.HasPrefix(line, "Manufacturer=") {
                    currentManufacturer = strings.TrimPrefix(line, "Manufacturer=")
                }
            }
            
            if currentModel != "" {
                model = currentModel
            }
            if currentSerial != "" {
                serial = currentSerial
            }
            if currentManufacturer != "" {
                vendor = currentManufacturer
            }
        }
    }

    // Сетевые диски
    if strings.HasPrefix(drive, "\\\\") {
        diskType = "Network"
        physical = false
    }

    return physical, diskType, model, serial, smartStatus, vendor // ДОБАВЛЕНО: возвращаем vendor
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

    // Пропускаем специальные точки монтирования, но НЕ пропускаем /media/ и /mnt/ для флешек
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

// add this function to clean up model names
func cleanModelName(model string) string {
    if model == "" {
        return "Unknown"
    }
    
    // Удаляем непечатаемые символы и лишние пробелы
    model = strings.Map(func(r rune) rune {
        // Удаляем управляющие символы (0x00-0x1F, 0x7F)
        if r <= 0x1F || r == 0x7F {
            return -1
        }
        // Заменяем нестандартные пробелы на обычные
        if r == 0xA0 || r == 0x3000 || r == 0x202F {
            return ' '
        }
        return r
    }, model)
    
    // Удаляем лишние пробелы
    model = strings.TrimSpace(model)
    model = regexp.MustCompile(`\s+`).ReplaceAllString(model, " ")
    
    return model
}


// getDiskVendor определяет вендора по модели диска
func getDiskVendor(model string) string {
    if model == "" || model == "Unknown" {
        return "Unknown"
    }
    
    // Сначала очищаем модель
    cleanModel := cleanModelName(model)
    modelLower := strings.ToLower(cleanModel)
    
    // Определяем вендора по ключевым словам в модели
    // Важно: сначала более специфичные совпадения
    
    // Western Digital - самый приоритетный для WD дисков
    if strings.Contains(modelLower, "wdc") || strings.Contains(modelLower, "western digital") || 
       strings.HasPrefix(modelLower, "wd ") || strings.HasPrefix(modelLower, "wd") ||
       strings.Contains(modelLower, "wd5000") || strings.Contains(modelLower, "wd ") {
        return "Western Digital"
    }
    
    // Crucial - только если явно указан
    if strings.Contains(modelLower, "crucial") && !strings.Contains(modelLower, "wdc") {
        return "Crucial"
    }
    
    // Intel - только если явно указан
    if strings.Contains(modelLower, "intel") && !strings.Contains(modelLower, "wdc") {
        return "Intel"
    }
    
    // Остальные вендоры...
    vendors := map[string][]string{
        "Seagate": {"seagate", "st", "barracuda", "ironwolf", "skyhawk"},
        "Western Digital": {"western digital", "wd ", "my passport", "elements", "wd5000", "wdc"},
        "Toshiba": {"toshiba", "mq", "dt", "canvio"},
        "Hitachi": {"hitachi", "hgst", "ultrastar"},
        "Samsung": {"samsung", "ssd", "evo", "pro", "qvo"},
        "Crucial": {"crucial", "mx", "bx", "p"},
        "Kingston": {"kingston", "kc", "a400", "suv"},
        "SanDisk": {"sandisk", "ultra", "extreme", "plus"},
        "Intel": {"intel", "optane", "dc"},
        "Micron": {"micron", "mt", "p", "m600"},
        "ADATA": {"adata", "sx", "su", "xpg"},
        "KingSpec": {"kingspec"},
        "Transcend": {"transcend"},
        "PNY": {"pny"},
        "Gigabyte": {"gigabyte", "aorus"},
        "ASUS": {"asus", "rog"},
        "MSI": {"msi"},
        "Corsair": {"corsair", "force", "mp"},
        "Team Group": {"team group", "team"},
        "Patriot": {"patriot", "burst", "p210"},
        "Lexar": {"lexar", "ns"},
        "Apacer": {"apacer"},
        "Silicon Power": {"silicon power", "sp"},
        "GoodRAM": {"goodram"},
        "Netac": {"netac"},
        "Colorful": {"colorful"},
        "Galax": {"galax"},
        "Hikvision": {"hikvision"},
        "Dell": {"dell", "poweredge"},
        "HP": {"hp ", "hewlett packard"},
        "IBM": {"ibm", "lenovo"},
        "Fujitsu": {"fujitsu"},
        "Maxtor": {"maxtor"},
    }
    
    for vendor, keywords := range vendors {
        for _, keyword := range keywords {
            if strings.Contains(modelLower, keyword) {
                return vendor
            }
        }
    }
    
    return "Unknown"
}

// НОВАЯ ФУНКЦИЯ: Определяет, является ли диск частью RAID
func isDiskPartOfRaid(device string) (bool, string) {
    if runtime.GOOS != "linux" {
        return false, ""
    }
    
    // Проверяем, является ли устройство частью RAID через mdadm
    cmd := exec.Command("mdadm", "--detail", "--scan")
    output, err := cmd.Output()
    if err != nil {
        return false, ""
    }
    
    lines := strings.Split(string(output), "\n")
    for _, line := range lines {
        if strings.Contains(line, device) {
            // Извлекаем имя RAID-устройства
            re := regexp.MustCompile(`(/dev/md\d+)`)
            if matches := re.FindStringSubmatch(line); len(matches) > 1 {
                return true, matches[1]
            }
        }
    }
    
    // Альтернативная проверка через sysfs
    raidPath := fmt.Sprintf("/sys/block/%s/md", device)
    if _, err := os.Stat(raidPath); err == nil {
        // Устройство является частью RAID
        // Пытаемся определить имя RAID-устройства
        if links, err := filepath.Glob(fmt.Sprintf("/sys/block/md*/slaves/%s", device)); err == nil && len(links) > 0 {
            for _, link := range links {
                parts := strings.Split(link, "/")
                for _, part := range parts {  // ИСПРАВЛЕНО: убрана неиспользуемая переменная i
                    if strings.HasPrefix(part, "md") {
                        return true, "/dev/" + part
                    }
                }
            }
        }
    }
    
    return false, ""
}

// НОВАЯ ФУНКЦИЯ: Получает использование RAID-массива
func getRaidUsage(raidDevice string) (float64, string, string) {
    if runtime.GOOS != "linux" {
        return 0.0, "", ""
    }
    
    // Получаем размер RAID через blockdev
    cmd := exec.Command("blockdev", "--getsize64", raidDevice)
    sizeOutput, err := cmd.Output()
    if err != nil {
        return 0.0, "", ""
    }
    
    sizeStr := strings.TrimSpace(string(sizeOutput))
    sizeBytes, err := strconv.ParseUint(sizeStr, 10, 64)
    if err != nil {
        return 0.0, "", ""
    }
    
    // Пытаемся получить использование через df
    cmd = exec.Command("df", "-B1", "--output=used,size", raidDevice)
    dfOutput, err := cmd.Output()
    if err != nil {
        // Если df не сработал, возможно RAID не смонтирован
        return 0.0, formatBytes(sizeBytes), formatBytes(0)
    }
    
    lines := strings.Split(string(dfOutput), "\n")
    if len(lines) >= 2 {
        fields := strings.Fields(lines[1])
        if len(fields) >= 2 {
            usedBytes, _ := strconv.ParseUint(fields[0], 10, 64)
            totalBytes, _ := strconv.ParseUint(fields[1], 10, 64)
            if totalBytes > 0 {
                usePercent := float64(usedBytes) / float64(totalBytes) * 100
                return usePercent, formatBytes(usedBytes), formatBytes(totalBytes)
            }
        }
    }
    
    return 0.0, formatBytes(sizeBytes), formatBytes(0)
}

// НОВАЯ ФУНКЦИЯ: Улучшает информацию о RAID-дисках
// НОВАЯ ФУНКЦИЯ: Улучшает информацию о RAID-дисках
func enhanceRaidDiskInfo(disks []DiskInfo) []DiskInfo {
    // Создаем карту RAID-устройств и их использования
    raidUsageMap := make(map[string]struct {
        usePercent float64
        used       string
        size       string
    })
    
    // Сначала собираем информацию о RAID-разделах (смонтированных)
    for _, disk := range disks {
        if strings.HasPrefix(disk.Filesystem, "/dev/md") && disk.MountedOn != "(unmounted)" && !strings.Contains(disk.MountedOn, "unmounted") {
            // Это смонтированный RAID-раздел - у него есть данные об использовании
            raidUsageMap[disk.Filesystem] = struct {
                usePercent float64
                used       string
                size       string
            }{
                usePercent: disk.UsePercent,
                used:       disk.Used,
                size:       disk.Size,
            }
            
            // Также добавляем базовое устройство RAID (без номера раздела)
            baseDevice := getBaseDevice(strings.TrimPrefix(disk.Filesystem, "/dev/"))
            if baseDevice != "" && !strings.Contains(baseDevice, "p") {
                // Только если это действительно базовое устройство (не раздел)
                raidKey := "/dev/" + baseDevice
                raidUsageMap[raidKey] = struct {
                    usePercent float64
                    used       string
                    size       string
                }{
                    usePercent: disk.UsePercent,
                    used:       disk.Used,
                    size:       disk.Size,
                }
            }
        }
    }
    
    // Теперь проверяем все диски на принадлежность к RAID
    for i := range disks {
        disk := &disks[i]
        device := strings.TrimPrefix(disk.Filesystem, "/dev/")
        
        // Пропускаем уже смонтированные RAID-разделы
        if strings.HasPrefix(disk.Filesystem, "/dev/md") && disk.MountedOn != "(unmounted)" && !strings.Contains(disk.MountedOn, "unmounted") {
            continue
        }
        
        // Проверяем, является ли диск частью RAID
        if isRaid, raidDevice := isDiskPartOfRaid(device); isRaid {
            // Обновляем информацию о диске
            disk.Physical = true
            disk.DiskType = "RAID Component"
            
            // Если у нас есть информация об использовании RAID
            if raidUsage, exists := raidUsageMap[raidDevice]; exists {
                disk.UsePercent = raidUsage.usePercent
                disk.Used = raidUsage.used
                disk.Size = raidUsage.size
                // Рассчитываем доступное пространство
                if usedMB := extractMemoryMB(raidUsage.used); usedMB > 0 {
                    if totalMB := extractMemoryMB(raidUsage.size); totalMB > 0 {
                        availableMB := totalMB - usedMB
                        disk.Available = formatBytes(uint64(availableMB) * 1024 * 1024)
                    }
                }
            } else {
                // Пытаемся получить использование RAID
                if usePercent, used, size := getRaidUsage(raidDevice); usePercent > 0 {
                    disk.UsePercent = usePercent
                    disk.Used = used
                    disk.Size = size
                    // Рассчитываем доступное пространство
                    if usedMB := extractMemoryMB(used); usedMB > 0 {
                        if totalMB := extractMemoryMB(size); totalMB > 0 {
                            availableMB := totalMB - usedMB
                            disk.Available = formatBytes(uint64(availableMB) * 1024 * 1024)
                        }
                    }
                }
            }
        }
    }
    
    // Теперь обновляем информацию для несмонтированных RAID-массивов
    for i := range disks {
        disk := &disks[i]
        
        // Пропускаем не RAID устройства
        if !strings.HasPrefix(disk.Filesystem, "/dev/md") {
            continue
        }
        
        // Пропускаем уже обработанные (с использованием > 0)
        if disk.UsePercent > 0 {
            continue
        }
        
        // Пропускаем смонтированные разделы
        if disk.MountedOn != "(unmounted)" && !strings.Contains(disk.MountedOn, "unmounted") && disk.MountedOn != "" {
            continue
        }
        
        // Проверяем, есть ли использование для этого RAID-массива
        device := strings.TrimPrefix(disk.Filesystem, "/dev/")
        
        // Ищем использование в карте
        for raidKey, raidUsage := range raidUsageMap {
            raidDevice := strings.TrimPrefix(raidKey, "/dev/")
            // Проверяем, является ли это использование для этого RAID-массива
            if strings.HasPrefix(raidDevice, device) || device == getBaseDevice(raidDevice) {
                disk.UsePercent = raidUsage.usePercent
                disk.Used = raidUsage.used
                disk.Size = raidUsage.size
                // Рассчитываем доступное пространство
                if usedMB := extractMemoryMB(raidUsage.used); usedMB > 0 {
                    if totalMB := extractMemoryMB(raidUsage.size); totalMB > 0 {
                        availableMB := totalMB - usedMB
                        disk.Available = formatBytes(uint64(availableMB) * 1024 * 1024)
                    }
                }
                break
            }
        }
    }
    
    return disks
}

// НОВАЯ ФУНКЦИЯ: Извлекает числовое значение памяти в MB
func extractMemoryMB(memoryStr string) int {
    // Пример: "8192 MB" -> 8192, "1.0 GB" -> 1024
    parts := strings.Fields(memoryStr)
    if len(parts) < 2 {
        return 0
    }
    
    valueStr := parts[0]
    unit := strings.ToUpper(parts[1])
    
    value, err := strconv.ParseFloat(valueStr, 64)
    if err != nil {
        return 0
    }
    
    switch {
    case strings.Contains(unit, "GB"):
        return int(value * 1024)
    case strings.Contains(unit, "MB"):
        return int(value)
    case strings.Contains(unit, "KB"):
        return int(value / 1024)
    case strings.Contains(unit, "B"):
        return int(value / (1024 * 1024))
    case strings.Contains(unit, "TB"):
        return int(value * 1024 * 1024)
    default:
        return int(value)
    }
}
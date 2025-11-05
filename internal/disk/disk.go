package disk

import (
	"os/exec"
	"regexp"
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
	// Use df command to get disk information
	cmd := exec.Command("df", "-h")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return parseDFOutput(string(output))
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
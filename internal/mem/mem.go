package mem

import (
	"os/exec"
	"regexp"
	//"strconv" удалено
	"strings"
)

type MemoryInfo struct {
	Total     string
	Used      string
	Free      string
	Available string
	SwapTotal string
	SwapUsed  string
	SwapFree  string
}

func Summary() (*MemoryInfo, error) {
	// Use free command to get memory information
	cmd := exec.Command("free", "-h")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return parseFreeOutput(string(output))
}

func parseFreeOutput(output string) (*MemoryInfo, error) {
	lines := strings.Split(output, "\n")
	info := &MemoryInfo{}

	// Parse memory line
	if len(lines) >= 2 {
		memFields := parseFreeLine(lines[1])
		if len(memFields) >= 7 {
			info.Total = memFields[1]
			info.Used = memFields[2]
			info.Free = memFields[3]
			info.Available = memFields[6]
		}
	}

	// Parse swap line
	if len(lines) >= 3 {
		swapFields := parseFreeLine(lines[2])
		if len(swapFields) >= 5 {
			info.SwapTotal = swapFields[1]
			info.SwapUsed = swapFields[2]
			info.SwapFree = swapFields[3]
		}
	}

	return info, nil
}

func parseFreeLine(line string) []string {
	re := regexp.MustCompile(`\s+`)
	fields := re.Split(strings.TrimSpace(line), -1)
	return fields
}
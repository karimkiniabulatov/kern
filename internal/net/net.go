package net

import (
	"os/exec"
	"regexp"
	"strings"
)

type NetworkInfo struct {
	Interface string
	RXBytes   string
	TXBytes   string
	Status    string
}

func Summary() ([]NetworkInfo, error) {
	// Use ip command to get network information
	cmd := exec.Command("ip", "-s", "link")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return parseIPOutput(string(output))
}

func parseIPOutput(output string) ([]NetworkInfo, error) {
	var interfaces []NetworkInfo
	lines := strings.Split(output, "\n")

	var currentInterface *NetworkInfo

	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		// Look for interface lines
		if strings.Contains(line, ":") && !strings.Contains(line, "RX:") && !strings.Contains(line, "TX:") {
			if currentInterface != nil {
				interfaces = append(interfaces, *currentInterface)
			}

			// Parse interface name and status
			parts := strings.SplitN(line, ":", 3)
			if len(parts) >= 3 {
				interfaceName := strings.TrimSpace(parts[1])
				status := "DOWN"
				if strings.Contains(parts[2], "state UP") {
					status = "UP"
				}

				currentInterface = &NetworkInfo{
					Interface: interfaceName,
					Status:    status,
				}
			}
		}

		// Look for RX bytes
		if currentInterface != nil && strings.Contains(line, "RX:") && i+1 < len(lines) {
			rxLine := strings.TrimSpace(lines[i+1])
			rxBytes := extractBytes(rxLine)
			if rxBytes != "" {
				currentInterface.RXBytes = rxBytes
			}
		}

		// Look for TX bytes
		if currentInterface != nil && strings.Contains(line, "TX:") && i+1 < len(lines) {
			txLine := strings.TrimSpace(lines[i+1])
			txBytes := extractBytes(txLine)
			if txBytes != "" {
				currentInterface.TXBytes = txBytes
			}
		}
	}

	// Add the last interface
	if currentInterface != nil {
		interfaces = append(interfaces, *currentInterface)
	}

	return interfaces, nil
}

func extractBytes(line string) string {
	re := regexp.MustCompile(`(\d+)\s*(\w*B)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) >= 3 {
		return matches[1] + " " + matches[2]
	}
	return ""
}
package testutil

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ParseOnlineConfFile parses OnlineConf configuration file and returns a map of environment variables
// Example: /xvpnback/bot/button/faq FAQ -> OC_xvpnback__bot__button__faq: "FAQ"
func ParseOnlineConfFile(filePath string) (map[string]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	result := make(map[string]string)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Skip comments
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}

		// Find first space or tab to separate path from value
		separatorIdx := -1
		for i, ch := range line {
			if ch == ' ' || ch == '\t' {
				separatorIdx = i
				break
			}
		}

		var path, value string
		if separatorIdx == -1 {
			// No separator found, entire line is the path with empty value
			path = line
			value = ""
		} else {
			path = line[:separatorIdx]
			value = strings.TrimLeft(line[separatorIdx+1:], " \t")
		}

		// Convert path to environment variable name
		// Remove leading /
		path = strings.TrimPrefix(path, "/")

		// Replace / with __
		envKey := strings.ReplaceAll(path, "/", "__")

		// Add OC_ prefix
		envKey = "OC_" + envKey

		result[envKey] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return result, nil
}

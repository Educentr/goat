package testutil

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// LoadEnvFile parses a .env file and returns a map of environment variables
// Note: We use a custom parser instead of godotenv because our variable names may contain hyphens
func LoadEnvFile(filePath string) (map[string]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	result := make(map[string]string)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split by first '=' to separate key and value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present and unescape internal quotes
		if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
			value = value[1 : len(value)-1]
			// Unescape quotes that were escaped by WriteEnvFile
			value = strings.ReplaceAll(value, "\\\"", "\"")
		}

		result[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return result, nil
}

// WriteEnvFile writes environment variables map to a .env file
func WriteEnvFile(filePath string, envVars map[string]string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	for key, value := range envVars {
		// Escape quotes in value
		escapedValue := strings.ReplaceAll(value, "\"", "\\\"")
		if _, err := fmt.Fprintf(file, "%s=\"%s\"\n", key, escapedValue); err != nil {
			return fmt.Errorf("failed to write to file: %w", err)
		}
	}

	return nil
}

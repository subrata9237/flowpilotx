package config

import (
    "bufio"
    "fmt"
    "os"
    "strings"
)

// LoadDotNetEnv loads environment variables from a .dotnetenv file
func LoadDotNetEnv(filename string) error {
    file, err := os.Open(filename)
    if err != nil {
        return fmt.Errorf("error opening .dotnetenv file: %w", err)
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())

        // Skip empty lines and comments
        if line == "" || strings.HasPrefix(line, "#") {
            continue
        }

        // Split on first equals sign
        parts := strings.SplitN(line, "=", 2)
        if len(parts) != 2 {
            continue
        }

        key := strings.TrimSpace(parts[0])
        value := strings.TrimSpace(parts[1])

        // Remove quotes if present
        value = strings.Trim(value, `"'`)

        // Set environment variable
        if err := os.Setenv(key, value); err != nil {
            return fmt.Errorf("error setting environment variable %s: %w", key, err)
        }
    }

    if err := scanner.Err(); err != nil {
        return fmt.Errorf("error reading .dotnetenv file: %w", err)
    }

    return nil
} 
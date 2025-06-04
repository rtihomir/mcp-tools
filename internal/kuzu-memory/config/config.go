// File: internal/kuzu-memory/config/config.go
package config

import (
	"fmt"
	"os"
)

// Config holds server configuration
type Config struct {
	DatabasePath string
	ReadOnly     bool
}

// ParseConfig parses configuration from command line args and environment variables
// Mimics the JS version logic exactly
func ParseConfig() (*Config, error) {
	var dbPath string

	// Check command line arguments first
	args := os.Args[1:]
	if len(args) == 0 {
		// No command line args, check environment variable
		envDbPath := os.Getenv("KUZU_DB_PATH")
		if envDbPath != "" {
			dbPath = envDbPath
		} else {
			return nil, fmt.Errorf("please provide a path to kuzu database as a command line argument")
		}
	} else {
		dbPath = args[0]
	}

	// Check read-only flag from environment (exactly like JS)
	isReadOnly := os.Getenv("KUZU_READ_ONLY") == "true"

	return &Config{
		DatabasePath: dbPath,
		ReadOnly:     isReadOnly,
	}, nil
}

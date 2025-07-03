package resources

import (
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rtihomir/mcp-tools/internal/duckdb/database"
	"github.com/rtihomir/mcp-tools/internal/duckdb/state"
)

// HomeDirectoryListing represents the structure of the home directory resource
type HomeDirectoryListing struct {
	HomeDir        string                     `json:"home_dir"`
	AvailableFiles map[string][]string        `json:"available_files"`
	TotalFiles     int                        `json:"total_files"`
}

// HandleListHomeDir handles the ListHomeDir resource request
func HandleListHomeDir(session *state.Session) (string, error) {
	homeDir := session.GetHomeDir()
	if homeDir == "" {
		return "", fmt.Errorf("no home directory configured. Use the 'configure' tool to set a home directory first")
	}

	// List files in the home directory
	files, err := database.ListSupportedFiles(homeDir)
	if err != nil {
		return "", fmt.Errorf("failed to list files in home directory: %w", err)
	}

	// Count total files
	totalFiles := 0
	for _, fileList := range files {
		totalFiles += len(fileList)
	}

	// Build response
	listing := HomeDirectoryListing{
		HomeDir:        homeDir,
		AvailableFiles: files,
		TotalFiles:     totalFiles,
	}

	// Convert to JSON
	jsonBytes, err := json.MarshalIndent(listing, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format response: %w", err)
	}

	return string(jsonBytes), nil
}

// GetListHomeDirResourceDefinition returns the MCP resource definition
func GetListHomeDirResourceDefinition() mcp.Resource {
	return mcp.Resource{
		URI:         "duckdb://home-directory",
		Name:        "Home Directory Listing",
		Description: "Lists available database and data files in the configured home directory",
		MIMEType:    "application/json",
	}
}
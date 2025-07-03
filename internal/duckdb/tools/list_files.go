package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rtihomir/mcp-tools/internal/duckdb/database"
	"github.com/rtihomir/mcp-tools/internal/duckdb/state"
)

// ListFilesResponse represents the response from the list_files tool
type ListFilesResponse struct {
	Success        bool                        `json:"success"`
	HomeDir        string                      `json:"home_dir"`
	AvailableFiles map[string][]string        `json:"available_files"`
	TotalFiles     int                        `json:"total_files"`
	Message        string                      `json:"message"`
}

// HandleListFiles handles the list_files tool request
func HandleListFiles(ctx context.Context, request mcp.CallToolRequest, session *state.Session) (*mcp.CallToolResult, error) {
	homeDir := session.GetHomeDir()
	if homeDir == "" {
		return mcp.NewToolResultError("No home directory configured. Use the 'configure' tool to set a home directory first."), nil
	}

	// List files in the home directory
	files, err := database.ListSupportedFiles(homeDir)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list files in home directory: %v", err)), nil
	}

	// Count total files
	totalFiles := 0
	for _, fileList := range files {
		totalFiles += len(fileList)
	}

	// Build response
	response := ListFilesResponse{
		Success:        true,
		HomeDir:        homeDir,
		AvailableFiles: files,
		TotalFiles:     totalFiles,
		Message:        fmt.Sprintf("Found %d supported files in %s", totalFiles, homeDir),
	}

	// Convert response to JSON
	responseBytes, err := json.Marshal(response)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to format response: %v", err)), nil
	}

	return mcp.NewToolResultText(string(responseBytes)), nil
}

// GetListFilesToolDefinition returns the MCP tool definition for list_files
func GetListFilesToolDefinition() mcp.Tool {
	return mcp.NewTool("list_files",
		mcp.WithDescription("List available database and data files in the configured home directory"),
	)
}
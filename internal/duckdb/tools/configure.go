package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rtihomir/mcp-tools/internal/duckdb/database"
	"github.com/rtihomir/mcp-tools/internal/duckdb/state"
)

// ConfigureRequest represents the parameters for the configure tool
type ConfigureRequest struct {
	DBPath   string `json:"db_path,omitempty"`
	HomeDir  string `json:"home_dir,omitempty"`
	ReadOnly bool   `json:"read_only,omitempty"`
}

// ConfigureResponse represents the response from the configure tool
type ConfigureResponse struct {
	Success       bool                        `json:"success"`
	Message       string                      `json:"message"`
	Connected     bool                        `json:"connected"`
	DBPath        string                      `json:"db_path,omitempty"`
	HomeDir       string                      `json:"home_dir,omitempty"`
	ReadOnly      bool                        `json:"read_only"`
	AvailableFiles map[string][]string       `json:"available_files,omitempty"`
}

// HandleConfigure handles the configure tool request
func HandleConfigure(ctx context.Context, request mcp.CallToolRequest, session *state.Session) (*mcp.CallToolResult, error) {
	// Parse the request arguments
	var req ConfigureRequest
	if request.Params.Arguments != nil {
		if args, ok := request.Params.Arguments.(map[string]interface{}); ok {
			if err := mapToStruct(args, &req); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Invalid arguments: %v", err)), nil
			}
		} else {
			return mcp.NewToolResultError("Invalid argument format"), nil
		}
	}

	// Validate that at least one of db_path or home_dir is provided
	if req.DBPath == "" && req.HomeDir == "" {
		return mcp.NewToolResultError("Either 'db_path' or 'home_dir' must be provided"), nil
	}

	// Validate paths exist
	if req.DBPath != "" && req.DBPath != ":memory:" {
		if _, err := os.Stat(req.DBPath); err != nil {
			if os.IsNotExist(err) {
				return mcp.NewToolResultError(fmt.Sprintf("Database file does not exist: %s", req.DBPath)), nil
			}
			return mcp.NewToolResultError(fmt.Sprintf("Cannot access database file: %v", err)), nil
		}
	}

	if req.HomeDir != "" {
		if info, err := os.Stat(req.HomeDir); err != nil {
			if os.IsNotExist(err) {
				return mcp.NewToolResultError(fmt.Sprintf("Home directory does not exist: %s", req.HomeDir)), nil
			}
			return mcp.NewToolResultError(fmt.Sprintf("Cannot access home directory: %v", err)), nil
		} else if !info.IsDir() {
			return mcp.NewToolResultError(fmt.Sprintf("Home path is not a directory: %s", req.HomeDir)), nil
		}
	}

	// Configure the session
	if err := session.Configure(req.DBPath, req.HomeDir, req.ReadOnly); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Configuration failed: %v", err)), nil
	}

	// Build response
	response := ConfigureResponse{
		Success:  true,
		ReadOnly: req.ReadOnly,
	}

	// Set connection info
	if req.DBPath != "" {
		response.Connected = true
		response.DBPath = req.DBPath
		response.Message = fmt.Sprintf("Successfully connected to database: %s", req.DBPath)
		if req.ReadOnly {
			response.Message += " (read-only mode)"
		}
	}

	// Set home directory info
	if req.HomeDir != "" {
		response.HomeDir = req.HomeDir
		
		// List available files in home directory
		files, err := database.ListSupportedFiles(req.HomeDir)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list files in home directory: %v", err)), nil
		}
		response.AvailableFiles = files

		if req.DBPath == "" {
			// Only home directory was set
			response.Message = fmt.Sprintf("Home directory set to: %s", req.HomeDir)
			totalFiles := 0
			for _, fileList := range files {
				totalFiles += len(fileList)
			}
			response.Message += fmt.Sprintf(" (found %d supported files)", totalFiles)
		} else {
			// Both were set
			response.Message += fmt.Sprintf(" | Home directory: %s", req.HomeDir)
		}
	}

	// Convert response to JSON
	responseBytes, err := json.Marshal(response)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to format response: %v", err)), nil
	}

	return mcp.NewToolResultText(string(responseBytes)), nil
}

// GetConfigureToolDefinition returns the MCP tool definition for configure
func GetConfigureToolDefinition() mcp.Tool {
	return mcp.NewTool("configure",
		mcp.WithDescription("Configure DuckDB database connection and/or working directory"),
		mcp.WithString("db_path",
			mcp.Description("Path to database file or ':memory:' for in-memory database"),
		),
		mcp.WithString("home_dir",
			mcp.Description("Directory to scan for available database and data files"),
		),
		mcp.WithBoolean("read_only",
			mcp.Description("Connect in read-only mode (default: false)"),
		),
	)
}

// mapToStruct converts a map[string]interface{} to a struct
func mapToStruct(m map[string]interface{}, v interface{}) error {
	jsonBytes, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonBytes, v)
}
package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/marcboeker/go-duckdb/v2"
)

// Client wraps DuckDB connection and operations
type Client struct {
	db       *sql.DB
	dbPath   string
	readOnly bool
}

// NewClient creates a new DuckDB client
func NewClient(dbPath string, readOnly bool) (*Client, error) {
	// Validate the database path
	if err := validateDBPath(dbPath); err != nil {
		return nil, err
	}

	// Build connection string
	connStr := dbPath
	if readOnly && dbPath != ":memory:" {
		connStr += "?access_mode=read_only"
	}

	// Open database connection
	db, err := sql.Open("duckdb", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &Client{
		db:       db,
		dbPath:   dbPath,
		readOnly: readOnly,
	}, nil
}

// validateDBPath validates that the database path exists (if not :memory:)
func validateDBPath(dbPath string) error {
	if dbPath == ":memory:" {
		return nil
	}

	// Check if file exists
	if _, err := os.Stat(dbPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("database file does not exist: %s", dbPath)
		}
		return fmt.Errorf("cannot access database file: %w", err)
	}

	return nil
}

// Query executes a SQL query and returns formatted results
func (c *Client) Query(query string) (string, error) {
	if c.db == nil {
		return "", fmt.Errorf("database connection is not established")
	}

	rows, err := c.db.Query(query)
	if err != nil {
		return "", fmt.Errorf("query execution failed: %w", err)
	}
	defer rows.Close()

	return formatResults(rows)
}

// formatResults converts sql.Rows to a formatted table string
func formatResults(rows *sql.Rows) (string, error) {
	columns, err := rows.Columns()
	if err != nil {
		return "", fmt.Errorf("failed to get columns: %w", err)
	}

	// Get column types
	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return "", fmt.Errorf("failed to get column types: %w", err)
	}

	// Build header
	var result strings.Builder
	result.WriteString("┌")
	for i, col := range columns {
		if i > 0 {
			result.WriteString("┬")
		}
		width := max(len(col), len(columnTypes[i].DatabaseTypeName()))
		result.WriteString(strings.Repeat("─", width+2))
	}
	result.WriteString("┐\n")

	// Header row
	result.WriteString("│")
	for i, col := range columns {
		if i > 0 {
			result.WriteString("│")
		}
		colType := columnTypes[i].DatabaseTypeName()
		width := max(len(col), len(colType))
		result.WriteString(fmt.Sprintf(" %-*s ", width, col))
	}
	result.WriteString("│\n")

	// Type row
	result.WriteString("│")
	for i, colType := range columnTypes {
		if i > 0 {
			result.WriteString("│")
		}
		col := columns[i]
		width := max(len(col), len(colType.DatabaseTypeName()))
		result.WriteString(fmt.Sprintf(" %-*s ", width, colType.DatabaseTypeName()))
	}
	result.WriteString("│\n")

	// Separator
	result.WriteString("├")
	for i, col := range columns {
		if i > 0 {
			result.WriteString("┼")
		}
		colType := columnTypes[i].DatabaseTypeName()
		width := max(len(col), len(colType))
		result.WriteString(strings.Repeat("─", width+2))
	}
	result.WriteString("┤\n")

	// Data rows
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	rowCount := 0
	for rows.Next() {
		rowCount++
		if err := rows.Scan(valuePtrs...); err != nil {
			return "", fmt.Errorf("failed to scan row: %w", err)
		}

		result.WriteString("│")
		for i, val := range values {
			if i > 0 {
				result.WriteString("│")
			}

			col := columns[i]
			colType := columnTypes[i].DatabaseTypeName()
			width := max(len(col), len(colType))

			var str string
			if val == nil {
				str = "NULL"
			} else {
				str = fmt.Sprintf("%v", val)
			}

			result.WriteString(fmt.Sprintf(" %-*s ", width, str))
		}
		result.WriteString("│\n")
	}

	// Bottom border
	result.WriteString("└")
	for i, col := range columns {
		if i > 0 {
			result.WriteString("┴")
		}
		colType := columnTypes[i].DatabaseTypeName()
		width := max(len(col), len(colType))
		result.WriteString(strings.Repeat("─", width+2))
	}
	result.WriteString("┘\n")

	// Add row count
	result.WriteString(fmt.Sprintf("\n(%d row%s)\n", rowCount, func() string {
		if rowCount != 1 {
			return "s"
		}
		return ""
	}()))

	return result.String(), nil
}

// GetDBPath returns the database path
func (c *Client) GetDBPath() string {
	return c.dbPath
}

// IsReadOnly returns whether the client is in read-only mode
func (c *Client) IsReadOnly() bool {
	return c.readOnly
}

// Close closes the database connection
func (c *Client) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// ListSupportedFiles lists files in a directory that DuckDB can work with
func ListSupportedFiles(homeDir string) (map[string][]string, error) {
	files := make(map[string][]string)

	entries, err := os.ReadDir(homeDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		ext := strings.ToLower(filepath.Ext(name))

		switch ext {
		case ".db", ".duckdb":
			files["DuckDB Databases"] = append(files["DuckDB Databases"], name)
		case ".csv":
			files["CSV Files"] = append(files["CSV Files"], name)
		case ".parquet":
			files["Parquet Files"] = append(files["Parquet Files"], name)
		case ".json":
			files["JSON Files"] = append(files["JSON Files"], name)
		case ".xlsx":
			files["Excel Files"] = append(files["Excel Files"], name)
		}
	}

	return files, nil
}

package state

import (
	"sync"

	"github.com/rtihomir/mcp-tools/internal/duckdb/database"
)

// Session holds the state for a DuckDB MCP session
type Session struct {
	mu         sync.RWMutex
	dbClient   *database.Client
	dbPath     string
	homeDir    string
	readOnly   bool
	configured bool
}

// NewSession creates a new session instance
func NewSession() *Session {
	return &Session{}
}

// Configure sets up the database connection and/or home directory
func (s *Session) Configure(dbPath, homeDir string, readOnly bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Close existing connection if any
	if s.dbClient != nil {
		s.dbClient.Close()
		s.dbClient = nil
	}

	// If dbPath is provided, establish connection
	if dbPath != "" {
		client, err := database.NewClient(dbPath, readOnly)
		if err != nil {
			return err
		}
		s.dbClient = client
		s.dbPath = dbPath
	}

	// Set home directory
	if homeDir != "" {
		s.homeDir = homeDir
	}

	s.readOnly = readOnly
	s.configured = true

	return nil
}

// GetClient returns the current database client (thread-safe)
func (s *Session) GetClient() *database.Client {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.dbClient
}

// GetHomeDir returns the current home directory (thread-safe)
func (s *Session) GetHomeDir() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.homeDir
}

// GetDBPath returns the current database path (thread-safe)
func (s *Session) GetDBPath() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.dbPath
}

// IsReadOnly returns whether the session is in read-only mode
func (s *Session) IsReadOnly() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.readOnly
}

// IsConfigured returns whether the session has been configured
func (s *Session) IsConfigured() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.configured
}

// HasConnection returns whether there's an active database connection
func (s *Session) HasConnection() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.dbClient != nil
}

// Close closes any active database connection
func (s *Session) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.dbClient != nil {
		s.dbClient.Close()
		s.dbClient = nil
	}
}
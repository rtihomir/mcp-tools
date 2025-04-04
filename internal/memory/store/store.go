package store

import (
	"github.com/rtihomir/mcp-tools/internal/memory/config"
)

type Store struct {
	config *config.Config
}

func NewStore(config *config.Config) *Store {
	return &Store{
		config: config,
	}
}

func (s *Store) GetConfig() *config.Config {
	return s.config
}

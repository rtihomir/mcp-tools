package config

import "flag"

type Config struct {
	DbPath string
}

func NewConfig() *Config {
	config := &Config{}
	flag.StringVar(&config.DbPath, "db-path", "./memory.sqlite3", "Path to the database file. Default is ./memory.sqlite3")
	return config
}

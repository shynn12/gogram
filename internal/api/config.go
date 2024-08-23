package api

import "github.com/shynn2/cmd-gram/pkg/client/postgresql"

type Config struct {
	BindAddr string `toml:"bind_addr"`
	LogLevel string `toml:"log_level"`
	Store    *postgresql.Config
}

func NewConfig() *Config {
	return &Config{
		BindAddr: ":8080",
		LogLevel: "debug",
		Store:    postgresql.NewConfig(),
	}
}

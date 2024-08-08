package main

import (
	"log/slog"

	// "github.com/BurntSushi/toml"
)

type Config struct {
	Write map[string][]string
	WriteGlobal []string
	Read map[string][]string
	ReadGlobal []string
}

func InitConfig() *Config {
	slog.Error("IMPLEMENT CONFIGS")
	privs := make(map[string][]string)
	privs["test"] = []string{"token", "token2"}
	return &Config{Write:privs, WriteGlobal:[]string{"global"}, Read: privs, ReadGlobal: []string{"global"}}
}

func loadConfig() {

	slog.Info("nil")
}
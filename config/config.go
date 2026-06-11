package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	ServerPort  *int   `json:"server_port"`
	FilesDir    string `json:"files_dir"`
	DebugMode   bool   `json:"debug_mode"`
	BindAddress string `json:"bind_address"`
}

func LoadConfig(configFilePath string) (*Config, error) {

	if configFilePath == "" {
		configFilePath = "config.json"
	}

	data, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

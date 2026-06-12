package config

import (
	"encoding/json"
	"log"
	"os"
)

/*
A config releváns paramétereit tartalmazó struktúra, csak ezeket olvassa ki a JSON fájlból.
*/
type Config struct {
	ConfigFilePath string
	ServerPort     *int   `json:"server_port"`
	FilesDir       string `json:"files_dir"`
	DebugMode      bool   `json:"debug_mode"`
	BindAddress    string `json:"bind_address"`
}

/*
Populál egy Config struktúrát és visszaadja azt.
Ehhez beolvassa a JSON fájlt és a beépített JSON parser
segítségével értelmezi azt.
*/
func loadConfig(configFilePath string) (*Config, error) {

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

func SetupConfig(configFilePath string, cliPort int) Config {
	cfg, err := loadConfig(configFilePath)
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	finalPort := cliPort

	if finalPort == 0 {
		if cfg.ServerPort == nil {
			log.Fatal("Error: Configuration is missing port field and no CLI port provided.")
		}
		finalPort = *cfg.ServerPort
	} else {
		if cfg.ServerPort == nil {
			cfg.ServerPort = new(int)
		}
		*cfg.ServerPort = finalPort
	}

	cfg.ConfigFilePath = configFilePath

	return *cfg
}

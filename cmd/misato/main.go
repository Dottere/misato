package main

import (
	"flag"
	"fmt"
	"log"
	"misato/config"
	"misato/internal/server"
	"os"
)

const VERSION string = "1.0.0"

func main() {
	configFilePath, cliPort := parseFlags()

	cfg := setupConfig(configFilePath, cliPort)

	server.Configure(cfg)
	server.Start()
}

func parseFlags() (string, int) {
	var showVersion bool
	var configFilePath string
	var port int

	flag.BoolVar(&showVersion, "version", false, "Print the version number\n")
	flag.BoolVar(&showVersion, "v", false, "Print the version number (shorthand)\n")

	flag.StringVar(&configFilePath, "config", "./config.json", "Path to the config file")
	flag.StringVar(&configFilePath, "c", "./config.json", "Path to the config file (shorthand)")

	flag.IntVar(&port, "port", 0, "Server port")
	flag.IntVar(&port, "p", 0, "Server port (shorthand)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "misato - Self hosted manga site\n\nUsage:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if showVersion {
		fmt.Printf("misato version %s\n", VERSION)
		os.Exit(0)
	}

	return configFilePath, port
}

func setupConfig(configFilePath string, cliPort int) config.Config {
	cfg, err := config.LoadConfig(configFilePath)
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
		*cfg.ServerPort = finalPort
	}

	return *cfg
}

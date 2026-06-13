package main

import (
	"flag"
	"fmt"
	"misato/config"
	"misato/internal/server"
	"os"
)

const VERSION string = "0.4.0"

var ConfigPath string

func main() {
	configFilePath, cliPort := parseFlags()

	cfg := config.SetupConfig(configFilePath, cliPort)

	srv := server.NewAppServer(cfg)
	srv.RegisterRoute("/", srv.ServeMainPage)
	srv.RegisterRoute("/about", srv.ServeAboutPage)
	srv.RegisterRoute("/comics", srv.ServeBrowserPage)
	srv.RegisterRoute(fmt.Sprintf("/%s/{comicName}", cfg.FilesDir), srv.ServeFilesPage) // Go 1.22+ wildcard regisztrálás
	srv.RegisterRoute("/api/image", srv.ServeComic)

	srv.Start()

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
		fmt.Fprintf(os.Stderr, "MISATO - Self hosted manga site\n\nUsage:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if showVersion {
		fmt.Printf("MISATO version %s\n", VERSION)
		os.Exit(0)
	}

	return configFilePath, port
}

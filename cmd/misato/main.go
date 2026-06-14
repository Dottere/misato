package main

import (
	"fmt"
	"log"
	"misato/internal/config"
	"misato/internal/server"
	"misato/internal/utils"
)

const VERSION string = "0.5.0"

var ConfigPath string

func main() {
	configFilePath, cliPort, verboseMode := utils.ParseFlags(VERSION)

	cfg, err := config.SetupConfig(configFilePath, cliPort, verboseMode)
	if err != nil {
		log.Fatalf("Server couldn't read config: %v", err)
	}

	srv := server.NewAppServer(cfg)

	defer srv.Shutdown()

	srv.RegisterRoute("/", srv.ServeMainPage)
	srv.RegisterRoute("/about", srv.ServeAboutPage)
	srv.RegisterRoute("/comics", srv.ServeBrowserPage)
	srv.RegisterRoute(fmt.Sprintf("/%s/{comicName}", cfg.FilesDir), srv.ServeFilesPage) // Go 1.22+ wildcard regisztrálás
	srv.RegisterRoute("/api/image", srv.ServeComic)
	srv.RegisterRoute("/api/rescan", srv.Rescan)

	srv.Start()

}

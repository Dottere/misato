package server

import (
	"fmt"
	"log"
	"misato/config"
	"net/http"
)

var serverConfig config.Config

func Start() {

	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("web"))

	mux.Handle("/static/", fs)

	/*
		Vedd fel ide a végpontokat dinamikusan betöltődnek!

		A handlert a saját fájljában definiáld, pl mint a server/serveIndex.go fájlban!
	*/

	fmt.Printf("/%s/", serverConfig.FilesDir)
	endpoints := map[string]http.HandlerFunc{
		"/":          serveMainPage,
		"/about":     serveAboutPage,
		"/comics":    serveBrowserPage,
		"/mangas/":   serveFilesPage,
		"/api/image": serveComic,
	}

	for path, handler := range endpoints {

		mux.HandleFunc(path, handler)
	}

	fmt.Printf("Server open on port %d...\n", *serverConfig.ServerPort)

	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", serverConfig.BindAddress, *serverConfig.ServerPort), mux))
}

func Configure(configuration config.Config) {
	serverConfig = configuration
}

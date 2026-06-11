package server

import (
	"fmt"
	"log"
	"misato/config"
	"net/http"
)

var cfg config.Config

func Start() {

	initUptime()

	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("web"))

	mux.Handle("/static/", fs)

	/*
		Vedd fel ide a végpontokat dinamikusan betöltődnek!

		A handlert a saját fájljában definiáld, pl mint a server/serveIndex.go fájlban!
	*/

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

	fmt.Println("MISATO - Manga Site")

	fmt.Println("\nInitial scan...")
	scan()
	go Listen()

	fmt.Printf("\nBinding server to address %s on port %d...\n", cfg.BindAddress, *cfg.ServerPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", cfg.BindAddress, *cfg.ServerPort), mux))
}

func Configure(configuration config.Config) {
	cfg = configuration
}

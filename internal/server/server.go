package server

import (
	"fmt"
	"log"
	"net/http"
)

func Start(port int) {
	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("web"))

	mux.Handle("/static/", fs)

	/*
		Vedd fel ide a végpontokat dinamikusan betöltődnek!

		A handlert a saját fájljában definiáld, pl mint a server/serveIndex.go fájlban!
	*/
	endpoints := map[string]http.HandlerFunc{
		"/":      serveMainPage,
		"/about": serveAboutPage,
	}

	for path, handler := range endpoints {

		mux.HandleFunc(path, handler)
	}

	fmt.Printf("Server open on port %d...\n", port)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), mux))
}

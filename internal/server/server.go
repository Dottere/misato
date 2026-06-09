package server

import (
	"fmt"
	"log"
	"net/http"
)

func Start(port int) {

	/*
		Vedd fel ide a végpontokat dinamikusan betöltődnek!

		A handlert a saját fájljában definiáld, pl server/serveIndex.go!
	*/
	endpoints := map[string]http.HandlerFunc{
		"/": serveMainPage,
	}

	for path, handler := range endpoints {

		http.HandleFunc(path, handler)
	}

	fmt.Printf("Server open on port %d...", port)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

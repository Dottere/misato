package server

import (
	"fmt"
	"log"
	"misato/internal/cbz"
	"net/http"
)

func Start(port int) {

	mux := http.NewServeMux()
	/*
		Vedd fel ide a végpontokat dinamikusan betöltődnek!

		A handlert a saját fájljában definiáld, pl mint a server/serveIndex.go fájlban!
	*/
	endpoints := map[string]http.HandlerFunc{
		"/": serveMainPage,
	}

	for path, handler := range endpoints {

		mux.HandleFunc(path, handler)
	}

	fmt.Printf("Server open on port %d...\n", port)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), mux))
}

func OpenCBZ(name string) {
	Cbz, err := cbz.OpenCbz(name)

	if err != nil {
		fmt.Print(err)
		return
	}

	fmt.Printf("Cbz: %v\n", Cbz)
}

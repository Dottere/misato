package server

import (
	"net/http"
	"path/filepath"
)

func serveMainPage(w http.ResponseWriter, r *http.Request) {
	path := filepath.Join("web", "src", "index.html")

	handleFileServing(path, w, r)
}

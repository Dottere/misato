package server

import (
	"net/http"
	"path/filepath"
)

func serveMainPage(w http.ResponseWriter, r *http.Request) {

	http.ServeFile(w, r, filepath.Join("web", "src", "index.html"))
}

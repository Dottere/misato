package server

import (
	"net/http"
)

func serveMainPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "internal\\web\\src\\index.html")
}

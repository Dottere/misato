package server

import "net/http"

func serveMainPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	renderTemplate(w, r, "home.html", PageData{
		Title:      "Home — Misato",
		ActivePage: "home",
	})

}

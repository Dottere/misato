package server

import "net/http"

/*
Betölti a kezdőoldalt ("/" elérési úton)

# Átadott paraméterek:
  - Oldal címe: Home — Misato
*/
func ServeMainPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	renderTemplate(w, r, "home.html", PageData{
		Title:      "Home — Misato",
		ActivePage: "home",
	})

}

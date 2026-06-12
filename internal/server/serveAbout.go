package server

import "net/http"

/*
Megjeleníti a "rólunk" oldalt ahol esetlegesen leírjuk mire való ez a projekt, stb

# Átadott paraméterek:
  - Oldal címe: About — Misato
*/
func ServeAboutPage(w http.ResponseWriter, r *http.Request) {

	renderTemplate(w, r, "about.html", PageData{
		Title:      "About — Misato",
		ActivePage: "about",
	})

}

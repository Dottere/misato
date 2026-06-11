package server

import "net/http"

func serveAboutPage(w http.ResponseWriter, r *http.Request) {

	renderTemplate(w, r, "about.html", PageData{
		Title:      "About — Misato",
		ActivePage: "about",
	})

}

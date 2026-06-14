package server

import (
	"bytes"
	"html/template"
	"net/http"
	"path/filepath"
)

// PageData összefogja mindazokat a változókat és állapotokat, amelyeket a szerver
// átad a HTML sablonoknak a dinamikus weboldalak legenerálásához.
type PageData struct {
	Title      string
	ActivePage string
	Items      []ComicCard
	Images     []string
	FilesDir   string
	IsLoggedIn bool
	Username   string
}

// renderTemplate feldolgozza és a http.ResponseWriter-be írja a megadott HTML sablont.
// Teljesítményoptimalizálás céljából "Double-Checked Locking" módszerrel memóriában tartja
// a már lefordított (parsed) fájlokat. Debug módban a gyorsítótár (cache) inaktív,
// így a HTML fájlok módosításai azonnal, újraindítás nélkül látszódnak.
func (srv *AppServer) renderTemplate(w http.ResponseWriter, _ *http.Request, name string, data PageData) {

	base := filepath.Join("web", "templates", "base.html")
	page := filepath.Join("web", "templates", name)

	var tmpl *template.Template
	var err error

	if srv.cfg.DebugMode {
		tmpl, err = template.ParseFiles(base, page)
	} else {
		srv.coreMutex.RLock()
		cachedTmpl, exists := srv.templateCache[name]
		srv.coreMutex.RUnlock()

		if exists {
			tmpl = cachedTmpl
		} else {
			srv.coreMutex.Lock()

			cachedTmpl, exists := srv.templateCache[name]

			if exists {
				tmpl = cachedTmpl
			} else {
				tmpl, err = template.ParseFiles(base, page)
				if err == nil {
					srv.templateCache[name] = tmpl
				}
			}

			srv.coreMutex.Unlock()
		}
	}

	if err != nil {
		http.Error(w, "template error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "base", data); err != nil {
		http.Error(w, "render error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	buf.WriteTo(w)
}

// ==============================================================================
// Sima oldalak betöltése (Alapvető, statikusabb nézetek)
// ==============================================================================

// ServeMainPage a weboldal kezdőlapját (home) szolgálja ki.
// Mivel a Go ServeMux a "/" útvonalra minden nem regisztrált kérést is ráirányít
// (catch-all), ez a függvény szigorúan ellenőrzi, hogy pontosan a gyökérkönyvtárat
// kérték-e. Ha nem, egy 404-es oldalra irányít.
func (srv *AppServer) ServeMainPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		srv.ServeNotFound(w, r)
		return
	}

	srv.renderTemplate(w, r, "home.html", PageData{
		Title:      "Home — Misato",
		ActivePage: "home",
	})
}

// ServeAboutPage a statikus "Rólunk" (About) oldalt jeleníti meg,
// ahol a projekt célját és leírását olvashatják a látogatók.
func (srv *AppServer) ServeAboutPage(w http.ResponseWriter, r *http.Request) {
	srv.renderTemplate(w, r, "about.html", PageData{
		Title:      "About — Misato",
		ActivePage: "about",
	})
}

// ServeNotFound lecseréli a Go beépített, egyszerű 404-es válaszát egy stílusában
// az oldalhoz illeszkedő, egyedi "Not Found" HTML sablonra. Ezen felül gondoskodik
// a megfelelő HTTP 404-es státuszkód beállításáról a fejlécben.
func (srv *AppServer) ServeNotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)

	srv.renderTemplate(w, r, "404.html", PageData{
		Title:      "404 Not Found — Misato",
		ActivePage: r.URL.Path,
	})
}

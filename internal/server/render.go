package server

import (
	"bytes"
	"html/template"
	"net/http"
	"path/filepath"
)

/*
Azokat az adatokat tárolja amiket esetlegesen át kell adni a HTML templateknek a dinamikus
megjelenítés érdekében.
*/
type PageData struct {
	Title      string
	ActivePage string
	Items      []ComicCard
	Images     []string
	FilesDir   string
	IsLoggedIn bool
	Username   string
}

/*
Dinamikusan jeleníti meg az oldalakat a Go beépített html template-jeinek segítségével.
Ezen felül ha a debug mód nincs bekapcsolva akkor cacheli a templateket (módosításkor nem olvasódnak be újraindításig)
*/
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

// Sima oldalak betöltése (index és about mert ezek nem túl extrák) //

/*
Betölti a kezdőoldalt ("/" elérési úton)

# Átadott paraméterek:
  - Oldal címe: Home — Misato
*/
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

/*
Megjeleníti a "rólunk" oldalt ahol esetlegesen leírjuk mire való ez a projekt, stb

# Átadott paraméterek:
  - Oldal címe: About — Misato
*/
func (srv *AppServer) ServeAboutPage(w http.ResponseWriter, r *http.Request) {
	srv.renderTemplate(w, r, "about.html", PageData{
		Title:      "About — Misato",
		ActivePage: "about",
	})
}

// A sima 404 oldal helyett kiszolgál egy sajátot.
func (srv *AppServer) ServeNotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)

	srv.renderTemplate(w, r, "404.html", PageData{
		Title:      "404 Not Found — Misato",
		ActivePage: r.URL.Path,
	})
}

package server

import (
	"bytes"
	"html/template"
	"net/http"
	"path/filepath"
)

// Ez azért kell hogy a kiválasztott oldalt számon tartsuk (jelenleg csak navbar kiemeléshez)
type PageData struct {
	Title      string
	ActivePage string
	Items      []ComicCard
	Images     []string
	FilesDir   string
}

func renderTemplate(w http.ResponseWriter, r *http.Request, name string, data PageData) {

	base := filepath.Join("web", "templates", "base.html")
	page := filepath.Join("web", "templates", name)

	logToCLI(page, r)

	tmpl, err := template.ParseFiles(base, page)
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

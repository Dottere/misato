package server

import (
	"archive/zip"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
)

type ComicCard struct {
	Title    string
	CoverURL string
}

/*
Megjeleníti az olvasót a kérelmezett mangával feltöltve
*/
func (srv *AppServer) ServeComic(w http.ResponseWriter, r *http.Request) {
	comicName := r.URL.Query().Get("comic")
	indexStr := r.URL.Query().Get("index")

	cleanComicName := filepath.Base(comicName)
	if cleanComicName == "." || cleanComicName == "/" {
		http.Error(w, "Invalid comic name", http.StatusBadRequest)
		return
	}

	if indexStr == "" {
		http.Error(w, "Missing parameters", http.StatusBadRequest)
		return
	}

	zipIndex, err := strconv.Atoi(indexStr)
	if err != nil {
		http.Error(w, "Invalid index", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(srv.cfg.FilesDir, comicName+".cbz")
	zr, err := zip.OpenReader(filePath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer zr.Close()

	if zipIndex < 0 || zipIndex >= len(zr.File) {
		http.Error(w, "Index out of bounds", http.StatusNotFound)
		return
	}

	imageFile := zr.File[zipIndex]
	ext := strings.ToLower(filepath.Ext(imageFile.Name))

	rc, err := imageFile.Open()
	if err != nil {
		http.Error(w, "Failed to read image from zip", http.StatusInternalServerError)
		return
	}
	defer rc.Close()

	switch ext {
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	case ".webp":
		w.Header().Set("Content-Type", "image/webp")
	default:
		w.Header().Set("Content-Type", "image/jpeg")
	}

	io.Copy(w, rc)
}

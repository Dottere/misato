package server

import (
	"archive/zip"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
)

type ComicCard struct {
	Title    string
	CoverURL string
}

func serveComic(w http.ResponseWriter, r *http.Request) {
	comicName := r.URL.Query().Get("comic")
	indexStr := r.URL.Query().Get("index")

	if comicName == "" || indexStr == "" {
		http.Error(w, "Missing parameters", http.StatusBadRequest)
		return
	}

	zipIndex, err := strconv.Atoi(indexStr)
	if err != nil {
		http.Error(w, "Invalid index", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(serverConfig.FilesDir, comicName+".cbz")
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

	rc, err := imageFile.Open()
	if err != nil {
		http.Error(w, "Failed to read image from zip", http.StatusInternalServerError)
		return
	}
	defer rc.Close()

	w.Header().Set("Content-Type", "image/jpeg")
	io.Copy(w, rc)
}

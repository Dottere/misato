package server

import (
	"fmt"
	"misato/internal/cbz"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
)

func serveFilesPage(w http.ResponseWriter, r *http.Request) {
	comicNameURL := strings.TrimPrefix(r.URL.Path, "/mangas/")

	comicName, err := url.PathUnescape(comicNameURL)
	if err != nil {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	if comicName == "" || strings.Contains(comicName, "..") {
		http.NotFound(w, r)
		return
	}

	fileName := comicName + ".cbz"
	filePath := filepath.Join(cfg.FilesDir, fileName)

	cbzFile, err := cbz.OpenCbz(filePath)

	if err != nil {
		fmt.Println("Couldn't open file: ", cbzFile.UrlPath)
		return
	}
	defer cbzFile.Handle.Close()

	var imageUrls []string
	for _, zipIndex := range cbzFile.FileIndicesToPages {
		imgUrl := fmt.Sprintf("/api/image?comic=%s&index=%d", url.QueryEscape(comicName), zipIndex)
		imageUrls = append(imageUrls, imgUrl)
	}

	data := PageData{
		Title:  comicName,
		Images: imageUrls,
	}

	renderTemplate(w, r, "reader.html", data)
}

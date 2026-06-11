package server

import (
	"fmt"
	"log"
	"misato/internal/cbz"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	storedItems []ComicCard
	itemsMutex  sync.RWMutex // The lock for our cache
)

func serveBrowserPage(w http.ResponseWriter, r *http.Request) {

	itemsMutex.RLock()

	itemsToRender := storedItems

	itemsMutex.RUnlock()

	renderTemplate(w, r, "comics.html", PageData{
		Title:      "Comics — Misato",
		ActivePage: "comics",
		Items:      itemsToRender,
		FilesDir:   cfg.FilesDir,
	})

}

// Átnézi a mappát ahol a mangák tárolva vannak és indexeli őket
func getAllStoredComics() []ComicCard {
	folderPath := cfg.FilesDir

	entries, err := os.ReadDir(folderPath)
	if err != nil {
		log.Println(err)
		return nil
	}

	var storedComics []ComicCard

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".cbz" {
			continue
		}

		comicName := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		filePath := filepath.Join(folderPath, entry.Name())

		fmt.Printf("\nScanning \"%s\"", comicName)

		cbzFile, err := cbz.OpenCbz(filePath)
		if err != nil {
			continue
		}

		coverZipIndex := cbzFile.FileIndicesToPages[0]

		cbzFile.Handle.Close()

		coverUrl := fmt.Sprintf("/api/image?comic=%s&index=%d", url.QueryEscape(comicName), coverZipIndex)

		storedComics = append(storedComics, ComicCard{
			Title:    comicName,
			CoverURL: coverUrl,
		})
	}
	fmt.Printf("\n\nScanning finished\n")
	fmt.Printf("Scanned %d comics\n", len(storedComics))
	return storedComics
}

func scan() {
	newItems := getAllStoredComics()

	itemsMutex.Lock()

	storedItems = newItems

	itemsMutex.Unlock()
}

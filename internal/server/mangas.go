package server

import (
	"fmt"
	"io"
	"log"
	"misato/internal/cbz"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

/*
A könnyed indexeléshez használt struktúra ami eltárolja a legelső oldalt (ami alapvetően a nyitóoldal szokott lenni)
és a manga címét (a zip fájl neve)
*/
type ComicCard struct {
	Title    string
	CoverURL string
}

/*
A ComicCards típus önmagában nincs használva, csak tömbként, így ezzel aliasoljuk majd ehhez
kötünk egy segédfüggvényt ami segít eldönteni, hogy egy manga indexelve van-e
*/
type ComicCards []ComicCard

/*
A segédfüggvény amit a ComicCards típushoz kötünk.

  - Igazat ad, hogyha bármelyik elem címe megegyezik a paraméterrel
  - Hamist ad, hogyha nem talál egyezést
*/
func (c ComicCards) Contains(e string) bool {
	for _, elem := range c {
		if elem.Title == e {
			return true
		}
	}
	return false
}

// Átnézi a mappát ahol a mangák tárolva vannak és indexeli őket
func (srv *AppServer) getAllStoredComics() []ComicCard {
	folderPath := srv.cfg.FilesDir

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

		if srv.cfg.DebugMode || srv.cfg.VerboseMode {
			fmt.Printf("\nScanning \"%s\"", comicName)
		}
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
	if srv.cfg.DebugMode || srv.cfg.VerboseMode {
		fmt.Printf("\n\nScanning finished\n")
		fmt.Printf("Scanned %d comics\n", len(storedComics))
	}
	return storedComics
}

/*
Meghívja a getAllStoredComics függvényt majd szálbiztos módon felülírja a szerver storedItems példányát
*/
func (srv *AppServer) scan() {
	newItems := srv.getAllStoredComics()

	srv.cacheMutex.Lock()

	srv.storedItems = newItems

	srv.cacheMutex.Unlock()
}

/*
Kiszolgálja a mangakérelmeket, amik a comics.html oldalról "/mangas/manga_név" formátumú kérésekként érkeznek

# Átadott paraméterek:
  - Oldal címe: A manga neve
*/
func (srv *AppServer) ServeFilesPage(w http.ResponseWriter, r *http.Request) {

	rawComicName := r.PathValue("comicName")

	comicName, err := url.PathUnescape(rawComicName)
	if err != nil || comicName == "" {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	if !srv.storedItems.Contains(comicName) {
		srv.ServeNotFound(w, r)
		return
	}

	cleanComicName := filepath.Base(comicName)
	if cleanComicName == "." || cleanComicName == "/" {
		http.NotFound(w, r)
		return
	}

	fileName := cleanComicName + ".cbz"
	filePath := filepath.Join(srv.cfg.FilesDir, fileName)

	cbzFile, err := cbz.OpenCbz(filePath)
	if err != nil {
		fmt.Println("Couldn't open file: ", filePath)
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

	srv.renderTemplate(w, r, "reader.html", data)
}

/*
Végigiterál az eltárolt mangákon és dinamikusan megjeleníti azokat az oldalon

# Átadott paraméterek:
  - Oldal címe: Comics — Misato
  - Az eltárolt mangák: itemsToRender
  - A mappa ahol a mangák tárolva vannak: srv.cfg.Filesdir (dinamikusan kiolvasott)
*/
func (srv *AppServer) ServeBrowserPage(w http.ResponseWriter, r *http.Request) {

	srv.cacheMutex.RLock()

	itemsToRender := srv.storedItems

	srv.cacheMutex.RUnlock()

	srv.renderTemplate(w, r, "comics.html", PageData{
		Title:      "Comics — Misato",
		ActivePage: "comics",
		Items:      itemsToRender,
		FilesDir:   srv.cfg.FilesDir,
	})

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

	filePath := filepath.Join(srv.cfg.FilesDir, cleanComicName+".cbz")
	zr, err := srv.getArchive(filePath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

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

func (srv *AppServer) Rescan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	srv.scan()

	w.WriteHeader(http.StatusOK)
}

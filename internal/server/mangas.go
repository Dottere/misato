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

// ComicCard egyetlen képregény (CBZ archívum) alapvető metaadatait tárolja a webes felület számára.
// Tartalmazza a címet és a borítóképhez (általában a legelső oldalhoz) vezető API végpont URL-jét.
type ComicCard struct {
	Title    string
	CoverURL string
}

// ComicCards a ComicCard elemek egyedi típus-aliasza.
// Segítségével egyedi metódusokat rendelhetünk a képregények listájához.
type ComicCards []ComicCard

// Contains leellenőrzi, hogy a megadott képregénycím (title) szerepel-e a memóriában tartott listában.
// Gyors validációra szolgál az URL-ből érkező útvonalak ellenőrzésekor.
func (c ComicCards) Contains(e string) bool {
	for _, elem := range c {
		if elem.Title == e {
			return true
		}
	}
	return false
}

// getAllStoredComics végigiterál a konfigurációban megadott könyvtáron, és feldolgozza a .cbz kiterjesztésű fájlokat.
// Kinyeri a címeket és legenerálja a borítóképek API végpontjait. Ez a folyamat I/O intenzív,
// ezért közvetlenül ritkán, főleg induláskor vagy manuális frissítéskor hívódik meg.
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

// scan egy szálbiztos (thread-safe) burkolófüggvény a getAllStoredComics köré.
// Gondoskodik róla, hogy az új képregények beolvasása közben a webes kiszolgálás zavartalan maradjon,
// majd a kész listát lecseréli a memóriában.
func (srv *AppServer) scan() {
	newItems := srv.getAllStoredComics()

	srv.cacheMutex.Lock()

	srv.storedItems = newItems

	srv.cacheMutex.Unlock()
}

// ServeFilesPage legenerálja és kiszolgálja az olvasó (reader) felületet egy konkrét képregényhez.
// Beolvassa az adott CBZ fájl belső szerkezetét, és átadja a sablonnak az összes oldalhoz tartozó
// API hivatkozást, hogy a böngésző dinamikusan betölthesse azokat.
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

// ServeBrowserPage kiszolgálja a fő könyvtárnézetet (catalog).
// Szálbiztos módon kiolvassa a memóriában tartott képregénylistát, és átadja azt
// a "comics.html" sablonnak a rácsos megjelenítéshez.
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

// ServeComic egy dedikált API végpont, amely egy konkrét képet (oldalt) szolgál ki a tömörített CBZ archívumból.
// A teljesítmény maximalizálása érdekében a memóriában tartott zipCache-t használja,
// így elkerüli a fájlok másodpercenkénti többszöri megnyitását. Támogatja a PNG, WEBP és JPEG formátumokat.
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

// Rescan egy POST végpont, amivel a kliens (vagy egy adminisztrátor) manuálisan kikényszerítheti
// a fájlrendszer újraolvasását, anélkül, hogy a szervert újra kellene indítani.
func (srv *AppServer) Rescan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	srv.scan()

	w.WriteHeader(http.StatusOK)
}

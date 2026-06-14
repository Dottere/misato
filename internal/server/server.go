package server

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"html/template"
	"log"
	"misato/internal/config"
	"misato/internal/database"
	"net/http"
	"sync"
	"time"
)

var ErrRouteExists = errors.New("route already exists!")

type CachedArchive struct {
	Reader     *zip.ReadCloser
	LastAccess time.Time
}

// Enkapszulálja az alap http.Server struktúrát egy
// saját implementációba ami tárolja a projektspecifikus
// változókat pl.: config, mutexek, saját ServeMux, végpontok, indexelt mangák, cachelt templatek
type AppServer struct {
	// A közös állapotokat írni és olvasni kell, ezért kell a mutex
	//
	// # Használva van:
	//
	// 	- Új utak regisztrálásakor
	// 	- Regisztrált utak olvasásakor
	//  - Leállításkor
	coreMutex sync.RWMutex

	cfg config.Config // A config fájl tartalma
	DB  *database.DB  // Az adatbázis kapcsolat

	// A Listen() ami elindítja a konzolt az egy goroutine,
	// ezt csak egyszer akarjuk elindítani
	listenOnce sync.Once
	// Ugyanaz az ötlet mint a Listen()-nel
	// csak ebben az esetben a Shutdown()-ra értelmezve
	shutdownOnce sync.Once

	srv *http.Server // Az enkapszulált szerver

	mux *http.ServeMux // Privát router

	// Egy map ami tárolja hogy melyik elérési utat milyen függvénnyel
	// kell kiszolgálni
	endpoints map[string]http.HandlerFunc

	// A mangákat tároló könyvtár olvasásához/írásához külön mutex
	cacheMutex sync.RWMutex

	// Az eltárolt mangák itt vannak cachelve és indexelve
	storedItems ComicCards

	// http template cache, hogy ne kelljen mindig újraolvasni
	templateCache map[string]*template.Template

	startTime time.Time

	zipMutex       sync.Mutex
	zipCache       map[string]*CachedArchive
	zipCleanupOnce sync.Once
}

// Az *AppServer struktúra publikus konstruktora, alapvetően
// csak ezzel lehet (kell) egy újat létrehozni belőle.
//
// A bekért cfg-t a config.Config-ban található LoadConfig(path string)
// függvénnyel lehet lekérni.
func NewAppServer(cfg config.Config) *AppServer {

	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("web/static"))

	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.BindAddress, cfg.ServerPort),
		Handler:      LoggingMiddleware(mux),
		ReadTimeout:  time.Duration(cfg.ReadTimeout),
		WriteTimeout: time.Duration(cfg.WriteTimeout),
		IdleTimeout:  time.Duration(cfg.IdleTimeout),
	}

	dbInstance, err := database.InitDB(cfg.DBPath)
	if err != nil {
		log.Fatalf("Could not initialize database: %v", err)
	} else {
		LogDebug("Successfully connected to SQLite database.")
	}

	return &AppServer{
		cfg:           cfg,
		mux:           mux,
		srv:           server,
		endpoints:     make(map[string]http.HandlerFunc),
		templateCache: make(map[string]*template.Template),
		zipCache:      make(map[string]*CachedArchive),
		DB:            dbInstance,
	}
}

// Regisztrál egy utat a szerverre, bekéri a mintát ami lehet pl "/"
// és a http.HandlerFunc formátumú függvényt ami kiszolgálja azt.
//
// Példának lásd a serveIndex.go fájlt.
func (srv *AppServer) RegisterRoute(route string, handler http.HandlerFunc) (err error) {

	srv.coreMutex.Lock()
	defer srv.coreMutex.Unlock()

	if _, exists := srv.endpoints[route]; exists {
		return ErrRouteExists
	}

	srv.endpoints[route] = handler
	srv.mux.Handle(route, handler)

	return nil
}

type Route struct {
	Pattern string
}

// Lekéri a regisztrált utakat és visszaadja azt egy olyan tömbben
// ami jelenleg csak az utak mintáját (pl "/") tárolja, viszont ez
// változhat a jövőben ha szükség van másra.
func (srv *AppServer) GetRoutes() []Route {
	out := make([]Route, 0, len(srv.endpoints))

	srv.coreMutex.RLock()

	for r := range srv.endpoints {
		out = append(out, Route{Pattern: r})
	}

	srv.coreMutex.RUnlock()

	return out
}

// Visszaadja a szerverben eltárolt konfigurációt
func (srv *AppServer) GetConfig() config.Config {
	out := config.Config{
		ConfigFilePath: srv.cfg.ConfigFilePath,
		ServerPort:     srv.cfg.ServerPort,
		FilesDir:       srv.cfg.FilesDir,
		DebugMode:      srv.cfg.DebugMode,
		BindAddress:    srv.cfg.BindAddress,
	}

	return out
}

func (srv *AppServer) getArchive(filePath string) (*zip.ReadCloser, error) {
	srv.zipMutex.Lock()
	defer srv.zipMutex.Unlock()

	if archive, exists := srv.zipCache[filePath]; exists {
		archive.LastAccess = time.Now()
		return archive.Reader, nil
	}

	zr, err := zip.OpenReader(filePath)
	if err != nil {
		return nil, err
	}

	srv.zipCache[filePath] = &CachedArchive{
		Reader:     zr,
		LastAccess: time.Now(),
	}

	return zr, nil
}

func (srv *AppServer) cleanupZipCache() {
	for {
		time.Sleep(2 * time.Minute)

		srv.zipMutex.Lock()
		for path, archive := range srv.zipCache {
			if time.Since(archive.LastAccess) > 5*time.Minute {
				archive.Reader.Close()
				delete(srv.zipCache, path)
				LogDebug("Closed inactive archive: " + path)
			}
		}
		srv.zipMutex.Unlock()
	}
}

// Elindítja a szervert, inicializálja a futamidőszámlálót,
// kiírja az üdvözlő üzenetet a konzolra és egyéb információkat.
//
// Ezen felül elindítja a parancssort is amin keresztül lehet interaktálni a szerverrel.
func (srv *AppServer) Start() {

	initUptime(&srv.startTime)
	srv.SetupDebug()

	fmt.Println("MISATO - Manga Site")
	if srv.cfg.DebugMode || srv.cfg.VerboseMode {
		fmt.Printf("\nBinding server to address %s on port %d...\n", srv.cfg.BindAddress, srv.cfg.ServerPort)

		fmt.Println("\nInitial scan...")
	}
	srv.scan()

	go func() {
		err := srv.srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	srv.zipCleanupOnce.Do(func() {
		go srv.cleanupZipCache()
	})

	srv.listenOnce.Do(func() {
		if srv.cfg.DebugMode || srv.cfg.VerboseMode {
			fmt.Println("\nSetting up console listener...")
		}
		Listen(srv)
	})
}

// Leállítja a szervert, fontos, nem csak elvágja a kapcsolatot,
// hanem megvárja míg minden folyamatban lévő tevékenység befejeződik
func (srv *AppServer) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.srv.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

}

// Lemásolja a régi konfigurációt, majd eldobja a régi szervert.
// Az új szerverben mindent újonnan inicializál kivéve a régi muxot, azt megtartja
// a regisztrált elérési utak megtartása érdekében. Ha ezzel végzett, akkor elindítja az új szervert
// és új uptime időzítőt is indít.
func (srv *AppServer) Restart() {
	fmt.Println("Initiating server restart...")

	configFilePath := srv.cfg.ConfigFilePath

	srv.Stop()
	fmt.Println("Old server instance stopped.")

	newCfg, err := config.SetupConfig(configFilePath, 0, false)
	if err != nil {
		log.Fatalf("Server couldn't read config after restart: %v", err)
	}

	srv.coreMutex.Lock()
	srv.cfg = newCfg

	srv.srv = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", srv.cfg.BindAddress, srv.cfg.ServerPort),
		Handler:      LoggingMiddleware(srv.mux), // A régi mux-ot (és a regisztrált utakat) megtartjuk!
		ReadTimeout:  time.Duration(srv.cfg.ReadTimeout),
		WriteTimeout: time.Duration(srv.cfg.WriteTimeout),
		IdleTimeout:  time.Duration(srv.cfg.IdleTimeout),
	}
	srv.coreMutex.Unlock()

	fmt.Printf("Restarting HTTP server on %s:%d...\n", srv.cfg.BindAddress, srv.cfg.ServerPort)

	initUptime(&srv.startTime)

	go func() {
		err := srv.srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("HTTP server error after restart: %v", err)
		}
	}()

	fmt.Println("Restart complete!")

}

func (srv *AppServer) Shutdown() {
	srv.shutdownOnce.Do(func() {
		srv.Stop()
		srv.DB.Conn.Close()

		for _, archive := range srv.zipCache {
			archive.Reader.Close()
		}
	})
}

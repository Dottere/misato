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

// CachedArchive egy memóriában tartott ZIP fájlolvasót és az utolsó hozzáférés idejét tárolja.
// Célja, hogy a folyamatos I/O műveletek minimalizálásával gyorsítsa a képregények olvasását.
type CachedArchive struct {
	Reader     *zip.ReadCloser
	LastAccess time.Time
}

// AppServer a Misato alkalmazás központi szerver struktúrája.
// Összefogja a HTTP szervert, a routert (ServeMux), az adatbázis kapcsolatot,
// a konfigurációt és a szálbiztos gyorsítótárakat (sablonok, fájlok).
type AppServer struct {
	// coreMutex védi a regisztrált végpontok (endpoints) és a konfiguráció olvasását/írását,
	// különösen az alkalmazás hot-restart folyamata alatt.
	coreMutex sync.RWMutex

	cfg config.Config // Az alkalmazás betöltött konfigurációja
	DB  *database.DB  // Az SQLite adatbázis aktív kapcsolata

	// Szinkronizációs primitívek, amik garantálják, hogy az adott logikák (konzol, leállás, takarítás)
	// az alkalmazás életciklusa alatt szigorúan csak egyszer fussanak le.
	listenOnce     sync.Once
	shutdownOnce   sync.Once
	zipCleanupOnce sync.Once

	srv *http.Server   // A beépített Go HTTP szerver példánya
	mux *http.ServeMux // Az egyedi router, ami az útvonalakat (route) kezeli

	// endpoints tárolja a regisztrált végpontokat a duplikációk elkerülése végett.
	endpoints map[string]http.HandlerFunc

	// cacheMutex védi a lemezről beolvasott és indexelt mangák listáját (storedItems).
	cacheMutex  sync.RWMutex
	storedItems ComicCards

	// templateCache a lefordított HTML sablonokat tárolja, elkerülve a folyamatos lemezolvasást.
	templateCache map[string]*template.Template

	// startTime a szerver legutóbbi indulásának (vagy újraindulásának) pontos ideje az uptime számításhoz.
	startTime time.Time

	// zipMutex védi a megnyitott képregények memóriatárát (zipCache).
	zipMutex sync.Mutex
	zipCache map[string]*CachedArchive
}

// NewAppServer inicializálja és alapértékekkel tölti fel az AppServer egy új példányát.
// Felépíti az adatbázis kapcsolatot, beállítja a statikus fájlok kiszolgálását,
// és betölti az alapvető időtúllépési (timeout) szabályokat.
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

// RegisterRoute biztonságosan rögzít egy új végpontot a szerveren.
// Ha a megadott útvonal minta (pl. "/api/rescan") már létezik, ErrRouteExists hibát dob.
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

// Route egy regisztrált végpont mintáját reprezentálja.
type Route struct {
	Pattern string
}

// GetRoutes visszaadja az alkalmazásban jelenleg regisztrált összes elérési utat.
func (srv *AppServer) GetRoutes() []Route {
	out := make([]Route, 0, len(srv.endpoints))

	srv.coreMutex.RLock()

	for r := range srv.endpoints {
		out = append(out, Route{Pattern: r})
	}

	srv.coreMutex.RUnlock()

	return out
}

// GetConfig biztonságosan (másolatként) adja vissza a szerver aktuális beállításait.
func (srv *AppServer) GetConfig() config.Config {
	srv.coreMutex.RLock()
	defer srv.coreMutex.RUnlock()

	return srv.cfg
}

// getArchive visszaad egy nyitott ZIP olvasót a kért fájlhoz.
// Ha a fájl már nyitva van a gyorsítótárban, frissíti az utolsó hozzáférés idejét,
// ha pedig nincs, megnyitja és regisztrálja a memóriában.
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

// cleanupZipCache egy háttérfolyamat, ami rendszeres időközönként felszabadítja az erőforrásokat.
// Lezárja és törli a memóriából azokat a ZIP fájlokat, amiket egy meghatározott ideje (5 perc) nem olvastak.
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

// Start elindítja a HTTP szervert, végrehajtja a kezdő könyvtárszkennelést,
// inicializálja az uptime számlálót, és beindítja a háttérfolyamatokat (pl. zip takarító, parancssor).
func (srv *AppServer) Start() {

	initUptime(&srv.startTime)
	srv.SetupDebug()

	fmt.Println("MISATO - Manga Site")
	if srv.cfg.DebugMode || srv.cfg.VerboseMode {
		log.Printf("Binding server to address %s on port %d...\n", srv.cfg.BindAddress, srv.cfg.ServerPort)

		log.Println("Initial scan...")
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

// Stop elindítja a HTTP szerver finom (graceful) leállítását.
// Nem szakítja meg azonnal a futó kéréseket, de maximum 5 másodpercet ad a befejezésükre.
func (srv *AppServer) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.srv.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

}

// Restart hot-restartolja az alkalmazást: leállítja az aktív http modult,
// újraolvassa a friss konfigurációt a lemezről, majd elindít egy új http szerver példányt.
// Fontos: a korábban regisztrált útvonalak (mux) és az aktív gyorsítótárak érintetlenek maradnak.
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

// Shutdown véglegesen leállítja az alkalmazást: bontja a hálózati kapcsolatokat,
// lezárja az adatbázist, és felszabadítja az összes nyitva maradt fájlleírót (file descriptor) a gyorsítótárból.
func (srv *AppServer) Shutdown() {
	srv.shutdownOnce.Do(func() {
		srv.Stop()
		srv.DB.Conn.Close()

		srv.zipMutex.Lock()
		defer srv.zipMutex.Unlock()
		for _, archive := range srv.zipCache {
			archive.Reader.Close()
		}
	})
}

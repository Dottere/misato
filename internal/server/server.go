package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"misato/config"
	"net/http"
	"sync"
	"time"
)

var ErrRouteExists = errors.New("route already exists!")

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

	// A Listen() ami elindítja a konzolt az egy goroutine,
	// ezt csak egyszer akarjuk elindítani
	listenOnce sync.Once

	srv *http.Server // Az enkapszulált szerver

	mux *http.ServeMux // Privát router

	// Egy map ami tárolja hogy melyik elérési utat milyen függvénnyel
	// kell kiszolgálni
	endpoints map[string]http.HandlerFunc

	// A mangákat tároló könyvtár olvasásához/írásához külön mutex
	cacheMutex sync.RWMutex

	// Az eltárolt mangák itt vannak cachelve és indexelve
	storedItems []ComicCard
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
		Addr:         fmt.Sprintf("%s:%d", cfg.BindAddress, *cfg.ServerPort),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return &AppServer{
		cfg:       cfg,
		mux:       mux,
		srv:       server,
		endpoints: make(map[string]http.HandlerFunc),
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
		ServerPort:     copyPortPtr(srv.cfg.ServerPort),
		FilesDir:       srv.cfg.FilesDir,
		DebugMode:      srv.cfg.DebugMode,
		BindAddress:    srv.cfg.BindAddress,
	}

	return out
}

// Elindítja a szervert, inicializálja a futamidőszámlálót,
// kiírja az üdvözlő üzenetet a konzolra és egyéb információkat.
//
// Ezen felül elindítja a parancssort is amin keresztül lehet interaktálni a szerverrel.
func (srv *AppServer) Start() {

	initUptime()

	fmt.Println("MISATO - Manga Site")
	fmt.Printf("\nBinding server to address %s on port %d...\n", srv.cfg.BindAddress, *srv.cfg.ServerPort)

	fmt.Println("\nInitial scan...")
	srv.scan()

	srv.listenOnce.Do(func() {
		fmt.Println("\nSetting up console listener...")
		go Listen(srv)
	})

	err := srv.srv.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Printf("HTTP server error: %v", err)
	}
}

// Leállítja a szervert, fontos, nem csak elvágja a kapcsolatot,
// hanem megvárja míg minden folyamatban lévő tevékenység befejeződik
func (srv *AppServer) Stop() {
	srv.coreMutex.RLock()
	serverInstance := srv.srv
	srv.coreMutex.RUnlock()

	if serverInstance != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := serverInstance.Shutdown(ctx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}
}

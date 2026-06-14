package server

import (
	"fmt"
	"log"
	"net"
	"net/http"
)

var debugMode bool

// statusRecorder egy egyedi http.ResponseWriter burkoló (wrapper).
// Célja, hogy elfogja és eltárolja a kliensnek visszaküldött HTTP státuszkódot
// a naplózás számára, mivel az eredeti ResponseWriterből ezt nem lehet közvetlenül kiolvasni.
type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader felülírja a http.ResponseWriter alapértelmezett metódusát.
// Eltárolja a kiküldött státuszkódot (pl. 200, 404) a struktúrában,
// mielőtt továbbadná azt a tényleges kliensnek.
func (rec *statusRecorder) WriteHeader(code int) {
	rec.statusCode = code
	rec.ResponseWriter.WriteHeader(code)
}

// LoggingMiddleware egy HTTP köztesréteg (middleware), amely minden beérkező kérést
// átenged a megadott http.Handler-en, miközben méri és naplózza a válasz kimenetelét.
// A státuszkód alapértelmezett értéke 200 (OK), ha a kiszolgáló explicit nem állít be mást.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		recorder := &statusRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(recorder, r)

		logRequestToCLI(r, recorder.statusCode)
	})
}

// logRequestToCLI a megadott formátumban a standard kimenetre (stdout) írja a befejezett kérés metaadatait.
// Példa kimenet: [200] OK | GET - /api/image (127.0.0.1)
func logRequestToCLI(r *http.Request, statusCode int) {

	request_ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		request_ip = r.RemoteAddr
	}

	switch statusCode {
	case http.StatusOK:
		fmt.Printf("[200] OK | %s - %s (%s)\n", r.Method, r.URL.Path, request_ip)
	case http.StatusNotModified:
		fmt.Printf("[304] From cache | %s - %s (%s)\n", r.Method, r.URL.Path, request_ip)
	case http.StatusNotFound:
		fmt.Printf("[404] Not Found | %s - %s (%s)\n", r.Method, r.URL.Path, request_ip)
	case http.StatusBadRequest:
		fmt.Printf("[400] Bad Request | %s - %s (%s)\n", r.Method, r.URL.Path, request_ip)
	case http.StatusInternalServerError:
		fmt.Printf("[500] Internal Error | %s - %s (%s)\n", r.Method, r.URL.Path, request_ip)
	default:
		fmt.Printf("[%d] | %s - %s (%s)\n", statusCode, r.Method, r.URL.Path, request_ip)
	}
}

// SetupDebug beállítja a csomagszintű (package-level) debugMode változót
// az AppServer konfigurációja alapján.
func (srv *AppServer) SetupDebug() {
	debugMode = srv.cfg.DebugMode
}

// LogDebug egy globális segédfüggvény, amely biztonságosan (feltételhez kötve)
// naplóz a standard log kimenetre, de csak akkor, ha a fejlesztői (debug) mód aktív.
func LogDebug(message string) {
	if !debugMode {
		return
	}

	log.Println(message)
}

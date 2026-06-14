package server

import (
	"fmt"
	"log"
	"net"
	"net/http"
)

var debugMode bool

/*
A státuszkódok dinamikus tárolására használt struktúra ami valójában egy http.ResponseWriter
*/
type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

/*
A http.ResponseWriter.WriteHeader megvalósítása a saját típusunkra ami eltárolja a státuszkódot is ezzel
dinamikussá és sokkal könnyebbé téve a logolást
*/
func (rec *statusRecorder) WriteHeader(code int) {
	rec.statusCode = code
	rec.ResponseWriter.WriteHeader(code)
}

/*
Ez egy olyan köztes állapot a mux számára ami feljegyzi a státuszkódot és kilogolja azt
a konzolra
*/
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

/*
Kiírja stdout-ra, hogy a beérkezett kérésre milyen válasz lett kiküldve, miféle kérés volt az, és hogy mi lett kérve, illetve
honnan lett kérve

Például:

	[200] OK | GET - mangas/testComic (127.0.0.1)
*/
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

func (srv *AppServer) SetupDebug() {
	debugMode = srv.cfg.DebugMode
}

func LogDebug(message string) {
	if !debugMode {
		return
	}

	log.Println(message)
}

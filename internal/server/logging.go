package server

import "net/http"

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

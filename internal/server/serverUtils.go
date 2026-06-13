package server

import (
	"fmt"
	"net"
	"net/http"
)

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

/*
A port pointerként van eltárolva, így csekkoljuk le, hogy létezik-e. Ez a függvény annak a másolását segíti elő,
hogy mindenképpen értékként legyen átadva és ne mutatóként.
*/
func copyPortPtr(v *int) *int {
	if v == nil {
		return nil
	}
	x := *v
	return &x
}

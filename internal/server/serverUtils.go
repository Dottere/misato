package server

import (
	"fmt"
	"net"
	"net/http"
	"os"
)

/*
Megmondja, hogy létezik-e egy elérési út.
*/
func isFilePathValid(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {

		return false
	}
	return true
}

/*
Kiírja stdout-ra, hogy a beérkezett kérésre milyen válasz lett kiküldve, miféle kérés volt az, és hogy mi lett kérve, illetve
honnan lett kérve

Például:

	[200] OK | GET - mangas/testComic (127.0.0.1)
*/
func logRequestToCLI(path string, r *http.Request) {

	request_ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		request_ip = r.RemoteAddr
	}

	if isFilePathValid(path) {

		fmt.Printf("\n[200] OK | %s - %s (%s)\n>", r.Method, path, request_ip)
	} else {
		fmt.Printf("\n[404] File not found | %s - %s (%s)\n>", r.Method, r.URL.Path, request_ip)
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

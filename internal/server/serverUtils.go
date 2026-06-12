package server

import (
	"fmt"
	"net/http"
	"os"
	"strings"
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

	request_ip_and_port := strings.Split(r.RemoteAddr, ":")

	request_ip := request_ip_and_port[0]
	// request_port := request_ip_and_port[1] // Talán később használva lesz

	if isFilePathValid(path) {

		fmt.Printf("\n[200] OK | %s - %s (%s)\n>", r.Method, path, request_ip)
	} else {
		fmt.Printf("\n[404] File not found | %s - %s (%s)\n>", r.Method, r.URL.Path, request_ip)
	}
}

/*
A port pointerként van eltárolva, így csekkoljuk le, hogy létezik-e. Ez annak a másolását segíti elő
*/
func copyPortPtr(v *int) *int {
	if v == nil {
		return nil
	}
	x := *v
	return &x
}

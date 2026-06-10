package server

import (
	"fmt"
	"net/http"
	"os"
	"strings"
)

func isFilePathValid(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {

		return false
	}
	return true
}

func logToCLI(path string, r *http.Request) {

	request_ip_and_port := strings.Split(r.RemoteAddr, ":")

	request_ip := request_ip_and_port[0]
	// request_port := request_ip_and_port[1]

	if isFilePathValid(path) {

		fmt.Printf("[200] OK | %s - %s (%s)\n", r.Method, path, request_ip)
	} else {
		fmt.Printf("[404] File not found | %s - %s (%s)\n", r.Method, path, request_ip)
	}
}

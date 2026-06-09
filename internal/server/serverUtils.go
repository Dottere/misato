package server

import (
	"fmt"
	"net/http"
	"os"
)

func isFilePathValid(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {

		return false
	}
	return true
}

func handleFileServing(path string, w http.ResponseWriter, r *http.Request) {
	if isFilePathValid(path) {
		http.ServeFile(w, r, path)
		fmt.Printf("[200] OK | %s - %s", r.Method, path)
	} else {
		http.NotFound(w, r)
		fmt.Printf("[404] File not found | %s - %s", r.Method, path)
	}
}

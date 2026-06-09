package main

import (
	"misato/internal/server"
)

func main() {
	server.OpenCBZ("files\\test.cbz")
	server.Start(8080)
}

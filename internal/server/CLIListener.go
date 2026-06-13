package server

import (
	"bufio"
	"fmt"
	"log"
	"misato/config"
	"misato/internal/memstat"
	"os"
	"runtime"
	"strings"
	"time"
)

/*
Megnyitja a konzolt amin keresztül lehet vezérelni a szervert parancsok segítségével
*/
func Listen(srv *AppServer) {

	cfg := srv.cfg

	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("Server started. Type a command ('help' for commands or 'exit' to quit):")

	for {
		fmt.Print("\n> ")

		if !scanner.Scan() {
			fmt.Println("Program terminated (SIGINT)")
			os.Exit(0)
		}

		if err := scanner.Err(); err != nil {
			log.Printf("Console input error: %v\n", err)
			return
		}

		text := scanner.Text()

		text = strings.TrimSpace(text)

		handleSingleWordCommands(srv, cfg, text)
	}
}

func handleSingleWordCommands(srv *AppServer, cfg config.Config, cmd string) {
	switch cmd {
	case "help":
		fmt.Println(`
help - Print this help
exit (alias stop) - Shut down server
restart - Hot restart server
ip - Get bound ip address
port - Get bound port
uptime - Get server uptime
list - Get a list of all comics loaded
count - Get a number of how many comics are loaded
rescan - Initiates a library rescan
stats - Prints memory statistics
ping - pong!
clear - Clear terminal
config - Print current config
routes - Print registered routes
debug - Toggle debug mode
gc - Force runtime garbage collection`)
	case "exit", "stop":
		fmt.Println("\nShutting down server...")
		srv.Stop()
		os.Exit(0)
	case "restart":
		srv.Restart()
	case "ip":
		fmt.Printf("\nServer is bound to address %s\n", cfg.BindAddress)
	case "port":
		fmt.Printf("\nServer is bound to port %d\n", *cfg.ServerPort)
	case "uptime":
		fmt.Printf("\n%s", getCurrentUptime())
	case "list":
		fmt.Printf("\nStored comics:\n\n")
		for idx, elem := range srv.getAllStoredComics() {
			fmt.Printf("(%d) %s\n", idx+1, elem.Title)
		}
	case "count":
		fmt.Printf("\nLoaded comics: (%d)\n", len(srv.storedItems))
	case "rescan":
		fmt.Println("\nInitiating rescan...")
		srv.scan()
	case "stats":
		memstat.PrintStats()
	case "ping":
		fmt.Println("\npong!")
	case "clear":
		fmt.Print("\033[H\033[2J")
	case "config":
		fmt.Printf("\nServer port: %d\n", *cfg.ServerPort)
		fmt.Printf("Library folder: %s\n", cfg.FilesDir)
		fmt.Printf("Debug mode: %t\n", cfg.DebugMode)
		fmt.Printf("Server IP: %s\n", cfg.BindAddress)
		fmt.Printf("Read timeout: %s\n", time.Duration(cfg.ReadTimeout).String())
		fmt.Printf("Write timeout: %s\n", time.Duration(cfg.WriteTimeout).String())
		fmt.Printf("Idle Timeout: %s\n", time.Duration(cfg.IdleTimeout).String())
	case "routes":
		fmt.Println("\nActive Routes:")
		for _, r := range srv.GetRoutes() {
			fmt.Printf("- %s\n", r.Pattern)
		}
	case "debug":
		srv.coreMutex.Lock()
		srv.cfg.DebugMode = !srv.cfg.DebugMode
		fmt.Printf("\nDebug mode set to: %t\n", srv.cfg.DebugMode)
		srv.coreMutex.Unlock()
	case "gc":
		fmt.Println("\nForcing garbage collection...")
		runtime.GC()
		fmt.Println("Done.")
	case "":
		return
	default:
		fmt.Println("Unknown command:", cmd)
	}
}

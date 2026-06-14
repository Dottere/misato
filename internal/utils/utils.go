package utils

import (
	"flag"
	"fmt"
	"os"
	"runtime"
)

func ParseFlags(VERSION string) (string, int, bool) {
	var showVersion bool
	var configFilePath string
	var port int
	var verboseMode bool

	flag.BoolVar(&showVersion, "version", false, "Print the version number\n")
	flag.BoolVar(&showVersion, "v", false, "Print the version number (shorthand)\n")

	flag.BoolVar(&verboseMode, "verbose", false, "Enable verbose startup\n")

	flag.StringVar(&configFilePath, "config", "./config.json", "Path to the config file")
	flag.StringVar(&configFilePath, "c", "./config.json", "Path to the config file (shorthand)")

	flag.IntVar(&port, "port", 0, "Set server port")
	flag.IntVar(&port, "p", 0, "Set server port (shorthand)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "MISATO - Self hosted manga site\n\nUsage:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if showVersion {
		fmt.Printf("MISATO version %s\n", VERSION)
		os.Exit(0)
	}

	return configFilePath, port, verboseMode
}

/*
Segédfüggvény ami memóriastatisztikákat ír ki a szerverről
*/
func PrintMemoryStats() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	fmt.Println("\n=== Memory Statistics ===")

	fmt.Printf("Alloc      = %v MiB (Active memory)\n", bToMb(m.Alloc))
	fmt.Printf("TotalAlloc = %v MiB (Memory allocated since startup)\n", bToMb(m.TotalAlloc))
	fmt.Printf("Sys        = %v MiB (Pre-fetched memory)\n", bToMb(m.Sys))

	// Részletek
	fmt.Printf("HeapObjects = %v (Existing heap objects)\n", m.HeapObjects)

	// Szemétgyűjtő (GC)
	fmt.Printf("Garbage collections = %v times\n", m.NumGC)
	fmt.Printf("Garbage collector pause time = %v ms (PauseTotalNs)\n", m.PauseTotalNs/1_000_000)

	// Goroutine-ok
	fmt.Printf("Goroutines = %v active threads\n", runtime.NumGoroutine())
	fmt.Println("============================")
}

// Segédfüggvény a bájtok megabájttá alakításához
func bToMb(b uint64) uint64 {
	return b / (1 << 20)
}

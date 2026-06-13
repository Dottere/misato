package memstat

import (
	"fmt"
	"runtime"
)

/*
Segédfüggvény ami memóriastatisztikákat ír ki a szerverről
*/
func PrintStats() {
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

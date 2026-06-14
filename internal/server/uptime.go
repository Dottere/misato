package server

import (
	"fmt"
	"time"
)

// initUptime beállítja a megadott memóriahelynél a szerver indulásának (vagy újraindulásának)
// pontos idejét. Ezt a memóriacímet jellemzően a központi AppServer struktúra tárolja.
func initUptime(startTime *time.Time) {
	*startTime = time.Now()
}

// getCurrentUptime kiszámolja az indulás óta eltelt időt, és visszaad egy ember által
// olvasható, formázott stringet.
// Példa kimenet:
// Server started on 2026. June 14.
// Running for 02h 45m 10s
func getCurrentUptime(startTime *time.Time) string {
	duration := time.Since(*startTime)

	hours := duration / time.Hour
	minutes := (duration % time.Hour) / time.Minute
	seconds := (duration % time.Minute) / time.Second

	// A Go beépített time.Format funkciója a "2005. July 01." referencia dátum
	// alapján mintaként formázza a mi startTime-unkat.
	startStr := startTime.Format("2005. July 01.")
	runningStr := fmt.Sprintf("%02dh %02dm %02ds", hours, minutes, seconds)

	return fmt.Sprintf("Server started on %s\nRunning for %s", startStr, runningStr)
}

package server

import (
	"fmt"
	"time"
)

// A szerver elindulásakor kimentődik az akkori időpont és eltárolódik ebben a változóban
var startTime time.Time

/*
Kimenti a jelenlegi időt egy startTime nevű globális változóba
*/
func initUptime() {
	startTime = time.Now()
}

/*
A startTime globális változó alapján meghatározza, hogy mennyi idő telt el az initUptime() hívása óta
*/
func getCurrentUptime() string {
	d := time.Since(startTime)

	h := d / time.Hour
	m := (d % time.Hour) / time.Minute
	s := (d % time.Minute) / time.Second

	startStr := fmt.Sprintf("%d. %s %02d.", startTime.Year(), startTime.Month(), startTime.Day())

	runningStr := fmt.Sprintf("%02dh %02dm %02ds", h, m, s)

	return fmt.Sprintf("Server started on %s\nRunning for %s", startStr, runningStr)
}

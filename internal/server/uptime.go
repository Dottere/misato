package server

import (
	"fmt"
	"time"
)

var startTime time.Time

func initUptime() {
	startTime = time.Now()
}

func getCurrentUptime() string {
	d := time.Since(startTime)

	h := d / time.Hour
	m := (d % time.Hour) / time.Minute
	s := (d % time.Minute) / time.Second

	startStr := fmt.Sprintf("%d. %s %02d.", startTime.Year(), startTime.Month(), startTime.Day())

	runningStr := fmt.Sprintf("%02dh %02dm %02ds", h, m, s)

	return fmt.Sprintf("Server started on %s\nRunning for %s", startStr, runningStr)
}

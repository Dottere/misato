package server

import (
	"fmt"
	"time"
)

func initUptime(startTime *time.Time) {
	*startTime = time.Now()
}

func getCurrentUptime(startTime *time.Time) string {
	duration := time.Since(*startTime)

	hours := duration / time.Hour
	minutes := (duration % time.Hour) / time.Minute
	seconds := (duration % time.Minute) / time.Second

	startStr := fmt.Sprintf("%d. %s %02d.", startTime.Year(), startTime.Month(), startTime.Day())
	runningStr := fmt.Sprintf("%02dh %02dm %02ds", hours, minutes, seconds)

	return fmt.Sprintf("Server started on %s\nRunning for %s", startStr, runningStr)
}

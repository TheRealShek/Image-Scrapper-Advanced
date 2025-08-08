package internal

import (
	"time"
)

// ScheduleTask runs a function at a given interval (simple cron replacement).
func ScheduleTask(interval time.Duration, task func()) chan struct{} {
	stop := make(chan struct{})
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				task()
			case <-stop:
				return
			}
		}
	}()
	return stop
}

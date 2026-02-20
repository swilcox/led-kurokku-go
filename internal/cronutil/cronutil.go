package cronutil

import (
	"log"
	"time"

	"github.com/robfig/cron/v3"
)

// MatchesNow returns true if time t falls on a scheduled minute for expr.
// Returns false (and logs) if expr is invalid.
func MatchesNow(expr string, t time.Time) bool {
	sched, err := cron.ParseStandard(expr)
	if err != nil {
		log.Printf("invalid cron expression %q: %v", expr, err)
		return false
	}
	minute := t.Truncate(time.Minute)
	return sched.Next(minute.Add(-time.Nanosecond)).Equal(minute)
}

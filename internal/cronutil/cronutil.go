package cronutil

import (
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"
)

// MatchesNow returns true if time t falls on a scheduled minute for expr.
// Returns false (and logs) if expr is invalid.
func MatchesNow(expr string, t time.Time) bool {
	sched, err := cron.ParseStandard(expr)
	if err != nil {
		slog.Warn("invalid cron expression", "expr", expr, "error", err)
		return false
	}
	minute := t.Truncate(time.Minute)
	return sched.Next(minute.Add(-time.Nanosecond)).Equal(minute)
}

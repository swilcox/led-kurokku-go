package widget

import (
	"context"
	"log/slog"
	"time"

	"github.com/swilcox/led-kurokku-go/config"
	"github.com/swilcox/led-kurokku-go/display"
)

// AlertFetcher fetches and deletes alerts from an external source.
type AlertFetcher interface {
	FetchAlerts(ctx context.Context) ([]config.AlertConfig, error)
	DeleteAlert(ctx context.Context, id string) error
}

// RedisAlert wraps alert display with Redis-backed alert fetching.
// On each Run it fetches alerts from Redis, falling back to the
// configured JSON alerts on error.
type RedisAlert struct {
	Fetcher     AlertFetcher
	Fallback    []config.AlertConfig
	ScrollSpeed time.Duration
}

func (ra *RedisAlert) Name() string { return "redis-alert" }

func (ra *RedisAlert) Run(ctx context.Context, disp display.Display) error {
	alerts, err := ra.Fetcher.FetchAlerts(ctx)
	if err != nil {
		slog.Warn("redis alert fetch failed, using fallback", "error", err)
		alerts = ra.Fallback
	}

	a := &Alert{
		Alerts:      alerts,
		ScrollSpeed: ra.ScrollSpeed,
		OnDelete: func(ctx context.Context, id string) {
			if err := ra.Fetcher.DeleteAlert(ctx, id); err != nil {
				slog.Error("redis alert delete failed", "id", id, "error", err)
			}
		},
	}
	return a.Run(ctx, disp)
}

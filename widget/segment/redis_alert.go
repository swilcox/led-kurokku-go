package segment

import (
	"context"
	"log/slog"
	"time"

	"github.com/swilcox/led-kurokku-go/config"
	"github.com/swilcox/led-kurokku-go/display"
	"github.com/swilcox/led-kurokku-go/segfont"
	"github.com/swilcox/led-kurokku-go/widget"
)

// RedisAlert wraps segment alert display with Redis-backed alert fetching.
type RedisAlert struct {
	Fetcher     widget.AlertFetcher
	Fallback    []config.AlertConfig
	ScrollSpeed time.Duration
	Encoder     segfont.Encoder
}

func (ra *RedisAlert) Name() string { return "segment-redis-alert" }

func (ra *RedisAlert) Run(ctx context.Context, disp display.Display) error {
	alerts, err := ra.Fetcher.FetchAlerts(ctx)
	if err != nil {
		slog.Warn("redis alert fetch failed, using fallback", "error", err)
		alerts = ra.Fallback
	} else if len(alerts) == 0 {
		slog.Debug("redis alert no active alerts found")
	}

	a := &Alert{
		Alerts:      alerts,
		ScrollSpeed: ra.ScrollSpeed,
		Encoder:     ra.Encoder,
		OnDelete: func(ctx context.Context, id string) {
			if err := ra.Fetcher.DeleteAlert(ctx, id); err != nil {
				slog.Error("redis alert delete failed", "id", id, "error", err)
			}
		},
	}
	return a.Run(ctx, disp)
}

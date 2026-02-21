package segment

import (
	"context"
	"sort"
	"time"

	"github.com/swilcox/led-kurokku-go/config"
	"github.com/swilcox/led-kurokku-go/display"
	"github.com/swilcox/led-kurokku-go/internal/cronutil"
	"github.com/swilcox/led-kurokku-go/segfont"
)

// Alert displays prioritized alert messages on a segment display.
type Alert struct {
	Alerts      []config.AlertConfig
	ScrollSpeed time.Duration
	OnDelete    func(ctx context.Context, id string)
	NowFunc     func() time.Time
	Encoder     segfont.Encoder
}

func (a *Alert) now() time.Time {
	if a.NowFunc != nil {
		return a.NowFunc()
	}
	return time.Now()
}

func (a *Alert) Name() string { return "segment-alert" }

func (a *Alert) Run(ctx context.Context, disp display.Display) error {
	if len(a.Alerts) == 0 {
		return nil
	}

	sorted := make([]int, len(a.Alerts))
	for i := range sorted {
		sorted[i] = i
	}
	sort.SliceStable(sorted, func(i, j int) bool {
		return a.Alerts[sorted[i]].Priority < a.Alerts[sorted[j]].Priority
	})

	var toDelete []int
	for _, idx := range sorted {
		alert := a.Alerts[idx]
		if alert.Priority == 10 && !cronutil.MatchesNow("*/10 * * * *", a.now()) {
			continue
		}
		dur := alert.DisplayDuration.Unwrap()
		if dur == 0 {
			dur = 5 * time.Second
		}

		alertCtx, cancel := context.WithTimeout(ctx, dur)

		msg := &Message{
			Text:        alert.Message,
			ScrollSpeed: a.ScrollSpeed,
			Repeats:     -1,
			Encoder:     a.Encoder,
		}
		msg.Run(alertCtx, disp)
		cancel()

		if ctx.Err() != nil {
			return ctx.Err()
		}

		if alert.DeleteAfterDisplay {
			if a.OnDelete != nil {
				a.OnDelete(ctx, alert.ID)
			} else {
				toDelete = append(toDelete, idx)
			}
		}
	}

	if a.OnDelete == nil {
		sort.Sort(sort.Reverse(sort.IntSlice(toDelete)))
		for _, idx := range toDelete {
			a.Alerts = append(a.Alerts[:idx], a.Alerts[idx+1:]...)
		}
	}

	return nil
}

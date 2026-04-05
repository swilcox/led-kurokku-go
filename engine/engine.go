package engine

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/nathan-osman/go-sunrise"
	"github.com/swilcox/led-kurokku-go/config"
	"github.com/swilcox/led-kurokku-go/display"
	"github.com/swilcox/led-kurokku-go/internal/cronutil"
	"github.com/swilcox/led-kurokku-go/redis"
	"github.com/swilcox/led-kurokku-go/segfont"
	"github.com/swilcox/led-kurokku-go/widget"
	"github.com/swilcox/led-kurokku-go/widget/animation"
	"github.com/swilcox/led-kurokku-go/widget/segment"
)

// redisStore is the interface the engine uses to interact with Redis.
// Using an interface allows engine tests to inject a mock without a real Redis server.
type redisStore interface {
	FetchAlerts(ctx context.Context) ([]config.AlertConfig, error)
	DeleteAlert(ctx context.Context, id string) error
	FetchMessageText(ctx context.Context, key string) (string, bool, error)
	SubscribeAlerts(ctx context.Context) (<-chan struct{}, error)
}

// Engine manages the widget cycling loop.
type Engine struct {
	disp    display.Display
	cfg     *config.Config
	rds     redisStore
	nowFunc func() time.Time
}

func (e *Engine) now() time.Time {
	if e.nowFunc != nil {
		return e.nowFunc()
	}
	return time.Now()
}

// New creates a new engine with the given display, config, and optional Redis client.
func New(disp display.Display, cfg *config.Config, rds *redis.Client) *Engine {
	e := &Engine{disp: disp, cfg: cfg}
	if rds != nil {
		e.rds = rds
	}
	return e
}

// Run starts the widget cycling loop. It blocks until ctx is cancelled.
func (e *Engine) Run(ctx context.Context) error {
	widgets, durations, crons := e.buildWidgets()
	if len(widgets) == 0 {
		return fmt.Errorf("no enabled widgets configured")
	}

	// Start brightness control goroutine
	go e.brightnessLoop(ctx)

	// Subscribe for alert interrupts if Redis is available.
	var alertCh <-chan struct{}
	if e.rds != nil {
		var err error
		alertCh, err = e.rds.SubscribeAlerts(ctx)
		if err != nil {
			slog.Warn("redis alert subscribe failed, interrupts disabled", "error", err)
		}
	}

	hasAlertWidget := e.hasAlertWidget()

	for {
	cycle:
		for i, w := range widgets {
			if ctx.Err() != nil {
				return nil
			}
			if crons[i] != "" && !cronutil.MatchesNow(crons[i], e.now()) {
				continue
			}
			slog.Info("widget", "name", w.Name())
			var wctx context.Context
			var cancel context.CancelFunc
			if durations[i] > 0 {
				wctx, cancel = context.WithTimeout(ctx, durations[i])
			} else {
				wctx, cancel = context.WithCancel(ctx)
			}

			// Run widget in a goroutine so we can select on interrupt.
			done := make(chan struct{})
			go func() {
				w.Run(wctx, e.disp)
				close(done)
			}()

			select {
			case <-done:
				cancel()
			case <-alertCh:
				// Alert interrupt: cancel current widget.
				cancel()
				<-done // wait for widget goroutine to finish
				if hasAlertWidget {
					// Restart the widget cycle so the configured alert
					// widget picks up the new alert naturally.
					slog.Info("alert interrupt, restarting widget cycle")
					break cycle
				}
				// No alert widget configured; display alerts ad-hoc.
				slog.Info("alert interrupt, no alert widget configured, displaying ad-hoc")
				e.runInterruptAlerts(ctx)
			case <-ctx.Done():
				cancel()
				<-done
				return nil
			}

			if ctx.Err() != nil {
				return nil
			}
		}
	}
}

// hasAlertWidget reports whether any configured widget is of type "alert".
func (e *Engine) hasAlertWidget() bool {
	for _, wc := range e.cfg.Widgets {
		if wc.Type == "alert" {
			return true
		}
	}
	return false
}

// runInterruptAlerts fetches alerts from Redis and displays them.
func (e *Engine) runInterruptAlerts(ctx context.Context) {
	if e.rds == nil {
		return
	}

	alerts, err := e.rds.FetchAlerts(ctx)
	if err != nil {
		slog.Error("redis interrupt fetch failed", "error", err)
		return
	}
	if len(alerts) == 0 {
		return
	}

	timeout := time.Duration(len(alerts)) * 10 * time.Second
	alertCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	scrollSpeed := e.cfg.Display.DefaultScrollSpeed.Unwrap()
	if e.cfg.Display.IsSegment() {
		a := &segment.Alert{
			Alerts:      alerts,
			ScrollSpeed: scrollSpeed,
			Encoder:     e.segmentEncoder(),
			OnDelete: func(ctx context.Context, id string) {
				if err := e.rds.DeleteAlert(ctx, id); err != nil {
					slog.Error("redis alert delete failed", "id", id, "error", err)
				}
			},
		}
		a.Run(alertCtx, e.disp)
	} else {
		a := &widget.Alert{
			Alerts:      alerts,
			ScrollSpeed: scrollSpeed,
			OnDelete: func(ctx context.Context, id string) {
				if err := e.rds.DeleteAlert(ctx, id); err != nil {
					slog.Error("redis alert delete failed", "id", id, "error", err)
				}
			},
		}
		a.Run(alertCtx, e.disp)
	}
}

// segmentAnimationFactory returns the appropriate animation constructor for the
// current display type. For 14-segment displays it prefers Registry14 variants,
// falling back to the base Registry.
func (e *Engine) segmentAnimationFactory(name string) (func() widget.Widget, bool) {
	if e.is14Seg() {
		if factory, ok := segment.Registry14[name]; ok {
			return factory, true
		}
	}
	factory, ok := segment.Registry[name]
	return factory, ok
}

func (e *Engine) is14Seg() bool {
	switch e.cfg.Display.Type {
	case config.DisplayHT16K33, config.DisplayTerminalSeg14:
		return true
	}
	return false
}

func (e *Engine) segmentEncoder() segfont.Encoder {
	switch e.cfg.Display.Type {
	case config.DisplayTM1637, config.DisplayTerminalSeg7:
		return segfont.Enc7
	default:
		return segfont.Enc14
	}
}

// scrollSpeed returns the widget's scroll speed, falling back to the display default.
func (e *Engine) scrollSpeed(wc config.WidgetConfig) time.Duration {
	if s := wc.ScrollSpeed.Unwrap(); s != 0 {
		return s
	}
	return e.cfg.Display.DefaultScrollSpeed.Unwrap()
}

func (e *Engine) buildWidgets() ([]widget.Widget, []time.Duration, []string) {
	var widgets []widget.Widget
	var durations []time.Duration
	var crons []string
	isSeg := e.cfg.Display.IsSegment()

	for _, wc := range e.cfg.Widgets {
		if !wc.Enabled {
			continue
		}

		var w widget.Widget
		switch wc.Type {
		case "clock":
			format24h := true
			if wc.Format24h != nil {
				format24h = *wc.Format24h
			}
			if isSeg {
				w = &segment.Clock{Format24h: format24h, Encoder: e.segmentEncoder()}
			} else {
				w = &widget.Clock{Format24h: format24h}
			}

		case "message":
			repeats := 1
			if wc.Repeats != nil {
				repeats = *wc.Repeats
			}
			if isSeg {
				if e.rds != nil && wc.DynamicSource != "" {
					w = &segment.RedisMessage{
						Fetcher:      e.rds,
						Key:          wc.DynamicSource,
						FallbackText: wc.Text,
						ScrollSpeed:  e.scrollSpeed(wc),
						Repeats:      repeats,
						SleepBetween: wc.SleepBetween.Unwrap(),
						Encoder:      e.segmentEncoder(),
					}
				} else {
					w = &segment.Message{
						Text:         wc.Text,
						ScrollSpeed:  e.scrollSpeed(wc),
						Repeats:      repeats,
						SleepBetween: wc.SleepBetween.Unwrap(),
						Encoder:      e.segmentEncoder(),
					}
				}
			} else {
				if e.rds != nil && wc.DynamicSource != "" {
					w = &widget.RedisMessage{
						Fetcher:      e.rds,
						Key:          wc.DynamicSource,
						FallbackText: wc.Text,
						ScrollSpeed:  e.scrollSpeed(wc),
						Repeats:      repeats,
						SleepBetween: wc.SleepBetween.Unwrap(),
					}
				} else {
					w = &widget.Message{
						Text:         wc.Text,
						ScrollSpeed:  e.scrollSpeed(wc),
						Repeats:      repeats,
						SleepBetween: wc.SleepBetween.Unwrap(),
					}
				}
			}

		case "alert":
			if isSeg {
				if e.rds != nil {
					w = &segment.RedisAlert{
						Fetcher:     e.rds,
						Fallback:    wc.Alerts,
						ScrollSpeed: e.scrollSpeed(wc),
						Encoder:     e.segmentEncoder(),
					}
				} else {
					w = &segment.Alert{
						Alerts:      wc.Alerts,
						ScrollSpeed: e.scrollSpeed(wc),
						Encoder:     e.segmentEncoder(),
					}
				}
			} else {
				if e.rds != nil {
					w = &widget.RedisAlert{
						Fetcher:     e.rds,
						Fallback:    wc.Alerts,
						ScrollSpeed: e.scrollSpeed(wc),
					}
				} else {
					w = &widget.Alert{
						Alerts:      wc.Alerts,
						ScrollSpeed: e.scrollSpeed(wc),
					}
				}
			}

		case "animation":
			if isSeg {
				if wc.AnimationType == "frames" || wc.AnimationType == "" {
					w = &segment.FrameAnimation{
						Frames:        wc.SegmentFrames,
						FrameDuration: wc.FrameDuration.Unwrap(),
					}
				} else if factory, ok := e.segmentAnimationFactory(wc.AnimationType); ok {
					w = factory()
				} else {
					slog.Warn("unknown segment animation type", "type", wc.AnimationType)
					continue
				}
			} else {
				if wc.AnimationType == "frames" || wc.AnimationType == "" {
					w = &animation.FrameAnimation{
						Frames:        wc.Frames,
						FrameDuration: wc.FrameDuration.Unwrap(),
					}
				} else if factory, ok := animation.Registry[wc.AnimationType]; ok {
					w = factory()
				} else {
					slog.Warn("unknown animation type", "type", wc.AnimationType)
					continue
				}
			}

		default:
			slog.Warn("unknown widget type", "type", wc.Type)
			continue
		}

		widgets = append(widgets, w)
		durations = append(durations, wc.Duration.Unwrap())
		crons = append(crons, wc.Cron)
	}
	return widgets, durations, crons
}

func (e *Engine) brightnessLoop(ctx context.Context) {
	e.updateBrightness()
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			e.updateBrightness()
		}
	}
}

func (e *Engine) updateBrightness() {
	bc := e.cfg.Brightness
	now := e.now()
	if bc.UseLocation {
		e.updateBrightnessFromLocation(now, bc)
		return
	}
	e.updateBrightnessFromTimes(now, bc)
}

func (e *Engine) updateBrightnessFromLocation(now time.Time, bc config.BrightnessConfig) {
	loc := e.cfg.Location
	if loc == nil {
		slog.Warn("brightness: use_location enabled but no location configured, defaulting to high")
		e.disp.SetBrightness(bc.High)
		return
	}
	tz, err := time.LoadLocation(loc.Timezone)
	if err != nil {
		slog.Warn("brightness: invalid timezone, defaulting to high", "timezone", loc.Timezone, "error", err)
		e.disp.SetBrightness(bc.High)
		return
	}
	nowLocal := now.In(tz)
	rise, set := sunrise.SunriseSunset(loc.Lat, loc.Lon, nowLocal.Year(), nowLocal.Month(), nowLocal.Day())
	if now.After(rise) && now.Before(set) {
		e.disp.SetBrightness(bc.High)
	} else {
		e.disp.SetBrightness(bc.Low)
	}
}

func (e *Engine) updateBrightnessFromTimes(now time.Time, bc config.BrightnessConfig) {
	dayStart, err1 := time.Parse("15:04", bc.DayStart)
	dayEnd, err2 := time.Parse("15:04", bc.DayEnd)
	if err1 != nil || err2 != nil {
		e.disp.SetBrightness(bc.High)
		return
	}
	nowMinutes := now.Hour()*60 + now.Minute()
	startMinutes := dayStart.Hour()*60 + dayStart.Minute()
	endMinutes := dayEnd.Hour()*60 + dayEnd.Minute()
	if nowMinutes >= startMinutes && nowMinutes < endMinutes {
		e.disp.SetBrightness(bc.High)
	} else {
		e.disp.SetBrightness(bc.Low)
	}
}

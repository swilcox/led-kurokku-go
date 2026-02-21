package engine

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/nathan-osman/go-sunrise"
	"github.com/swilcox/led-kurokku-go/config"
	"github.com/swilcox/led-kurokku-go/display"
	"github.com/swilcox/led-kurokku-go/internal/cronutil"
	"github.com/swilcox/led-kurokku-go/redis"
	"github.com/swilcox/led-kurokku-go/widget"
	"github.com/swilcox/led-kurokku-go/widget/animation"
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
			log.Printf("redis alert subscribe failed, interrupts disabled: %v", err)
		}
	}

	for {
		for i, w := range widgets {
			if ctx.Err() != nil {
				return nil
			}
			if crons[i] != "" && !cronutil.MatchesNow(crons[i], e.now()) {
				continue
			}
			log.Printf("widget: %s", w.Name())
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
				// Alert interrupt: cancel current widget and show alerts.
				cancel()
				<-done // wait for widget goroutine to finish
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

// runInterruptAlerts fetches alerts from Redis and displays them.
func (e *Engine) runInterruptAlerts(ctx context.Context) {
	if e.rds == nil {
		return
	}

	alerts, err := e.rds.FetchAlerts(ctx)
	if err != nil {
		log.Printf("redis interrupt fetch failed: %v", err)
		return
	}
	if len(alerts) == 0 {
		return
	}

	a := &widget.Alert{
		Alerts:      alerts,
		ScrollSpeed: 50 * time.Millisecond,
		OnDelete: func(ctx context.Context, id string) {
			if err := e.rds.DeleteAlert(ctx, id); err != nil {
				log.Printf("redis alert delete %s: %v", id, err)
			}
		},
	}

	// Give alerts a generous timeout so they can all display.
	timeout := time.Duration(len(alerts)) * 10 * time.Second
	alertCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	a.Run(alertCtx, e.disp)
}

func (e *Engine) buildWidgets() ([]widget.Widget, []time.Duration, []string) {
	var widgets []widget.Widget
	var durations []time.Duration
	var crons []string

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
			w = &widget.Clock{Format24h: format24h}

		case "message":
			repeats := 1
			if wc.Repeats != nil {
				repeats = *wc.Repeats
			}
			if e.rds != nil && wc.DynamicSource != "" {
				w = &widget.RedisMessage{
					Fetcher:      e.rds,
					Key:          wc.DynamicSource,
					FallbackText: wc.Text,
					ScrollSpeed:  wc.ScrollSpeed.Unwrap(),
					Repeats:      repeats,
					SleepBetween: wc.SleepBetween.Unwrap(),
				}
			} else {
				w = &widget.Message{
					Text:         wc.Text,
					ScrollSpeed:  wc.ScrollSpeed.Unwrap(),
					Repeats:      repeats,
					SleepBetween: wc.SleepBetween.Unwrap(),
				}
			}

		case "alert":
			if e.rds != nil {
				w = &widget.RedisAlert{
					Fetcher:     e.rds,
					Fallback:    wc.Alerts,
					ScrollSpeed: wc.ScrollSpeed.Unwrap(),
				}
			} else {
				w = &widget.Alert{
					Alerts:      wc.Alerts,
					ScrollSpeed: wc.ScrollSpeed.Unwrap(),
				}
			}

		case "animation":
			if wc.AnimationType == "frames" || wc.AnimationType == "" {
				w = &animation.FrameAnimation{
					Frames:        wc.Frames,
					FrameDuration: wc.FrameDuration.Unwrap(),
				}
			} else if factory, ok := animation.Registry[wc.AnimationType]; ok {
				w = factory()
			} else {
				log.Printf("unknown animation type: %s", wc.AnimationType)
				continue
			}

		default:
			log.Printf("unknown widget type: %s", wc.Type)
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
		log.Print("brightness: use_location=true but no location configured, defaulting to high brightness")
		e.disp.SetBrightness(bc.High)
		return
	}
	tz, err := time.LoadLocation(loc.Timezone)
	if err != nil {
		log.Printf("brightness: invalid timezone %q: %v, defaulting to high brightness", loc.Timezone, err)
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

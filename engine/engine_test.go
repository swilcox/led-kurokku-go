package engine

import (
	"context"
	"testing"
	"time"

	"github.com/swilcox/led-kurokku-go/config"
	"github.com/swilcox/led-kurokku-go/display/testutil"
)

// mockRedis implements redisStore for testing without a real Redis server.
type mockRedis struct {
	alerts []config.AlertConfig
	err    error
}

func (m *mockRedis) FetchAlerts(_ context.Context) ([]config.AlertConfig, error) {
	return m.alerts, m.err
}

func (m *mockRedis) DeleteAlert(_ context.Context, _ string) error { return nil }

func (m *mockRedis) FetchMessageText(_ context.Context, _ string) (string, bool, error) {
	return "", false, nil
}

func (m *mockRedis) SubscribeAlerts(_ context.Context) (<-chan struct{}, error) {
	return make(chan struct{}), nil
}

func brightnessCfg() config.BrightnessConfig {
	return config.BrightnessConfig{
		High:     15,
		Low:      1,
		DayStart: "08:00",
		DayEnd:   "22:00",
	}
}

func TestUpdateBrightness_DayTime(t *testing.T) {
	spy := &testutil.SpyDisplay{}
	cfg := &config.Config{Brightness: brightnessCfg()}

	e := New(spy, cfg, nil)
	e.nowFunc = func() time.Time {
		return time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC) // noon → day
	}

	e.updateBrightness()

	if len(spy.Brightness) == 0 {
		t.Fatal("expected SetBrightness to be called")
	}
	if spy.Brightness[0] != 15 {
		t.Errorf("expected brightness 15 (high/day), got %d", spy.Brightness[0])
	}
}

func TestUpdateBrightness_NightTime(t *testing.T) {
	spy := &testutil.SpyDisplay{}
	cfg := &config.Config{Brightness: brightnessCfg()}

	e := New(spy, cfg, nil)
	e.nowFunc = func() time.Time {
		return time.Date(2024, 1, 1, 2, 0, 0, 0, time.UTC) // 2 AM → night
	}

	e.updateBrightness()

	if len(spy.Brightness) == 0 {
		t.Fatal("expected SetBrightness to be called")
	}
	if spy.Brightness[0] != 1 {
		t.Errorf("expected brightness 1 (low/night), got %d", spy.Brightness[0])
	}
}

func TestUpdateBrightness_InvalidTimes_UsesHigh(t *testing.T) {
	spy := &testutil.SpyDisplay{}
	cfg := &config.Config{
		Brightness: config.BrightnessConfig{
			High: 15, Low: 1,
			DayStart: "bad", DayEnd: "also-bad",
		},
	}

	e := New(spy, cfg, nil)
	e.updateBrightness()

	if len(spy.Brightness) == 0 {
		t.Fatal("expected SetBrightness to be called")
	}
	if spy.Brightness[0] != 15 {
		t.Errorf("expected brightness 15 (high) when times are invalid, got %d", spy.Brightness[0])
	}
}

func TestEngine_Run_CancelledByContext(t *testing.T) {
	spy := &testutil.SpyDisplay{}
	cfg := &config.Config{
		Brightness: brightnessCfg(),
		Widgets: []config.WidgetConfig{
			{
				Type:    "message",
				Enabled: true,
				// Duration 0 → runs until context cancelled
				Text: "Hi",
			},
		},
	}

	e := New(spy, cfg, nil)
	e.nowFunc = func() time.Time {
		return time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := e.Run(ctx)
	if err != nil {
		t.Errorf("expected nil error on context cancellation, got %v", err)
	}
	if len(spy.Frames) == 0 {
		t.Error("expected at least one frame to be written before cancellation")
	}
}

func TestEngine_Run_NoWidgets_ReturnsError(t *testing.T) {
	spy := &testutil.SpyDisplay{}
	cfg := &config.Config{
		Brightness: brightnessCfg(),
		Widgets:    []config.WidgetConfig{}, // none enabled
	}

	e := New(spy, cfg, nil)
	err := e.Run(context.Background())
	if err == nil {
		t.Error("expected error when no widgets are configured")
	}
}

func TestEngine_New_NilRedis_DoesNotPanic(t *testing.T) {
	spy := &testutil.SpyDisplay{}
	cfg := &config.Config{Brightness: brightnessCfg()}

	e := New(spy, cfg, nil)
	if e.rds != nil {
		t.Error("expected e.rds to be nil when passed nil *redis.Client")
	}
}

func TestUpdateBrightness_UseLocation_Daytime(t *testing.T) {
	spy := &testutil.SpyDisplay{}
	cfg := &config.Config{
		Location: &config.LocationConfig{
			Lat: 36.166, Lon: -86.784, Timezone: "America/Chicago",
		},
		Brightness: config.BrightnessConfig{High: 15, Low: 1, UseLocation: true},
	}
	e := New(spy, cfg, nil)
	e.nowFunc = func() time.Time {
		// 3 PM UTC on summer solstice — well within daylight for Nashville
		return time.Date(2024, 6, 21, 15, 0, 0, 0, time.UTC)
	}
	e.updateBrightness()
	if len(spy.Brightness) == 0 {
		t.Fatal("expected SetBrightness to be called")
	}
	if spy.Brightness[0] != 15 {
		t.Errorf("expected high brightness (15) during daytime, got %d", spy.Brightness[0])
	}
}

func TestUpdateBrightness_UseLocation_Nighttime(t *testing.T) {
	spy := &testutil.SpyDisplay{}
	cfg := &config.Config{
		Location: &config.LocationConfig{
			Lat: 36.166, Lon: -86.784, Timezone: "America/Chicago",
		},
		Brightness: config.BrightnessConfig{High: 15, Low: 1, UseLocation: true},
	}
	e := New(spy, cfg, nil)
	e.nowFunc = func() time.Time {
		// 5 AM UTC on summer solstice — before sunrise in Nashville (~10:30 UTC)
		return time.Date(2024, 6, 21, 5, 0, 0, 0, time.UTC)
	}
	e.updateBrightness()
	if len(spy.Brightness) == 0 {
		t.Fatal("expected SetBrightness to be called")
	}
	if spy.Brightness[0] != 1 {
		t.Errorf("expected low brightness (1) during nighttime, got %d", spy.Brightness[0])
	}
}

func TestUpdateBrightness_UseLocation_MissingLocation_FallsBackToHigh(t *testing.T) {
	spy := &testutil.SpyDisplay{}
	cfg := &config.Config{
		Brightness: config.BrightnessConfig{High: 15, Low: 1, UseLocation: true},
	}
	e := New(spy, cfg, nil)
	e.updateBrightness()
	if len(spy.Brightness) == 0 {
		t.Fatal("expected SetBrightness to be called")
	}
	if spy.Brightness[0] != 15 {
		t.Errorf("expected high brightness fallback when location is nil, got %d", spy.Brightness[0])
	}
}

func TestUpdateBrightness_UseLocation_InvalidTimezone_FallsBackToHigh(t *testing.T) {
	spy := &testutil.SpyDisplay{}
	cfg := &config.Config{
		Location: &config.LocationConfig{
			Lat: 36.166, Lon: -86.784, Timezone: "Not/ATimezone",
		},
		Brightness: config.BrightnessConfig{High: 15, Low: 1, UseLocation: true},
	}
	e := New(spy, cfg, nil)
	e.updateBrightness()
	if len(spy.Brightness) == 0 {
		t.Fatal("expected SetBrightness to be called")
	}
	if spy.Brightness[0] != 15 {
		t.Errorf("expected high brightness fallback for invalid timezone, got %d", spy.Brightness[0])
	}
}

func TestEngine_Run_SkipsWidgetWithNonMatchingCron(t *testing.T) {
	spy := &testutil.SpyDisplay{}
	// "0 12 * * *" matches only at 12:00; nowFunc returns 10:00 — should not match
	cfg := &config.Config{
		Brightness: brightnessCfg(),
		Widgets: []config.WidgetConfig{
			{
				Type:     "message",
				Enabled:  true,
				Duration: config.Duration(50 * time.Millisecond),
				Text:     "cron-msg",
				Cron:     "0 12 * * *",
			},
		},
	}

	e := New(spy, cfg, nil)
	e.nowFunc = func() time.Time {
		return time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	e.Run(ctx) //nolint:errcheck

	if len(spy.Frames) != 0 {
		t.Errorf("expected no frames for non-matching cron widget, got %d", len(spy.Frames))
	}
}

func TestEngine_Run_RunsWidgetWithMatchingCron(t *testing.T) {
	spy := &testutil.SpyDisplay{}
	// "0 10 * * *" matches at 10:00; nowFunc returns 10:00 — should match
	cfg := &config.Config{
		Brightness: brightnessCfg(),
		Widgets: []config.WidgetConfig{
			{
				Type:     "message",
				Enabled:  true,
				Duration: config.Duration(50 * time.Millisecond),
				Text:     "cron-msg",
				Cron:     "0 10 * * *",
			},
		},
	}

	e := New(spy, cfg, nil)
	e.nowFunc = func() time.Time {
		return time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	e.Run(ctx) //nolint:errcheck

	if len(spy.Frames) == 0 {
		t.Error("expected frames for matching cron widget, got none")
	}
}

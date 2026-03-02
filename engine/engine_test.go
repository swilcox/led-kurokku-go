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

func TestBuildWidgets_ClockPixel(t *testing.T) {
	spy := &testutil.SpyDisplay{}
	cfg := &config.Config{
		Display:    config.DisplayConfig{Type: config.DisplayTerminal},
		Brightness: brightnessCfg(),
		Widgets: []config.WidgetConfig{
			{Type: "clock", Enabled: true, Duration: config.Duration(5 * time.Second)},
		},
	}
	e := New(spy, cfg, nil)
	widgets, durations, crons := e.buildWidgets()
	if len(widgets) != 1 {
		t.Fatalf("expected 1 widget, got %d", len(widgets))
	}
	if widgets[0].Name() != "clock" {
		t.Errorf("widget name = %q, want %q", widgets[0].Name(), "clock")
	}
	if durations[0] != 5*time.Second {
		t.Errorf("duration = %v, want 5s", durations[0])
	}
	if crons[0] != "" {
		t.Errorf("cron = %q, want empty", crons[0])
	}
}

func TestBuildWidgets_ClockSegment(t *testing.T) {
	spy := &testutil.SpySegmentDisplay{}
	cfg := &config.Config{
		Display:    config.DisplayConfig{Type: config.DisplayTM1637},
		Brightness: brightnessCfg(),
		Widgets: []config.WidgetConfig{
			{Type: "clock", Enabled: true, Duration: config.Duration(5 * time.Second)},
		},
	}
	e := New(spy, cfg, nil)
	widgets, _, _ := e.buildWidgets()
	if len(widgets) != 1 {
		t.Fatalf("expected 1 widget, got %d", len(widgets))
	}
	if widgets[0].Name() != "segment-clock" {
		t.Errorf("widget name = %q, want %q", widgets[0].Name(), "segment-clock")
	}
}

func TestBuildWidgets_MessagePixel(t *testing.T) {
	spy := &testutil.SpyDisplay{}
	cfg := &config.Config{
		Display:    config.DisplayConfig{Type: config.DisplayTerminal},
		Brightness: brightnessCfg(),
		Widgets: []config.WidgetConfig{
			{Type: "message", Enabled: true, Text: "Hello"},
		},
	}
	e := New(spy, cfg, nil)
	widgets, _, _ := e.buildWidgets()
	if len(widgets) != 1 {
		t.Fatalf("expected 1 widget, got %d", len(widgets))
	}
	if widgets[0].Name() != "message" {
		t.Errorf("widget name = %q, want %q", widgets[0].Name(), "message")
	}
}

func TestBuildWidgets_AlertPixel(t *testing.T) {
	spy := &testutil.SpyDisplay{}
	cfg := &config.Config{
		Display:    config.DisplayConfig{Type: config.DisplayTerminal},
		Brightness: brightnessCfg(),
		Widgets: []config.WidgetConfig{
			{Type: "alert", Enabled: true},
		},
	}
	e := New(spy, cfg, nil)
	widgets, _, _ := e.buildWidgets()
	if len(widgets) != 1 {
		t.Fatalf("expected 1 widget, got %d", len(widgets))
	}
	if widgets[0].Name() != "alert" {
		t.Errorf("widget name = %q, want %q", widgets[0].Name(), "alert")
	}
}

func TestBuildWidgets_AlertSegment(t *testing.T) {
	spy := &testutil.SpySegmentDisplay{}
	cfg := &config.Config{
		Display:    config.DisplayConfig{Type: config.DisplayTerminalSeg7},
		Brightness: brightnessCfg(),
		Widgets: []config.WidgetConfig{
			{Type: "alert", Enabled: true},
		},
	}
	e := New(spy, cfg, nil)
	widgets, _, _ := e.buildWidgets()
	if len(widgets) != 1 {
		t.Fatalf("expected 1 widget, got %d", len(widgets))
	}
	if widgets[0].Name() != "segment-alert" {
		t.Errorf("widget name = %q, want %q", widgets[0].Name(), "segment-alert")
	}
}

func TestBuildWidgets_AnimationPixelFromRegistry(t *testing.T) {
	spy := &testutil.SpyDisplay{}
	cfg := &config.Config{
		Display:    config.DisplayConfig{Type: config.DisplayTerminal},
		Brightness: brightnessCfg(),
		Widgets: []config.WidgetConfig{
			{Type: "animation", Enabled: true, AnimationType: "bounce"},
		},
	}
	e := New(spy, cfg, nil)
	widgets, _, _ := e.buildWidgets()
	if len(widgets) != 1 {
		t.Fatalf("expected 1 widget, got %d", len(widgets))
	}
	if widgets[0].Name() != "bounce" {
		t.Errorf("widget name = %q, want %q", widgets[0].Name(), "bounce")
	}
}

func TestBuildWidgets_AnimationFrames(t *testing.T) {
	spy := &testutil.SpyDisplay{}
	cfg := &config.Config{
		Display:    config.DisplayConfig{Type: config.DisplayTerminal},
		Brightness: brightnessCfg(),
		Widgets: []config.WidgetConfig{
			{Type: "animation", Enabled: true, AnimationType: "frames"},
		},
	}
	e := New(spy, cfg, nil)
	widgets, _, _ := e.buildWidgets()
	if len(widgets) != 1 {
		t.Fatalf("expected 1 widget, got %d", len(widgets))
	}
}

func TestBuildWidgets_SkipsDisabled(t *testing.T) {
	spy := &testutil.SpyDisplay{}
	cfg := &config.Config{
		Display:    config.DisplayConfig{Type: config.DisplayTerminal},
		Brightness: brightnessCfg(),
		Widgets: []config.WidgetConfig{
			{Type: "clock", Enabled: false},
			{Type: "message", Enabled: true, Text: "Hi"},
		},
	}
	e := New(spy, cfg, nil)
	widgets, _, _ := e.buildWidgets()
	if len(widgets) != 1 {
		t.Fatalf("expected 1 widget (disabled skipped), got %d", len(widgets))
	}
	if widgets[0].Name() != "message" {
		t.Errorf("widget name = %q, want %q", widgets[0].Name(), "message")
	}
}

func TestBuildWidgets_SkipsUnknownType(t *testing.T) {
	spy := &testutil.SpyDisplay{}
	cfg := &config.Config{
		Display:    config.DisplayConfig{Type: config.DisplayTerminal},
		Brightness: brightnessCfg(),
		Widgets: []config.WidgetConfig{
			{Type: "unknown_widget", Enabled: true},
		},
	}
	e := New(spy, cfg, nil)
	widgets, _, _ := e.buildWidgets()
	if len(widgets) != 0 {
		t.Errorf("expected 0 widgets for unknown type, got %d", len(widgets))
	}
}

func TestBuildWidgets_SkipsUnknownAnimationType(t *testing.T) {
	spy := &testutil.SpyDisplay{}
	cfg := &config.Config{
		Display:    config.DisplayConfig{Type: config.DisplayTerminal},
		Brightness: brightnessCfg(),
		Widgets: []config.WidgetConfig{
			{Type: "animation", Enabled: true, AnimationType: "nonexistent"},
		},
	}
	e := New(spy, cfg, nil)
	widgets, _, _ := e.buildWidgets()
	if len(widgets) != 0 {
		t.Errorf("expected 0 widgets for unknown animation type, got %d", len(widgets))
	}
}

func TestBuildWidgets_WithRedis_MessagePixel(t *testing.T) {
	spy := &testutil.SpyDisplay{}
	cfg := &config.Config{
		Display:    config.DisplayConfig{Type: config.DisplayTerminal},
		Brightness: brightnessCfg(),
		Widgets: []config.WidgetConfig{
			{Type: "message", Enabled: true, Text: "fallback", DynamicSource: "mykey"},
		},
	}
	e := New(spy, cfg, nil)
	e.rds = &mockRedis{} // inject mock Redis
	widgets, _, _ := e.buildWidgets()
	if len(widgets) != 1 {
		t.Fatalf("expected 1 widget, got %d", len(widgets))
	}
	if widgets[0].Name() != "redis-message" {
		t.Errorf("widget name = %q, want %q", widgets[0].Name(), "redis-message")
	}
}

func TestBuildWidgets_WithRedis_AlertPixel(t *testing.T) {
	spy := &testutil.SpyDisplay{}
	cfg := &config.Config{
		Display:    config.DisplayConfig{Type: config.DisplayTerminal},
		Brightness: brightnessCfg(),
		Widgets: []config.WidgetConfig{
			{Type: "alert", Enabled: true},
		},
	}
	e := New(spy, cfg, nil)
	e.rds = &mockRedis{}
	widgets, _, _ := e.buildWidgets()
	if len(widgets) != 1 {
		t.Fatalf("expected 1 widget, got %d", len(widgets))
	}
	if widgets[0].Name() != "redis-alert" {
		t.Errorf("widget name = %q, want %q", widgets[0].Name(), "redis-alert")
	}
}

func TestBuildWidgets_MessageSegment(t *testing.T) {
	spy := &testutil.SpySegmentDisplay{}
	cfg := &config.Config{
		Display:    config.DisplayConfig{Type: config.DisplayTerminalSeg14},
		Brightness: brightnessCfg(),
		Widgets: []config.WidgetConfig{
			{Type: "message", Enabled: true, Text: "Hi"},
		},
	}
	e := New(spy, cfg, nil)
	widgets, _, _ := e.buildWidgets()
	if len(widgets) != 1 {
		t.Fatalf("expected 1 widget, got %d", len(widgets))
	}
	if widgets[0].Name() != "segment-message" {
		t.Errorf("widget name = %q, want %q", widgets[0].Name(), "segment-message")
	}
}

func TestBuildWidgets_CronPassedThrough(t *testing.T) {
	spy := &testutil.SpyDisplay{}
	cfg := &config.Config{
		Display:    config.DisplayConfig{Type: config.DisplayTerminal},
		Brightness: brightnessCfg(),
		Widgets: []config.WidgetConfig{
			{Type: "clock", Enabled: true, Cron: "0 9 * * *"},
		},
	}
	e := New(spy, cfg, nil)
	_, _, crons := e.buildWidgets()
	if len(crons) != 1 || crons[0] != "0 9 * * *" {
		t.Errorf("cron = %v, want [\"0 9 * * *\"]", crons)
	}
}

func TestSegmentEncoder_TM1637(t *testing.T) {
	spy := &testutil.SpySegmentDisplay{}
	cfg := &config.Config{
		Display:    config.DisplayConfig{Type: config.DisplayTM1637},
		Brightness: brightnessCfg(),
	}
	e := New(spy, cfg, nil)
	enc := e.segmentEncoder()
	// Enc7 maps '1' to 0x06 (segments b,c)
	if enc('1') != 0x06 {
		t.Errorf("Enc7('1') = 0x%04X, want 0x0006", enc('1'))
	}
}

func TestSegmentEncoder_HT16K33(t *testing.T) {
	spy := &testutil.SpySegmentDisplay{}
	cfg := &config.Config{
		Display:    config.DisplayConfig{Type: config.DisplayHT16K33},
		Brightness: brightnessCfg(),
	}
	e := New(spy, cfg, nil)
	enc := e.segmentEncoder()
	// Enc14 maps 'A' to 0x00F7 which differs from Enc7's 0x77
	val := enc('A')
	if val == uint16(0x77) {
		t.Error("expected Enc14 to differ from Enc7 for 'A'")
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

func TestBuildWidgets_SegmentAnimationFromRegistry(t *testing.T) {
	spy := &testutil.SpySegmentDisplay{}
	cfg := &config.Config{
		Display:    config.DisplayConfig{Type: config.DisplayTerminalSeg7},
		Brightness: brightnessCfg(),
		Widgets: []config.WidgetConfig{
			{Type: "animation", Enabled: true, AnimationType: "rain"},
		},
	}
	e := New(spy, cfg, nil)
	widgets, _, _ := e.buildWidgets()
	if len(widgets) != 1 {
		t.Fatalf("expected 1 widget, got %d", len(widgets))
	}
}

func TestBuildWidgets_SegmentAnimationFrames(t *testing.T) {
	spy := &testutil.SpySegmentDisplay{}
	cfg := &config.Config{
		Display:    config.DisplayConfig{Type: config.DisplayTerminalSeg7},
		Brightness: brightnessCfg(),
		Widgets: []config.WidgetConfig{
			{Type: "animation", Enabled: true, AnimationType: "frames"},
		},
	}
	e := New(spy, cfg, nil)
	widgets, _, _ := e.buildWidgets()
	if len(widgets) != 1 {
		t.Fatalf("expected 1 widget, got %d", len(widgets))
	}
}

func TestBuildWidgets_SegmentUnknownAnimation(t *testing.T) {
	spy := &testutil.SpySegmentDisplay{}
	cfg := &config.Config{
		Display:    config.DisplayConfig{Type: config.DisplayTerminalSeg7},
		Brightness: brightnessCfg(),
		Widgets: []config.WidgetConfig{
			{Type: "animation", Enabled: true, AnimationType: "nonexistent"},
		},
	}
	e := New(spy, cfg, nil)
	widgets, _, _ := e.buildWidgets()
	if len(widgets) != 0 {
		t.Errorf("expected 0 widgets for unknown segment animation, got %d", len(widgets))
	}
}

func TestBuildWidgets_WithRedis_MessageSegment(t *testing.T) {
	spy := &testutil.SpySegmentDisplay{}
	cfg := &config.Config{
		Display:    config.DisplayConfig{Type: config.DisplayTerminalSeg14},
		Brightness: brightnessCfg(),
		Widgets: []config.WidgetConfig{
			{Type: "message", Enabled: true, Text: "fallback", DynamicSource: "key"},
		},
	}
	e := New(spy, cfg, nil)
	e.rds = &mockRedis{}
	widgets, _, _ := e.buildWidgets()
	if len(widgets) != 1 {
		t.Fatalf("expected 1 widget, got %d", len(widgets))
	}
	if widgets[0].Name() != "segment-redis-message" {
		t.Errorf("widget name = %q, want %q", widgets[0].Name(), "segment-redis-message")
	}
}

func TestBuildWidgets_WithRedis_AlertSegment(t *testing.T) {
	spy := &testutil.SpySegmentDisplay{}
	cfg := &config.Config{
		Display:    config.DisplayConfig{Type: config.DisplayTerminalSeg7},
		Brightness: brightnessCfg(),
		Widgets: []config.WidgetConfig{
			{Type: "alert", Enabled: true},
		},
	}
	e := New(spy, cfg, nil)
	e.rds = &mockRedis{}
	widgets, _, _ := e.buildWidgets()
	if len(widgets) != 1 {
		t.Fatalf("expected 1 widget, got %d", len(widgets))
	}
	if widgets[0].Name() != "segment-redis-alert" {
		t.Errorf("widget name = %q, want %q", widgets[0].Name(), "segment-redis-alert")
	}
}

func TestBuildWidgets_ClockFormat24hExplicitFalse(t *testing.T) {
	spy := &testutil.SpyDisplay{}
	f := false
	cfg := &config.Config{
		Display:    config.DisplayConfig{Type: config.DisplayTerminal},
		Brightness: brightnessCfg(),
		Widgets: []config.WidgetConfig{
			{Type: "clock", Enabled: true, Format24h: &f},
		},
	}
	e := New(spy, cfg, nil)
	widgets, _, _ := e.buildWidgets()
	if len(widgets) != 1 {
		t.Fatalf("expected 1 widget, got %d", len(widgets))
	}
}

func TestEngine_Run_WithRedis(t *testing.T) {
	spy := &testutil.SpyDisplay{}
	cfg := &config.Config{
		Brightness: brightnessCfg(),
		Widgets: []config.WidgetConfig{
			{Type: "message", Enabled: true, Duration: config.Duration(50 * time.Millisecond), Text: "Hi"},
		},
	}
	e := New(spy, cfg, nil)
	e.rds = &mockRedis{}
	e.nowFunc = func() time.Time {
		return time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err := e.Run(ctx)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestEngine_Run_SegmentWidget(t *testing.T) {
	spy := &testutil.SpySegmentDisplay{}
	cfg := &config.Config{
		Display:    config.DisplayConfig{Type: config.DisplayTerminalSeg7},
		Brightness: brightnessCfg(),
		Widgets: []config.WidgetConfig{
			{Type: "message", Enabled: true, Duration: config.Duration(50 * time.Millisecond), Text: "Hi"},
		},
	}
	e := New(spy, cfg, nil)
	e.nowFunc = func() time.Time {
		return time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err := e.Run(ctx)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if len(spy.Calls) == 0 {
		t.Error("expected segment writes for segment display")
	}
}

func TestEngine_Run_WidgetWithTimeout(t *testing.T) {
	spy := &testutil.SpyDisplay{}
	cfg := &config.Config{
		Brightness: brightnessCfg(),
		Widgets: []config.WidgetConfig{
			{Type: "clock", Enabled: true, Duration: config.Duration(100 * time.Millisecond)},
		},
	}
	e := New(spy, cfg, nil)
	e.nowFunc = func() time.Time {
		return time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	err := e.Run(ctx)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

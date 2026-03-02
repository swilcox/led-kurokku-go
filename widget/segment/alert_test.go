package segment_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/swilcox/led-kurokku-go/config"
	"github.com/swilcox/led-kurokku-go/display/testutil"
	"github.com/swilcox/led-kurokku-go/segfont"
	"github.com/swilcox/led-kurokku-go/widget/segment"
)

type mockAlertFetcher struct {
	alerts  []config.AlertConfig
	err     error
	deleted []string
}

func (m *mockAlertFetcher) FetchAlerts(_ context.Context) ([]config.AlertConfig, error) {
	return m.alerts, m.err
}

func (m *mockAlertFetcher) DeleteAlert(_ context.Context, id string) error {
	m.deleted = append(m.deleted, id)
	return nil
}

type mockMessageFetcher struct {
	text string
	ok   bool
	err  error
}

func (m *mockMessageFetcher) FetchMessageText(_ context.Context, _ string) (string, bool, error) {
	return m.text, m.ok, m.err
}

func TestSegmentAlert_Empty(t *testing.T) {
	spy := &testutil.SpySegmentDisplay{}
	a := &segment.Alert{Encoder: segfont.Enc7}
	err := a.Run(context.Background(), spy)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if len(spy.Calls) != 0 {
		t.Errorf("expected no calls for empty alerts, got %d", len(spy.Calls))
	}
}

func TestSegmentAlert_DisplaysInPriorityOrder(t *testing.T) {
	spy := &testutil.SpySegmentDisplay{}
	alerts := []config.AlertConfig{
		{ID: "low", Message: "LOW", Priority: 5, DisplayDuration: config.Duration(50 * time.Millisecond)},
		{ID: "high", Message: "HI", Priority: 1, DisplayDuration: config.Duration(50 * time.Millisecond)},
	}

	a := &segment.Alert{
		Alerts:      alerts,
		ScrollSpeed: time.Millisecond,
		Encoder:     segfont.Enc7,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	a.Run(ctx, spy)

	if len(spy.Calls) == 0 {
		t.Fatal("expected segment writes")
	}
}

func TestSegmentAlert_DeleteAfterDisplay(t *testing.T) {
	spy := &testutil.SpySegmentDisplay{}
	alerts := []config.AlertConfig{
		{
			ID:                 "del",
			Message:            "DEL",
			Priority:           1,
			DisplayDuration:    config.Duration(20 * time.Millisecond),
			DeleteAfterDisplay: true,
		},
	}

	a := &segment.Alert{
		Alerts:      alerts,
		ScrollSpeed: time.Millisecond,
		Encoder:     segfont.Enc7,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	a.Run(ctx, spy)

	if len(a.Alerts) != 0 {
		t.Errorf("expected alert to be deleted after display, got %d remaining", len(a.Alerts))
	}
}

func TestSegmentAlert_Name(t *testing.T) {
	a := &segment.Alert{}
	if a.Name() != "segment-alert" {
		t.Errorf("Name() = %q, want %q", a.Name(), "segment-alert")
	}
}

func TestSegmentAlert_OnDelete_Called(t *testing.T) {
	spy := &testutil.SpySegmentDisplay{}
	var deleted []string
	alerts := []config.AlertConfig{
		{
			ID:                 "x1",
			Message:            "Hi",
			Priority:           1,
			DisplayDuration:    config.Duration(20 * time.Millisecond),
			DeleteAfterDisplay: true,
		},
	}

	a := &segment.Alert{
		Alerts:      alerts,
		ScrollSpeed: time.Millisecond,
		Encoder:     segfont.Enc7,
		OnDelete: func(_ context.Context, id string) {
			deleted = append(deleted, id)
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	a.Run(ctx, spy)

	if len(deleted) != 1 || deleted[0] != "x1" {
		t.Errorf("expected OnDelete called with 'x1', got %v", deleted)
	}
}

func TestSegmentRedisAlert_Name(t *testing.T) {
	ra := &segment.RedisAlert{}
	if ra.Name() != "segment-redis-alert" {
		t.Errorf("Name() = %q, want %q", ra.Name(), "segment-redis-alert")
	}
}

func TestSegmentRedisAlert_FetchesAndDisplays(t *testing.T) {
	spy := &testutil.SpySegmentDisplay{}
	fetcher := &mockAlertFetcher{
		alerts: []config.AlertConfig{
			{ID: "r1", Message: "Hi", Priority: 1, DisplayDuration: config.Duration(20 * time.Millisecond)},
		},
	}

	ra := &segment.RedisAlert{
		Fetcher:     fetcher,
		ScrollSpeed: time.Millisecond,
		Encoder:     segfont.Enc7,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	ra.Run(ctx, spy)

	if len(spy.Calls) == 0 {
		t.Error("expected segment writes from redis alert")
	}
}

func TestSegmentRedisAlert_FallsBackOnError(t *testing.T) {
	spy := &testutil.SpySegmentDisplay{}
	fetcher := &mockAlertFetcher{err: errors.New("redis down")}

	ra := &segment.RedisAlert{
		Fetcher: fetcher,
		Fallback: []config.AlertConfig{
			{ID: "fb", Message: "FB", Priority: 1, DisplayDuration: config.Duration(20 * time.Millisecond)},
		},
		ScrollSpeed: time.Millisecond,
		Encoder:     segfont.Enc7,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	ra.Run(ctx, spy)

	if len(spy.Calls) == 0 {
		t.Error("expected segment writes from fallback alerts")
	}
}

func TestSegmentRedisMessage_Name(t *testing.T) {
	rm := &segment.RedisMessage{}
	if rm.Name() != "segment-redis-message" {
		t.Errorf("Name() = %q, want %q", rm.Name(), "segment-redis-message")
	}
}

func TestSegmentRedisMessage_UsesOverride(t *testing.T) {
	spy := &testutil.SpySegmentDisplay{}
	fetcher := &mockMessageFetcher{text: "OK", ok: true}

	rm := &segment.RedisMessage{
		Fetcher:      fetcher,
		Key:          "mykey",
		FallbackText: "FB",
		Encoder:      segfont.Enc7,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	rm.Run(ctx, spy)

	if len(spy.Calls) == 0 {
		t.Error("expected segment writes")
	}
}

func TestSegmentRedisMessage_UsesFallbackOnError(t *testing.T) {
	spy := &testutil.SpySegmentDisplay{}
	fetcher := &mockMessageFetcher{err: errors.New("redis down")}

	rm := &segment.RedisMessage{
		Fetcher:      fetcher,
		Key:          "mykey",
		FallbackText: "FB",
		Encoder:      segfont.Enc7,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	rm.Run(ctx, spy)

	if len(spy.Calls) == 0 {
		t.Error("expected segment writes from fallback")
	}
}

func TestSegmentFrameAnimation_Name(t *testing.T) {
	a := &segment.FrameAnimation{}
	if a.Name() != "segment-animation" {
		t.Errorf("Name() = %q, want %q", a.Name(), "segment-animation")
	}
}

func TestSegmentFrameAnimation_Empty(t *testing.T) {
	spy := &testutil.SpySegmentDisplay{}
	a := &segment.FrameAnimation{}
	err := a.Run(context.Background(), spy)
	if err != nil {
		t.Errorf("expected nil for empty frames, got %v", err)
	}
	if len(spy.Calls) != 0 {
		t.Errorf("expected no calls for empty animation, got %d", len(spy.Calls))
	}
}

func TestSegmentFrameAnimation_PlaysFrames(t *testing.T) {
	spy := &testutil.SpySegmentDisplay{}
	frames := []config.SegmentFrameConfig{
		{Data: []uint16{0x01, 0x02, 0x03, 0x04}, Colon: true, Duration: config.Duration(time.Millisecond)},
		{Data: []uint16{0x10, 0x20, 0x30, 0x40}, Colon: false, Duration: config.Duration(time.Millisecond)},
	}

	a := &segment.FrameAnimation{Frames: frames}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	a.Run(ctx, spy)

	if len(spy.Calls) < 2 {
		t.Fatalf("expected at least 2 calls, got %d", len(spy.Calls))
	}
	if !spy.Calls[0].Colon {
		t.Error("expected first frame colon=true")
	}
	if spy.Calls[1].Colon {
		t.Error("expected second frame colon=false")
	}
}

func TestSegmentFrameAnimation_FallbackDuration(t *testing.T) {
	spy := &testutil.SpySegmentDisplay{}
	frames := []config.SegmentFrameConfig{
		{Data: []uint16{0x01, 0x02, 0x03, 0x04}}, // no per-frame duration
	}

	a := &segment.FrameAnimation{
		Frames:        frames,
		FrameDuration: time.Millisecond,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	a.Run(ctx, spy)

	if len(spy.Calls) == 0 {
		t.Error("expected at least one frame with fallback duration")
	}
}

func TestSegmentMessage_Name(t *testing.T) {
	m := &segment.Message{}
	if m.Name() != "segment-message" {
		t.Errorf("Name() = %q, want %q", m.Name(), "segment-message")
	}
}

func TestSegmentMessage_DefaultEncoder(t *testing.T) {
	spy := &testutil.SpySegmentDisplay{}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	msg := &segment.Message{
		Text: "Hi",
		// No encoder set — should default to Enc7
	}

	msg.Run(ctx, spy)

	if len(spy.Calls) == 0 {
		t.Fatal("expected writes with default encoder")
	}
}

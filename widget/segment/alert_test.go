package segment_test

import (
	"context"
	"testing"
	"time"

	"github.com/swilcox/led-kurokku-go/config"
	"github.com/swilcox/led-kurokku-go/display/testutil"
	"github.com/swilcox/led-kurokku-go/segfont"
	"github.com/swilcox/led-kurokku-go/widget/segment"
)

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

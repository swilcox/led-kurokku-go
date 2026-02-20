package widget_test

import (
	"context"
	"testing"
	"time"

	"github.com/swilcox/led-kurokku-go/config"
	"github.com/swilcox/led-kurokku-go/display/testutil"
	"github.com/swilcox/led-kurokku-go/widget"
)

func TestAlert_NoAlerts_ReturnsImmediately(t *testing.T) {
	spy := &testutil.SpyDisplay{}
	a := &widget.Alert{}
	a.Run(context.Background(), spy) //nolint:errcheck
	if len(spy.Frames) != 0 {
		t.Errorf("expected no frames for empty alert list, got %d", len(spy.Frames))
	}
}

func TestAlert_OnDelete_CalledForDeleteAfterDisplay(t *testing.T) {
	var deleted []string
	spy := &testutil.SpyDisplay{}

	a := &widget.Alert{
		Alerts: []config.AlertConfig{
			{
				ID:                 "alert1",
				Message:            "Hi",
				Priority:           1,
				DisplayDuration:    config.Duration(time.Millisecond),
				DeleteAfterDisplay: true,
			},
		},
		OnDelete: func(_ context.Context, id string) {
			deleted = append(deleted, id)
		},
	}

	a.Run(context.Background(), spy) //nolint:errcheck

	if len(deleted) != 1 || deleted[0] != "alert1" {
		t.Errorf("expected OnDelete called with 'alert1', got %v", deleted)
	}
}

func TestAlert_PriorityOrder_ViaOnDelete(t *testing.T) {
	var order []string
	spy := &testutil.SpyDisplay{}

	a := &widget.Alert{
		Alerts: []config.AlertConfig{
			{
				ID:                 "low",
				Message:            "Lo",
				Priority:           10,
				DisplayDuration:    config.Duration(time.Millisecond),
				DeleteAfterDisplay: true,
			},
			{
				ID:                 "high",
				Message:            "Hi",
				Priority:           1,
				DisplayDuration:    config.Duration(time.Millisecond),
				DeleteAfterDisplay: true,
			},
		},
		OnDelete: func(_ context.Context, id string) {
			order = append(order, id)
		},
		// minute 0 matches */10 * * * *, so the priority-10 alert is not throttled
		NowFunc: func() time.Time {
			return time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
		},
	}

	a.Run(context.Background(), spy) //nolint:errcheck

	if len(order) != 2 {
		t.Fatalf("expected 2 deletes, got %d", len(order))
	}
	if order[0] != "high" {
		t.Errorf("expected 'high' (priority 1) first, got %q", order[0])
	}
	if order[1] != "low" {
		t.Errorf("expected 'low' (priority 10) second, got %q", order[1])
	}
}

func TestAlert_Priority10_SkippedWhenCronNotMatch(t *testing.T) {
	spy := &testutil.SpyDisplay{}

	a := &widget.Alert{
		Alerts: []config.AlertConfig{
			{
				ID:              "p10",
				Message:         "Low-priority",
				Priority:        10,
				DisplayDuration: config.Duration(time.Millisecond),
			},
		},
		// minute 5 does NOT match */10 * * * *
		NowFunc: func() time.Time {
			return time.Date(2024, 1, 1, 12, 5, 0, 0, time.UTC)
		},
	}

	a.Run(context.Background(), spy) //nolint:errcheck

	if len(spy.Frames) != 0 {
		t.Errorf("expected no frames for throttled priority-10 alert, got %d", len(spy.Frames))
	}
}

func TestAlert_Priority10_DisplayedWhenCronMatches(t *testing.T) {
	spy := &testutil.SpyDisplay{}

	a := &widget.Alert{
		Alerts: []config.AlertConfig{
			{
				ID:              "p10",
				Message:         "Low-priority",
				Priority:        10,
				DisplayDuration: config.Duration(time.Millisecond),
			},
		},
		// minute 0 matches */10 * * * *
		NowFunc: func() time.Time {
			return time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
		},
	}

	a.Run(context.Background(), spy) //nolint:errcheck

	if len(spy.Frames) == 0 {
		t.Error("expected frames for priority-10 alert at matching cron minute, got none")
	}
}

func TestAlert_Priority1_AlwaysDisplayed(t *testing.T) {
	spy := &testutil.SpyDisplay{}

	a := &widget.Alert{
		Alerts: []config.AlertConfig{
			{
				ID:              "p1",
				Message:         "Urgent",
				Priority:        1,
				DisplayDuration: config.Duration(time.Millisecond),
			},
		},
		// minute 7 — not a */10 boundary, but priority 1 is never throttled
		NowFunc: func() time.Time {
			return time.Date(2024, 1, 1, 12, 7, 0, 0, time.UTC)
		},
	}

	a.Run(context.Background(), spy) //nolint:errcheck

	if len(spy.Frames) == 0 {
		t.Error("expected frames for priority-1 alert regardless of time, got none")
	}
}

func TestAlert_OnDelete_NotCalledWithoutFlag(t *testing.T) {
	var deleted []string
	spy := &testutil.SpyDisplay{}

	a := &widget.Alert{
		Alerts: []config.AlertConfig{
			{
				ID:                 "no-delete",
				Message:            "Hi",
				Priority:           1,
				DisplayDuration:    config.Duration(time.Millisecond),
				DeleteAfterDisplay: false,
			},
		},
		OnDelete: func(_ context.Context, id string) {
			deleted = append(deleted, id)
		},
	}

	a.Run(context.Background(), spy) //nolint:errcheck

	if len(deleted) != 0 {
		t.Errorf("expected OnDelete not called when DeleteAfterDisplay=false, got %v", deleted)
	}
}

package widget_test

import (
	"context"
	"testing"
	"time"

	"github.com/swilcox/led-kurokku-go/display/testutil"
	"github.com/swilcox/led-kurokku-go/widget"
)

func TestClock_24h_WritesFrames(t *testing.T) {
	fixed := time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC)
	spy := &testutil.SpyDisplay{}
	ctx, cancel := context.WithCancel(context.Background())

	clk := &widget.Clock{
		Format24h: true,
		NowFunc:   func() time.Time { return fixed },
	}

	go func() {
		// Allow at least one frame to be written before cancelling.
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	clk.Run(ctx, spy) //nolint:errcheck

	if len(spy.Frames) == 0 {
		t.Error("expected at least one frame to be written")
	}
}

func TestClock_12h_AM_WritesFrames(t *testing.T) {
	fixed := time.Date(2024, 1, 15, 9, 5, 0, 0, time.UTC) // 9:05 AM
	spy := &testutil.SpyDisplay{}
	ctx, cancel := context.WithCancel(context.Background())

	clk := &widget.Clock{
		Format24h: false,
		NowFunc:   func() time.Time { return fixed },
	}

	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	clk.Run(ctx, spy) //nolint:errcheck

	if len(spy.Frames) == 0 {
		t.Error("expected at least one frame to be written")
	}
}

func TestClock_12h_PM_WritesFrames(t *testing.T) {
	fixed := time.Date(2024, 1, 15, 15, 45, 0, 0, time.UTC) // 3:45 PM
	spy := &testutil.SpyDisplay{}
	ctx, cancel := context.WithCancel(context.Background())

	clk := &widget.Clock{
		Format24h: false,
		NowFunc:   func() time.Time { return fixed },
	}

	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	clk.Run(ctx, spy) //nolint:errcheck

	if len(spy.Frames) == 0 {
		t.Error("expected at least one frame to be written for PM time")
	}
}

func TestClock_NowFuncUsed(t *testing.T) {
	callCount := 0
	spy := &testutil.SpyDisplay{}
	ctx, cancel := context.WithCancel(context.Background())

	clk := &widget.Clock{
		Format24h: true,
		NowFunc: func() time.Time {
			callCount++
			return time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
		},
	}

	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	clk.Run(ctx, spy) //nolint:errcheck

	if callCount == 0 {
		t.Error("expected NowFunc to be called at least once")
	}
}

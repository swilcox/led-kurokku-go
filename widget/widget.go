package widget

import (
	"context"
	"time"

	"github.com/swilcox/led-kurokku-go/display"
	"github.com/swilcox/led-kurokku-go/font"
	"github.com/swilcox/led-kurokku-go/framebuf"
)

// Widget is the interface all display widgets implement.
type Widget interface {
	Name() string
	Run(ctx context.Context, disp display.Display) error
}

// SleepOrCancel sleeps for d or returns early if ctx is cancelled.
func SleepOrCancel(ctx context.Context, d time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(d):
		return nil
	}
}

// ScrollText scrolls text across the display. If repeats <= 0, it scrolls
// until the context is cancelled.
func ScrollText(ctx context.Context, disp display.Display, text string,
	scrollSpeed time.Duration, repeats int, sleepBetween time.Duration) error {

	cols := font.RenderText(text)
	totalWidth := len(cols)
	dispWidth := disp.Width()

	// Pad so text scrolls fully on and off
	padded := make([]byte, dispWidth+totalWidth+dispWidth)
	copy(padded[dispWidth:], cols)

	count := 0
	for {
		for offset := 0; offset <= len(padded)-dispWidth; offset++ {
			var f framebuf.Frame
			copy(f[:], padded[offset:offset+dispWidth])
			disp.WriteFramebuffer(f.Bytes())
			if err := SleepOrCancel(ctx, scrollSpeed); err != nil {
				return err
			}
		}
		count++
		if repeats > 0 && count >= repeats {
			return nil
		}
		if err := SleepOrCancel(ctx, sleepBetween); err != nil {
			return err
		}
	}
}

package widget

import (
	"context"
	"fmt"
	"time"

	"github.com/swilcox/led-kurokku-go/display"
	"github.com/swilcox/led-kurokku-go/framebuf"
)

// Clock displays the current time with a blinking colon.
type Clock struct {
	Format24h bool
	NowFunc   func() time.Time
}

func (c *Clock) now() time.Time {
	if c.NowFunc != nil {
		return c.NowFunc()
	}
	return time.Now()
}

func (c *Clock) Name() string { return "clock" }

func (c *Clock) Run(ctx context.Context, disp display.Display) error {
	for {
		now := c.now()
		hour := now.Hour()
		minute := now.Minute()
		isPM := false

		if !c.Format24h {
			isPM = hour >= 12
			hour = hour % 12
			if hour == 0 {
				hour = 12
			}
		}

		colonOn := fmt.Sprintf("%d:%02d", hour, minute)
		colonOff := fmt.Sprintf("%d %02d", hour, minute)
		if c.Format24h || hour >= 10 {
			colonOn = fmt.Sprintf("%02d:%02d", hour, minute)
			colonOff = fmt.Sprintf("%02d %02d", hour, minute)
		}

		if !c.Format24h && isPM {
			// PM double blink: 150ms on, 200ms off, 150ms on, 500ms off
			if err := c.showText(ctx, disp, colonOn, 150*time.Millisecond); err != nil {
				return err
			}
			if err := c.showText(ctx, disp, colonOff, 200*time.Millisecond); err != nil {
				return err
			}
			if err := c.showText(ctx, disp, colonOn, 150*time.Millisecond); err != nil {
				return err
			}
			if err := c.showText(ctx, disp, colonOff, 500*time.Millisecond); err != nil {
				return err
			}
		} else {
			// 24h or AM: 500ms on, 500ms off
			if err := c.showText(ctx, disp, colonOn, 500*time.Millisecond); err != nil {
				return err
			}
			if err := c.showText(ctx, disp, colonOff, 500*time.Millisecond); err != nil {
				return err
			}
		}
	}
}

func (c *Clock) showText(ctx context.Context, disp display.Display, text string, d time.Duration) error {
	var f framebuf.Frame
	w := framebuf.BlitText(&f, text, 0)
	// Center the text
	if w < 32 {
		var centered framebuf.Frame
		offset := (32 - w) / 2
		framebuf.BlitText(&centered, text, offset)
		disp.WriteFramebuffer(centered.Bytes())
	} else {
		disp.WriteFramebuffer(f.Bytes())
	}
	return SleepOrCancel(ctx, d)
}

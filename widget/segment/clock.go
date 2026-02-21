package segment

import (
	"context"
	"fmt"
	"time"

	"github.com/swilcox/led-kurokku-go/display"
	"github.com/swilcox/led-kurokku-go/segfont"
	"github.com/swilcox/led-kurokku-go/widget"
)

// Clock displays the current time on a segment display with a blinking colon.
type Clock struct {
	Format24h bool
	NowFunc   func() time.Time
	Encoder   segfont.Encoder
}

func (c *Clock) now() time.Time {
	if c.NowFunc != nil {
		return c.NowFunc()
	}
	return time.Now()
}

func (c *Clock) enc() segfont.Encoder {
	if c.Encoder != nil {
		return c.Encoder
	}
	return segfont.Enc7
}

func (c *Clock) Name() string { return "segment-clock" }

func (c *Clock) Run(ctx context.Context, disp display.Display) error {
	sd := disp.(display.SegmentDisplay)
	enc := c.enc()

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

		var text string
		if !c.Format24h && hour < 10 {
			text = fmt.Sprintf(" %d%02d", hour, minute)
		} else {
			text = fmt.Sprintf("%02d%02d", hour, minute)
		}

		segments := segfont.EncodeText(enc, text)

		if !c.Format24h && isPM {
			// PM double blink: 150ms on, 200ms off, 150ms on, 500ms off
			sd.WriteSegments(segments, true)
			if err := widget.SleepOrCancel(ctx, 150*time.Millisecond); err != nil {
				return err
			}
			sd.WriteSegments(segments, false)
			if err := widget.SleepOrCancel(ctx, 200*time.Millisecond); err != nil {
				return err
			}
			sd.WriteSegments(segments, true)
			if err := widget.SleepOrCancel(ctx, 150*time.Millisecond); err != nil {
				return err
			}
			sd.WriteSegments(segments, false)
			if err := widget.SleepOrCancel(ctx, 500*time.Millisecond); err != nil {
				return err
			}
		} else {
			// 24h or AM: 500ms on, 500ms off
			sd.WriteSegments(segments, true)
			if err := widget.SleepOrCancel(ctx, 500*time.Millisecond); err != nil {
				return err
			}
			sd.WriteSegments(segments, false)
			if err := widget.SleepOrCancel(ctx, 500*time.Millisecond); err != nil {
				return err
			}
		}
	}
}

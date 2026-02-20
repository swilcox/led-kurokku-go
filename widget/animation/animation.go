package animation

import (
	"context"
	"time"

	"github.com/swilcox/led-kurokku-go/config"
	"github.com/swilcox/led-kurokku-go/display"
	"github.com/swilcox/led-kurokku-go/framebuf"
	"github.com/swilcox/led-kurokku-go/widget"
)

// FrameAnimation plays a list of pre-defined frames in a loop.
type FrameAnimation struct {
	Frames        []config.FrameConfig
	FrameDuration time.Duration
}

func (a *FrameAnimation) Name() string { return "animation" }

func (a *FrameAnimation) Run(ctx context.Context, disp display.Display) error {
	if len(a.Frames) == 0 {
		return nil
	}

	for {
		for _, fc := range a.Frames {
			var f framebuf.Frame
			f = fc.Data
			disp.WriteFramebuffer(f.Bytes())

			dur := fc.Duration.Unwrap()
			if dur == 0 {
				dur = a.FrameDuration
			}
			if dur == 0 {
				dur = 100 * time.Millisecond
			}
			if err := widget.SleepOrCancel(ctx, dur); err != nil {
				return err
			}
		}
	}
}

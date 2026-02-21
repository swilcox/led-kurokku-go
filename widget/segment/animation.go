package segment

import (
	"context"
	"time"

	"github.com/swilcox/led-kurokku-go/config"
	"github.com/swilcox/led-kurokku-go/display"
	"github.com/swilcox/led-kurokku-go/widget"
)

// FrameAnimation plays a list of pre-defined segment frames in a loop.
type FrameAnimation struct {
	Frames        []config.SegmentFrameConfig
	FrameDuration time.Duration
}

func (a *FrameAnimation) Name() string { return "segment-animation" }

func (a *FrameAnimation) Run(ctx context.Context, disp display.Display) error {
	sd := disp.(display.SegmentDisplay)

	if len(a.Frames) == 0 {
		return nil
	}

	for {
		for _, fc := range a.Frames {
			sd.WriteSegments(fc.Data, fc.Colon)

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

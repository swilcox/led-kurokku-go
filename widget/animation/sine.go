package animation

import (
	"context"
	"math"
	"time"

	"github.com/swilcox/led-kurokku-go/display"
	"github.com/swilcox/led-kurokku-go/framebuf"
)

// Sine displays a scrolling sine wave across the display.
type Sine struct{}

func (s *Sine) Name() string { return "sine" }

func (s *Sine) Run(ctx context.Context, disp display.Display) error {
	var phase float64

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}

		var f framebuf.Frame
		for x := 0; x < 32; x++ {
			// Scale to full 8-row height (0–7).
			y := int(math.Round(3.5 + 3.5*math.Sin(phase+float64(x)*0.35)))
			f.SetPixel(x, y, true)
		}
		disp.WriteFramebuffer(f.Bytes())
		phase += 0.2
	}
}

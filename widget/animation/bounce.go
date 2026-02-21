package animation

import (
	"context"
	"time"

	"github.com/swilcox/led-kurokku-go/display"
	"github.com/swilcox/led-kurokku-go/framebuf"
)

// Bounce simulates a pixel bouncing around the display with a short trail.
type Bounce struct{}

func (b *Bounce) Name() string { return "bounce" }

func (b *Bounce) Run(ctx context.Context, disp display.Display) error {
	x, y := 16.0, 4.0
	dx, dy := 0.7, 0.5

	// Previous two positions for trail.
	px, py := int(x), int(y)
	ppx, ppy := int(x), int(y)

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}

		ppx, ppy = px, py
		px, py = int(x), int(y)

		x += dx
		y += dy

		if x <= 0 {
			x = 0
			dx = -dx
		} else if x >= 31 {
			x = 31
			dx = -dx
		}
		if y <= 0 {
			y = 0
			dy = -dy
		} else if y >= 7 {
			y = 7
			dy = -dy
		}

		var f framebuf.Frame
		f.SetPixel(ppx, ppy, true)
		f.SetPixel(px, py, true)
		f.SetPixel(int(x), int(y), true)
		disp.WriteFramebuffer(f.Bytes())
	}
}

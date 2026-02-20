package animation

import (
	"context"
	"math/rand"
	"time"

	"github.com/swilcox/led-kurokku-go/display"
	"github.com/swilcox/led-kurokku-go/framebuf"
)

// Rain simulates raindrops falling down the display.
type Rain struct{}

func (r *Rain) Name() string { return "rain" }

func (r *Rain) Run(ctx context.Context, disp display.Display) error {
	// Each column has a drop position (-1 = inactive)
	var drops [32]int
	for i := range drops {
		drops[i] = -1
	}

	ticker := time.NewTicker(80 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}

		// Randomly spawn new drops
		for x := 0; x < 32; x++ {
			if drops[x] < 0 && rand.Intn(20) == 0 {
				drops[x] = 0
			}
		}

		var f framebuf.Frame
		for x := 0; x < 32; x++ {
			if drops[x] < 0 {
				continue
			}
			// Draw the drop head and 1-2 trailing pixels
			for t := 0; t < 3; t++ {
				y := drops[x] - t
				if y >= 0 && y < 8 {
					f.SetPixel(x, y, true)
				}
			}
			drops[x]++
			// Reset when the trail is fully off screen
			if drops[x] > 10 {
				drops[x] = -1
			}
		}

		disp.WriteFramebuffer(f.Bytes())
	}
}

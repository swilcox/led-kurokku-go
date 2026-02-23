package animation

import (
	"context"
	"math/rand"
	"time"

	"github.com/swilcox/led-kurokku-go/display"
	"github.com/swilcox/led-kurokku-go/framebuf"
)

// Static displays TV-static random pixel noise.
type Static struct{}

func (r *Static) Name() string { return "static" }

func (r *Static) Run(ctx context.Context, disp display.Display) error {
	pd := disp.(display.PixelDisplay)

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
			f[x] = byte(rand.Intn(256))
		}
		pd.WriteFramebuffer(f.Bytes())
	}
}

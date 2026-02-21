package animation

import (
	"context"
	"time"

	"github.com/swilcox/led-kurokku-go/display"
	"github.com/swilcox/led-kurokku-go/framebuf"
)

// Scanner sweeps a bright column back and forth across the display (KITT-style),
// with a directional trail that fades away from the head.
type Scanner struct{}

func (s *Scanner) Name() string { return "scanner" }

func (s *Scanner) Run(ctx context.Context, disp display.Display) error {
	pos := 0
	dir := 1

	ticker := time.NewTicker(40 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}

		var f framebuf.Frame
		f[pos] = 0xFF // full bright column at head

		// Trail extends opposite to the direction of travel.
		t1 := pos - dir
		t2 := pos - dir*2
		t3 := pos - dir*3
		if t1 >= 0 && t1 < 32 {
			f[t1] = 0xAA // dense: every other pixel
		}
		if t2 >= 0 && t2 < 32 {
			f[t2] = 0x44 // sparse
		}
		if t3 >= 0 && t3 < 32 {
			f[t3] = 0x11 // very sparse
		}

		disp.WriteFramebuffer(f.Bytes())

		pos += dir
		if pos >= 31 {
			pos = 31
			dir = -1
		} else if pos <= 0 {
			pos = 0
			dir = 1
		}
	}
}

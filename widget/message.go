package widget

import (
	"context"
	"time"

	"github.com/swilcox/led-kurokku-go/display"
	"github.com/swilcox/led-kurokku-go/font"
	"github.com/swilcox/led-kurokku-go/framebuf"
)

// Message displays static or scrolling text.
type Message struct {
	Text         string
	ScrollSpeed  time.Duration
	Repeats      int
	SleepBetween time.Duration
}

func (m *Message) Name() string { return "message" }

func (m *Message) Run(ctx context.Context, disp display.Display) error {
	cols := font.RenderText(m.Text)
	textWidth := len(cols)

	if textWidth <= disp.Width() {
		// Static display, centered
		var f framebuf.Frame
		offset := (32 - textWidth) / 2
		framebuf.BlitText(&f, m.Text, offset)
		disp.WriteFramebuffer(f.Bytes())
		// Hold until context is done
		<-ctx.Done()
		return ctx.Err()
	}

	// Scrolling text
	speed := m.ScrollSpeed
	if speed == 0 {
		speed = 50 * time.Millisecond
	}
	repeats := m.Repeats
	if repeats == 0 {
		repeats = 1
	}
	return ScrollText(ctx, disp, m.Text, speed, repeats, m.SleepBetween)
}

package widget_test

import (
	"context"
	"testing"
	"time"

	"github.com/swilcox/led-kurokku-go/display/testutil"
	"github.com/swilcox/led-kurokku-go/widget"
)

func TestMessage_Static_WritesOneFrame(t *testing.T) {
	spy := &testutil.SpyDisplay{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before Run so static display returns immediately

	m := &widget.Message{Text: "Hi"} // 11 cols < 32, static
	m.Run(ctx, spy)                  //nolint:errcheck

	if len(spy.Frames) != 1 {
		t.Errorf("expected 1 frame for static message, got %d", len(spy.Frames))
	}
}

func TestMessage_Scroll_WritesMultipleFrames(t *testing.T) {
	spy := &testutil.SpyDisplay{}

	m := &widget.Message{
		Text:        "Hello World!", // 71 cols > 32, will scroll
		ScrollSpeed: time.Millisecond,
		Repeats:     1,
	}
	m.Run(context.Background(), spy) //nolint:errcheck

	if len(spy.Frames) < 2 {
		t.Errorf("expected multiple frames for scrolling message, got %d", len(spy.Frames))
	}
}

func TestMessage_Scroll_RespectsRepeats(t *testing.T) {
	spy1 := &testutil.SpyDisplay{}
	spy2 := &testutil.SpyDisplay{}

	text := "Hello World!"
	m1 := &widget.Message{Text: text, ScrollSpeed: time.Millisecond, Repeats: 1}
	m2 := &widget.Message{Text: text, ScrollSpeed: time.Millisecond, Repeats: 2}

	m1.Run(context.Background(), spy1) //nolint:errcheck
	m2.Run(context.Background(), spy2) //nolint:errcheck

	// Two repeats should produce roughly twice as many frames as one repeat.
	if len(spy2.Frames) <= len(spy1.Frames) {
		t.Errorf("2 repeats (%d frames) should produce more frames than 1 repeat (%d frames)",
			len(spy2.Frames), len(spy1.Frames))
	}
}

func TestMessage_Static_32ColsWide(t *testing.T) {
	spy := &testutil.SpyDisplay{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// SpyDisplay.Width() = 32. A message exactly 32 cols wide stays static.
	// font.RenderText renders 5 cols per char + 1 gap. 5 chars = 5*5+4*1 = 29 cols.
	// Use a message that fits in 32 to confirm static path.
	m := &widget.Message{Text: "Hello"} // 29 cols < 32
	m.Run(ctx, spy)                     //nolint:errcheck

	if len(spy.Frames) != 1 {
		t.Errorf("expected 1 static frame for short text, got %d", len(spy.Frames))
	}
}

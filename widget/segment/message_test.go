package segment_test

import (
	"context"
	"testing"
	"time"

	"github.com/swilcox/led-kurokku-go/display/testutil"
	"github.com/swilcox/led-kurokku-go/segfont"
	"github.com/swilcox/led-kurokku-go/widget/segment"
)

func TestSegmentMessage_Short_Static(t *testing.T) {
	spy := &testutil.SpySegmentDisplay{}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	msg := &segment.Message{
		Text:    "Hi",
		Encoder: segfont.Enc7,
	}

	msg.Run(ctx, spy)

	if len(spy.Calls) == 0 {
		t.Fatal("expected at least one WriteSegments call")
	}
	// "Hi" is 2 chars on a 4-digit display → centered with 1 blank on each side
	call := spy.Calls[0]
	if len(call.Segments) != 4 {
		t.Fatalf("expected 4 segments, got %d", len(call.Segments))
	}
	// First segment should be blank (padding)
	if call.Segments[0] != 0 {
		t.Errorf("expected blank padding at position 0, got 0x%04X", call.Segments[0])
	}
	// Last segment should be blank (padding)
	if call.Segments[3] != 0 {
		t.Errorf("expected blank padding at position 3, got 0x%04X", call.Segments[3])
	}
}

func TestSegmentMessage_Long_Scrolls(t *testing.T) {
	spy := &testutil.SpySegmentDisplay{}
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	msg := &segment.Message{
		Text:        "Hello World",
		ScrollSpeed: time.Millisecond,
		Repeats:     1,
		Encoder:     segfont.Enc7,
	}

	msg.Run(ctx, spy)

	if len(spy.Calls) < 2 {
		t.Errorf("expected multiple scroll steps, got %d", len(spy.Calls))
	}
}

func TestSegmentMessage_DefaultScrollSpeed(t *testing.T) {
	spy := &testutil.SpySegmentDisplay{}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	msg := &segment.Message{
		Text:    "ABCDE", // longer than display
		Repeats: -1,
		Encoder: segfont.Enc7,
	}

	msg.Run(ctx, spy)

	// With 300ms default speed and 50ms timeout, we should get very few calls
	if len(spy.Calls) > 2 {
		t.Errorf("expected few calls at 300ms default speed, got %d", len(spy.Calls))
	}
}

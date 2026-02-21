package segment

import (
	"context"
	"time"

	"github.com/swilcox/led-kurokku-go/display"
	"github.com/swilcox/led-kurokku-go/segfont"
	"github.com/swilcox/led-kurokku-go/widget"
)

// Message displays static or scrolling text on a segment display.
type Message struct {
	Text         string
	ScrollSpeed  time.Duration
	Repeats      int
	SleepBetween time.Duration
	Encoder      segfont.Encoder
}

func (m *Message) enc() segfont.Encoder {
	if m.Encoder != nil {
		return m.Encoder
	}
	return segfont.Enc7
}

func (m *Message) Name() string { return "segment-message" }

func (m *Message) Run(ctx context.Context, disp display.Display) error {
	sd := disp.(display.SegmentDisplay)
	enc := m.enc()
	dispLen := sd.DisplayLength()

	runes := []rune(m.Text)
	encoded := segfont.EncodeText(enc, m.Text)

	if len(runes) <= dispLen {
		// Static display, centered with blank padding
		segments := make([]uint16, dispLen)
		offset := (dispLen - len(encoded)) / 2
		copy(segments[offset:], encoded)
		sd.WriteSegments(segments, false)
		<-ctx.Done()
		return ctx.Err()
	}

	// Scrolling text
	speed := m.ScrollSpeed
	if speed == 0 {
		speed = 300 * time.Millisecond
	}
	repeats := m.Repeats
	if repeats == 0 {
		repeats = 1
	}

	// Pad with blanks on both sides for smooth scroll on/off
	padded := make([]uint16, dispLen+len(encoded)+dispLen)
	copy(padded[dispLen:], encoded)

	count := 0
	for {
		for offset := 0; offset <= len(padded)-dispLen; offset++ {
			sd.WriteSegments(padded[offset:offset+dispLen], false)
			if err := widget.SleepOrCancel(ctx, speed); err != nil {
				return err
			}
		}
		count++
		if repeats > 0 && count >= repeats {
			return nil
		}
		if err := widget.SleepOrCancel(ctx, m.SleepBetween); err != nil {
			return err
		}
	}
}

package segment_test

import (
	"context"
	"testing"
	"time"

	"github.com/swilcox/led-kurokku-go/display/testutil"
	"github.com/swilcox/led-kurokku-go/segfont"
	"github.com/swilcox/led-kurokku-go/widget/segment"
)

func TestSegmentClock_24h_WritesSegments(t *testing.T) {
	fixed := time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC)
	spy := &testutil.SpySegmentDisplay{}
	ctx, cancel := context.WithCancel(context.Background())

	clk := &segment.Clock{
		Format24h: true,
		NowFunc:   func() time.Time { return fixed },
		Encoder:   segfont.Enc7,
	}

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	clk.Run(ctx, spy)

	if len(spy.Calls) == 0 {
		t.Fatal("expected segment writes")
	}
	// First call should have colon on (24h: 500ms on, 500ms off)
	if !spy.Calls[0].Colon {
		t.Error("expected first call to have colon on")
	}
	// Should have 4 segments
	if len(spy.Calls[0].Segments) != 4 {
		t.Errorf("expected 4 segments, got %d", len(spy.Calls[0].Segments))
	}
	// "1430" → digit 0 = '1' (0x06), digit 1 = '4' (0x66)
	if spy.Calls[0].Segments[0] != uint16(segfont.Seg7['1']) {
		t.Errorf("digit 0 = 0x%04X, want 0x%04X", spy.Calls[0].Segments[0], segfont.Seg7['1'])
	}
	if spy.Calls[0].Segments[1] != uint16(segfont.Seg7['4']) {
		t.Errorf("digit 1 = 0x%04X, want 0x%04X", spy.Calls[0].Segments[1], segfont.Seg7['4'])
	}
}

func TestSegmentClock_12h_AM(t *testing.T) {
	// 9:05 AM
	fixed := time.Date(2024, 1, 15, 9, 5, 0, 0, time.UTC)
	spy := &testutil.SpySegmentDisplay{}
	ctx, cancel := context.WithCancel(context.Background())

	clk := &segment.Clock{
		Format24h: false,
		NowFunc:   func() time.Time { return fixed },
		Encoder:   segfont.Enc7,
	}

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	clk.Run(ctx, spy)

	if len(spy.Calls) == 0 {
		t.Fatal("expected segment writes")
	}
	// 9:05 AM → " 905" with leading blank
	// digit 0 should be space (0x00)
	if spy.Calls[0].Segments[0] != 0 {
		t.Errorf("digit 0 = 0x%04X, want 0x0000 (blank for single-digit hour)", spy.Calls[0].Segments[0])
	}
}

func TestSegmentClock_12h_PM_DoubleBlink(t *testing.T) {
	// 2:30 PM = 14:30
	fixed := time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC)
	spy := &testutil.SpySegmentDisplay{}
	ctx, cancel := context.WithCancel(context.Background())

	clk := &segment.Clock{
		Format24h: false,
		NowFunc:   func() time.Time { return fixed },
		Encoder:   segfont.Enc7,
	}

	go func() {
		// PM double blink is ~1000ms total per cycle, run briefly
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	clk.Run(ctx, spy)

	if len(spy.Calls) == 0 {
		t.Fatal("expected segment writes")
	}
	// PM pattern: colon on, off, on, off
	if !spy.Calls[0].Colon {
		t.Error("expected first PM call to have colon on")
	}
}

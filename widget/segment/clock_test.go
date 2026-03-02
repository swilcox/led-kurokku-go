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

func TestSegmentClock_24h_Midnight(t *testing.T) {
	fixed := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
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
	// "0000" → all four digits should be '0' (0x3F in Seg7)
	for i := 0; i < 4; i++ {
		if spy.Calls[0].Segments[i] != uint16(segfont.Seg7['0']) {
			t.Errorf("digit %d = 0x%04X, want 0x%04X", i, spy.Calls[0].Segments[i], segfont.Seg7['0'])
		}
	}
}

func TestSegmentClock_12h_Midnight_Shows12(t *testing.T) {
	// Midnight in 12h = 12:00 AM
	fixed := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
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
	// "1200" → digit 0 = '1', digit 1 = '2'
	if spy.Calls[0].Segments[0] != uint16(segfont.Seg7['1']) {
		t.Errorf("digit 0 = 0x%04X, want '1' (0x%04X)", spy.Calls[0].Segments[0], segfont.Seg7['1'])
	}
	if spy.Calls[0].Segments[1] != uint16(segfont.Seg7['2']) {
		t.Errorf("digit 1 = 0x%04X, want '2' (0x%04X)", spy.Calls[0].Segments[1], segfont.Seg7['2'])
	}
}

func TestSegmentClock_12h_Noon_IsPM(t *testing.T) {
	// Noon = 12:00 PM
	fixed := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
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
	// Noon is PM → first call colon on (PM double blink)
	if !spy.Calls[0].Colon {
		t.Error("expected colon on for PM (noon)")
	}
	// "1200" digits
	if spy.Calls[0].Segments[0] != uint16(segfont.Seg7['1']) {
		t.Errorf("digit 0 = 0x%04X, want '1'", spy.Calls[0].Segments[0])
	}
}

func TestSegmentClock_DefaultEncoder(t *testing.T) {
	fixed := time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC)
	spy := &testutil.SpySegmentDisplay{}
	ctx, cancel := context.WithCancel(context.Background())

	clk := &segment.Clock{
		Format24h: true,
		NowFunc:   func() time.Time { return fixed },
		// No Encoder set — should default to Enc7
	}

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	clk.Run(ctx, spy)

	if len(spy.Calls) == 0 {
		t.Fatal("expected segment writes")
	}
	// Should use Enc7 by default, so '1' = 0x06
	if spy.Calls[0].Segments[0] != 0x06 {
		t.Errorf("digit 0 = 0x%04X, want 0x0006 (default Enc7)", spy.Calls[0].Segments[0])
	}
}

func TestSegmentClock_With14SegEncoder(t *testing.T) {
	fixed := time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC)
	spy := &testutil.SpySegmentDisplay{}
	ctx, cancel := context.WithCancel(context.Background())

	clk := &segment.Clock{
		Format24h: true,
		NowFunc:   func() time.Time { return fixed },
		Encoder:   segfont.Enc14,
	}

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	clk.Run(ctx, spy)

	if len(spy.Calls) == 0 {
		t.Fatal("expected segment writes")
	}
}

func TestSegmentClock_Name(t *testing.T) {
	clk := &segment.Clock{}
	if clk.Name() != "segment-clock" {
		t.Errorf("Name() = %q, want %q", clk.Name(), "segment-clock")
	}
}

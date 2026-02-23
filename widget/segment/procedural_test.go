package segment_test

import (
	"context"
	"testing"
	"time"

	"github.com/swilcox/led-kurokku-go/display/testutil"
	"github.com/swilcox/led-kurokku-go/widget/segment"
)

func TestSegmentRain_WritesSegments(t *testing.T) {
	spy := &testutil.SpySegmentDisplay{}
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	(&segment.Rain{}).Run(ctx, spy)
	if len(spy.Calls) == 0 {
		t.Error("expected segment writes from Rain animation")
	}
}

func TestSegmentRain14_WritesSegments(t *testing.T) {
	spy := &testutil.SpySegmentDisplay{}
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	(&segment.Rain14{}).Run(ctx, spy)
	if len(spy.Calls) == 0 {
		t.Error("expected segment writes from Rain14 animation")
	}
}

func TestSegmentStatic_WritesSegments(t *testing.T) {
	spy := &testutil.SpySegmentDisplay{}
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	(&segment.Static{}).Run(ctx, spy)
	if len(spy.Calls) == 0 {
		t.Error("expected segment writes from Static animation")
	}
	// Each call should have 4 segments
	for i, call := range spy.Calls {
		if len(call.Segments) != 4 {
			t.Errorf("call[%d] has %d segments, want 4", i, len(call.Segments))
			break
		}
	}
}

func TestSegmentStatic14_WritesSegments(t *testing.T) {
	spy := &testutil.SpySegmentDisplay{}
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	(&segment.Static14{}).Run(ctx, spy)
	if len(spy.Calls) == 0 {
		t.Error("expected segment writes from Static14 animation")
	}
}

func TestSegmentStatic_SegmentsVary(t *testing.T) {
	spy := &testutil.SpySegmentDisplay{}
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	(&segment.Static{}).Run(ctx, spy)
	if len(spy.Calls) < 2 {
		t.Skip("not enough calls to check variation")
	}
	allSame := true
	for _, call := range spy.Calls[1:] {
		for j, seg := range call.Segments {
			if seg != spy.Calls[0].Segments[j] {
				allSame = false
				break
			}
		}
		if !allSame {
			break
		}
	}
	if allSame {
		t.Error("expected static segments to vary between calls")
	}
}

func TestSegmentScanner_WritesSegments(t *testing.T) {
	spy := &testutil.SpySegmentDisplay{}
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	(&segment.Scanner{}).Run(ctx, spy)
	if len(spy.Calls) == 0 {
		t.Error("expected segment writes from Scanner animation")
	}
	// Each call should have exactly one non-zero digit (the bar)
	for i, call := range spy.Calls {
		nonZero := 0
		for _, s := range call.Segments {
			if s != 0 {
				nonZero++
			}
		}
		if nonZero != 1 {
			t.Errorf("call[%d] has %d non-zero segments, want 1", i, nonZero)
			break
		}
	}
}

func TestSegmentScanner14_WritesSegments(t *testing.T) {
	spy := &testutil.SpySegmentDisplay{}
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	(&segment.Scanner14{}).Run(ctx, spy)
	if len(spy.Calls) == 0 {
		t.Error("expected segment writes from Scanner14 animation")
	}
}

func TestSegmentRace_WritesSegments(t *testing.T) {
	spy := &testutil.SpySegmentDisplay{}
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	(&segment.Race{}).Run(ctx, spy)
	if len(spy.Calls) == 0 {
		t.Error("expected segment writes from Race animation")
	}
	// Each call should have exactly two non-zero segment bits across all digits
	for i, call := range spy.Calls {
		totalBits := 0
		for _, s := range call.Segments {
			for b := 0; b < 7; b++ {
				if s&(1<<uint(b)) != 0 {
					totalBits++
				}
			}
		}
		if totalBits != 2 {
			t.Errorf("call[%d] has %d lit segment bits, want 2", i, totalBits)
			break
		}
	}
}

func TestSegmentRace14_WritesSegments(t *testing.T) {
	spy := &testutil.SpySegmentDisplay{}
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	(&segment.Race14{}).Run(ctx, spy)
	if len(spy.Calls) == 0 {
		t.Error("expected segment writes from Race14 animation")
	}
}

func TestSegmentRegistry(t *testing.T) {
	for _, name := range []string{"rain", "static", "scanner", "race"} {
		factory, ok := segment.Registry[name]
		if !ok {
			t.Errorf("Registry missing %q", name)
			continue
		}
		w := factory()
		if w == nil {
			t.Errorf("Registry[%q] returned nil", name)
		}
	}
}

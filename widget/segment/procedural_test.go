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

func TestSegmentRandom_WritesSegments(t *testing.T) {
	spy := &testutil.SpySegmentDisplay{}
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	(&segment.Random{}).Run(ctx, spy)
	if len(spy.Calls) == 0 {
		t.Error("expected segment writes from Random animation")
	}
	// Each call should have 4 segments
	for i, call := range spy.Calls {
		if len(call.Segments) != 4 {
			t.Errorf("call[%d] has %d segments, want 4", i, len(call.Segments))
			break
		}
	}
}

func TestSegmentRandom14_WritesSegments(t *testing.T) {
	spy := &testutil.SpySegmentDisplay{}
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	(&segment.Random14{}).Run(ctx, spy)
	if len(spy.Calls) == 0 {
		t.Error("expected segment writes from Random14 animation")
	}
}

func TestSegmentRandom_SegmentsVary(t *testing.T) {
	spy := &testutil.SpySegmentDisplay{}
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	(&segment.Random{}).Run(ctx, spy)
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
		t.Error("expected random segments to vary between calls")
	}
}

func TestSegmentRegistry(t *testing.T) {
	for _, name := range []string{"rain", "random"} {
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

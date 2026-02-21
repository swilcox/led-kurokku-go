package animation_test

import (
	"context"
	"testing"
	"time"

	"github.com/swilcox/led-kurokku-go/config"
	"github.com/swilcox/led-kurokku-go/display/testutil"
	"github.com/swilcox/led-kurokku-go/widget/animation"
)

func TestFrameAnimation_Empty_ReturnsImmediately(t *testing.T) {
	spy := &testutil.SpyDisplay{}
	a := &animation.FrameAnimation{}
	a.Run(context.Background(), spy) //nolint:errcheck
	if len(spy.Frames) != 0 {
		t.Errorf("expected no frames for empty animation, got %d", len(spy.Frames))
	}
}

func TestFrameAnimation_WritesFramesInOrder(t *testing.T) {
	spy := &testutil.SpyDisplay{}
	ctx, cancel := context.WithCancel(context.Background())

	frames := []config.FrameConfig{
		{Data: [32]byte{0x01}, Duration: config.Duration(time.Millisecond)},
		{Data: [32]byte{0x02}, Duration: config.Duration(time.Millisecond)},
		{Data: [32]byte{0x03}, Duration: config.Duration(time.Millisecond)},
	}

	a := &animation.FrameAnimation{
		Frames:        frames,
		FrameDuration: time.Millisecond,
	}

	go func() {
		// Cancel after enough time for at least one full cycle.
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	a.Run(ctx, spy) //nolint:errcheck

	if len(spy.Frames) < 3 {
		t.Fatalf("expected at least 3 frames, got %d", len(spy.Frames))
	}
	// First cycle: frames should appear in order.
	if spy.Frames[0][0] != 0x01 {
		t.Errorf("frame[0][0] = %02x, want 0x01", spy.Frames[0][0])
	}
	if spy.Frames[1][0] != 0x02 {
		t.Errorf("frame[1][0] = %02x, want 0x02", spy.Frames[1][0])
	}
	if spy.Frames[2][0] != 0x03 {
		t.Errorf("frame[2][0] = %02x, want 0x03", spy.Frames[2][0])
	}
}

func TestFrameAnimation_UsesFrameDurationFallback(t *testing.T) {
	spy := &testutil.SpyDisplay{}
	ctx, cancel := context.WithCancel(context.Background())

	// Frame has no per-frame duration; FrameDuration should be used.
	frames := []config.FrameConfig{
		{Data: [32]byte{0xFF}}, // Duration = 0 → falls back to FrameDuration
	}

	a := &animation.FrameAnimation{
		Frames:        frames,
		FrameDuration: time.Millisecond,
	}

	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	a.Run(ctx, spy) //nolint:errcheck

	if len(spy.Frames) == 0 {
		t.Error("expected at least one frame")
	}
	if spy.Frames[0][0] != 0xFF {
		t.Errorf("frame[0][0] = %02x, want 0xFF", spy.Frames[0][0])
	}
}

func TestBounce_WritesFrames(t *testing.T) {
	spy := &testutil.SpyDisplay{}
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	(&animation.Bounce{}).Run(ctx, spy) //nolint:errcheck
	if len(spy.Frames) == 0 {
		t.Error("expected frames from Bounce animation")
	}
	// Each frame should have at least one lit pixel.
	lit := false
	for _, col := range spy.Frames[0] {
		if col != 0 {
			lit = true
			break
		}
	}
	if !lit {
		t.Error("expected at least one lit pixel in Bounce frame")
	}
}

func TestSine_WritesFrames(t *testing.T) {
	spy := &testutil.SpyDisplay{}
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	(&animation.Sine{}).Run(ctx, spy) //nolint:errcheck
	if len(spy.Frames) == 0 {
		t.Error("expected frames from Sine animation")
	}
	// Each frame should have exactly 32 lit pixels — one per column.
	bits := 0
	for _, col := range spy.Frames[0] {
		for b := 0; b < 8; b++ {
			if col&(1<<uint(b)) != 0 {
				bits++
			}
		}
	}
	if bits != 32 {
		t.Errorf("expected 32 lit pixels in first Sine frame (one per column), got %d", bits)
	}
}

func TestScanner_WritesFrames(t *testing.T) {
	spy := &testutil.SpyDisplay{}
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	(&animation.Scanner{}).Run(ctx, spy) //nolint:errcheck
	if len(spy.Frames) == 0 {
		t.Error("expected frames from Scanner animation")
	}
	// The head column (pos=0 on first frame) should be fully lit.
	if spy.Frames[0][0] != 0xFF {
		t.Errorf("expected first Scanner column fully lit (0xFF), got 0x%02X", spy.Frames[0][0])
	}
}

func TestLife_WritesFrames(t *testing.T) {
	spy := &testutil.SpyDisplay{}
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	(&animation.Life{}).Run(ctx, spy) //nolint:errcheck
	if len(spy.Frames) == 0 {
		t.Error("expected frames from Life animation")
	}
}

func TestLife_FramesChangeOverTime(t *testing.T) {
	spy := &testutil.SpyDisplay{}
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	(&animation.Life{}).Run(ctx, spy) //nolint:errcheck
	if len(spy.Frames) < 2 {
		t.Fatal("expected at least 2 frames")
	}
	// It would be extremely unlikely for all frames to be identical.
	allSame := true
	for _, f := range spy.Frames[1:] {
		if !bytesEqual(f, spy.Frames[0]) {
			allSame = false
			break
		}
	}
	if allSame {
		t.Error("expected Life frames to change over time")
	}
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

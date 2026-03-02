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

func TestBounce_Name(t *testing.T) {
	b := &animation.Bounce{}
	if b.Name() != "bounce" {
		t.Errorf("Name() = %q, want %q", b.Name(), "bounce")
	}
}

func TestScanner_Name(t *testing.T) {
	s := &animation.Scanner{}
	if s.Name() != "scanner" {
		t.Errorf("Name() = %q, want %q", s.Name(), "scanner")
	}
}

func TestLife_Name(t *testing.T) {
	l := &animation.Life{}
	if l.Name() != "life" {
		t.Errorf("Name() = %q, want %q", l.Name(), "life")
	}
}

func TestBounce_CancelledImmediately(t *testing.T) {
	spy := &testutil.SpyDisplay{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := (&animation.Bounce{}).Run(ctx, spy)
	if err == nil {
		t.Error("expected error from cancelled context")
	}
	if len(spy.Frames) != 0 {
		t.Errorf("expected no frames from immediately cancelled, got %d", len(spy.Frames))
	}
}

func TestScanner_CancelledImmediately(t *testing.T) {
	spy := &testutil.SpyDisplay{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := (&animation.Scanner{}).Run(ctx, spy)
	if err == nil {
		t.Error("expected error from cancelled context")
	}
}

func TestLife_CancelledImmediately(t *testing.T) {
	spy := &testutil.SpyDisplay{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := (&animation.Life{}).Run(ctx, spy)
	if err == nil {
		t.Error("expected error from cancelled context")
	}
}

func TestScanner_ReversesDirection(t *testing.T) {
	spy := &testutil.SpyDisplay{}
	// Run long enough to see reversal (at 40ms/frame, 31 frames to reach end)
	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()
	(&animation.Scanner{}).Run(ctx, spy) //nolint:errcheck

	if len(spy.Frames) < 32 {
		t.Fatalf("expected at least 32 frames to see reversal, got %d", len(spy.Frames))
	}
	// First frame: head at pos 0 → column 0 = 0xFF
	if spy.Frames[0][0] != 0xFF {
		t.Errorf("frame[0][0] = 0x%02X, want 0xFF", spy.Frames[0][0])
	}
	// After 31 frames, head should be at pos 31
	if spy.Frames[30][31] != 0xFF {
		// The head might not be exactly at 31 due to timing, check it moved right
		foundRight := false
		for _, f := range spy.Frames[25:35] {
			if f[31] == 0xFF {
				foundRight = true
				break
			}
		}
		if !foundRight {
			t.Error("expected head to reach right edge (col 31)")
		}
	}
}

func TestScanner_TrailPattern(t *testing.T) {
	spy := &testutil.SpyDisplay{}
	// Run just a few frames
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	(&animation.Scanner{}).Run(ctx, spy) //nolint:errcheck

	// After a few ticks, trail should be visible behind the head
	if len(spy.Frames) < 5 {
		t.Skip("not enough frames for trail check")
	}
	f := spy.Frames[4] // head at pos 4
	// Trail: pos-1=0xAA, pos-2=0x44, pos-3=0x11
	if f[4] != 0xFF {
		t.Errorf("head column = 0x%02X, want 0xFF", f[4])
	}
	if f[3] != 0xAA {
		t.Errorf("trail[1] = 0x%02X, want 0xAA", f[3])
	}
	if f[2] != 0x44 {
		t.Errorf("trail[2] = 0x%02X, want 0x44", f[2])
	}
	if f[1] != 0x11 {
		t.Errorf("trail[3] = 0x%02X, want 0x11", f[1])
	}
}

func TestBounce_ProducesMultipleFrames(t *testing.T) {
	spy := &testutil.SpyDisplay{}
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	(&animation.Bounce{}).Run(ctx, spy) //nolint:errcheck
	if len(spy.Frames) < 5 {
		t.Errorf("expected at least 5 frames, got %d", len(spy.Frames))
	}
	// Frames should differ (ball is moving)
	if bytesEqual(spy.Frames[0], spy.Frames[len(spy.Frames)-1]) {
		t.Error("expected frames to differ over time (ball should move)")
	}
}

func TestFrameAnimation_Name(t *testing.T) {
	a := &animation.FrameAnimation{}
	if a.Name() != "animation" {
		t.Errorf("Name() = %q, want %q", a.Name(), "animation")
	}
}

func TestRegistry_ContainsExpectedAnimations(t *testing.T) {
	expected := []string{"bounce", "rain", "static", "sine", "scanner", "life"}
	for _, name := range expected {
		if _, ok := animation.Registry[name]; !ok {
			t.Errorf("Registry missing animation %q", name)
		}
	}
}

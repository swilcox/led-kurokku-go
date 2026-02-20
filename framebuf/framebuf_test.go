package framebuf_test

import (
	"testing"

	"github.com/swilcox/led-kurokku-go/framebuf"
)

func TestSetGetPixel_RoundTrip(t *testing.T) {
	var f framebuf.Frame
	f.SetPixel(5, 3, true)
	if !f.GetPixel(5, 3) {
		t.Error("pixel (5,3) should be set")
	}
	f.SetPixel(5, 3, false)
	if f.GetPixel(5, 3) {
		t.Error("pixel (5,3) should be cleared after SetPixel false")
	}
}

func TestGetPixel_OutOfBounds(t *testing.T) {
	var f framebuf.Frame
	cases := [][2]int{{-1, 0}, {32, 0}, {0, -1}, {0, 8}}
	for _, c := range cases {
		if f.GetPixel(c[0], c[1]) {
			t.Errorf("GetPixel(%d,%d) should return false for out-of-bounds", c[0], c[1])
		}
	}
}

func TestSetPixel_OutOfBounds_NoOp(t *testing.T) {
	var f framebuf.Frame
	// Should not panic
	f.SetPixel(-1, 0, true)
	f.SetPixel(32, 0, true)
	f.SetPixel(0, -1, true)
	f.SetPixel(0, 8, true)
	for x := 0; x < 32; x++ {
		for y := 0; y < 8; y++ {
			if f.GetPixel(x, y) {
				t.Errorf("out-of-bounds SetPixel should not affect pixel (%d,%d)", x, y)
			}
		}
	}
}

func TestClear(t *testing.T) {
	var f framebuf.Frame
	f.SetPixel(0, 0, true)
	f.SetPixel(31, 7, true)
	f.Clear()
	for x := 0; x < 32; x++ {
		for y := 0; y < 8; y++ {
			if f.GetPixel(x, y) {
				t.Errorf("pixel (%d,%d) should be cleared", x, y)
			}
		}
	}
}

func TestBlitText_Width(t *testing.T) {
	var f framebuf.Frame
	w := framebuf.BlitText(&f, "A", 0)
	if w != 5 {
		t.Errorf("expected width 5 for single char 'A', got %d", w)
	}
}

func TestBlitText_Offset(t *testing.T) {
	var f1, f2 framebuf.Frame
	framebuf.BlitText(&f1, "A", 0)
	framebuf.BlitText(&f2, "A", 5)
	b1 := f1.Bytes()
	b2 := f2.Bytes()
	for i := 0; i < 5; i++ {
		if b1[i] != b2[i+5] {
			t.Errorf("blit offset mismatch at col %d: %02x != %02x", i, b1[i], b2[i+5])
		}
	}
}

func TestBlitText_ClipsAtEdge(t *testing.T) {
	var f framebuf.Frame
	// Blitting at offset 30 with a 5-col glyph: only cols 30 and 31 written.
	w := framebuf.BlitText(&f, "A", 30)
	if w != 5 {
		t.Errorf("BlitText should return full text width even when clipped, got %d", w)
	}
	// Columns 0-29 should be zero (never written).
	for x := 0; x < 30; x++ {
		if f.Bytes()[x] != 0 {
			t.Errorf("col %d should be 0 when blitting at offset 30", x)
		}
	}
}

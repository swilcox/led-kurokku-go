package font_test

import (
	"testing"

	"github.com/swilcox/led-kurokku-go/font"
)

func TestRenderText_Empty(t *testing.T) {
	cols := font.RenderText("")
	if len(cols) != 0 {
		t.Errorf("expected 0 cols for empty string, got %d", len(cols))
	}
}

func TestRenderText_SingleChar(t *testing.T) {
	// A single glyph is 5 columns wide; no trailing gap.
	cols := font.RenderText("A")
	if len(cols) != 5 {
		t.Errorf("expected 5 cols for single char, got %d", len(cols))
	}
}

func TestRenderText_TwoChars(t *testing.T) {
	// Two glyphs: 5 + 1 (gap) + 5 = 11 columns.
	cols := font.RenderText("Hi")
	if len(cols) != 11 {
		t.Errorf("expected 11 cols for two chars, got %d", len(cols))
	}
}

func TestRenderText_UnknownCharFallsBackToQuestion(t *testing.T) {
	// \x01 is not in the font (ASCII 32-126); should fall back to '?'
	colsUnknown := font.RenderText("\x01")
	colsQ := font.RenderText("?")
	if len(colsUnknown) != len(colsQ) {
		t.Fatalf("unknown char width %d != '?' width %d", len(colsUnknown), len(colsQ))
	}
	for i := range colsQ {
		if colsUnknown[i] != colsQ[i] {
			t.Errorf("unknown char fallback: col[%d] = %02x, want %02x", i, colsUnknown[i], colsQ[i])
		}
	}
}

func TestRenderText_GapIsZero(t *testing.T) {
	// The inter-character gap column must be 0x00.
	cols := font.RenderText("AB") // 5 + gap + 5
	if len(cols) != 11 {
		t.Fatalf("expected 11 cols, got %d", len(cols))
	}
	if cols[5] != 0x00 {
		t.Errorf("gap column should be 0x00, got %02x", cols[5])
	}
}

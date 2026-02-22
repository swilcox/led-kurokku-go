package font_test

import (
	"testing"

	"github.com/swilcox/led-kurokku-go/font"
)

func TestIsHangul(t *testing.T) {
	tests := []struct {
		r    rune
		want bool
	}{
		{'가', true},  // U+AC00 first syllable
		{'힣', true},  // U+D7A3 last syllable
		{'한', true},  // common syllable
		{'A', false},  // ASCII
		{'0', false},  // digit
		{0xABFF, false}, // just before range
		{0xD7A4, false}, // just after range
	}
	for _, tt := range tests {
		if got := font.IsHangul(tt.r); got != tt.want {
			t.Errorf("IsHangul(%q) = %v, want %v", tt.r, got, tt.want)
		}
	}
}

func TestHangulGlyph_KnownSyllables(t *testing.T) {
	for _, r := range []rune{'가', '한', '글'} {
		g, ok := font.HangulGlyph(r)
		if !ok {
			t.Errorf("HangulGlyph(%q) returned false", r)
			continue
		}
		// Glyph should not be all zeros (blank).
		allZero := true
		for _, col := range g {
			if col != 0 {
				allZero = false
				break
			}
		}
		if allZero {
			t.Errorf("HangulGlyph(%q) returned all-zero bitmap", r)
		}
	}
}

func TestHangulGlyph_NonHangul(t *testing.T) {
	_, ok := font.HangulGlyph('A')
	if ok {
		t.Error("HangulGlyph('A') should return false")
	}
}

func TestRenderText_HangulWidth(t *testing.T) {
	// Two Hangul characters: 8 + 1 (gap) + 8 = 17 columns.
	cols := font.RenderText("가나")
	if len(cols) != 17 {
		t.Errorf("expected 17 cols for two Hangul chars, got %d", len(cols))
	}
}

func TestRenderText_SingleHangul(t *testing.T) {
	cols := font.RenderText("가")
	if len(cols) != 8 {
		t.Errorf("expected 8 cols for single Hangul char, got %d", len(cols))
	}
}

func TestRenderText_MixedKoreanASCII(t *testing.T) {
	// "Hi 안녕" = H(5) + gap + i(5) + gap + space(5) + gap + 안(8) + gap + 녕(8)
	// = 5+1+5+1+5+1+8+1+8 = 35
	cols := font.RenderText("Hi 안녕")
	expected := 5 + 1 + 5 + 1 + 5 + 1 + 8 + 1 + 8
	if len(cols) != expected {
		t.Errorf("expected %d cols for mixed text, got %d", expected, len(cols))
	}
}

func TestRenderText_HangulFallback(t *testing.T) {
	// A rune in the Hangul range that's in the font should render as 8 cols.
	// If we somehow had a missing entry, it would fall through to '?'.
	// We test with a non-Hangul, non-ASCII rune to verify fallback.
	cols := font.RenderText("\u00FF") // ÿ - not in Font5x7, not Hangul
	colsQ := font.RenderText("?")
	if len(cols) != len(colsQ) {
		t.Errorf("non-Hangul non-ASCII should fall back to '?': got %d cols, want %d", len(cols), len(colsQ))
	}
}

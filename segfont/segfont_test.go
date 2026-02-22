package segfont

import "testing"

func TestSeg7_Digits(t *testing.T) {
	tests := []struct {
		r    rune
		want byte
	}{
		{'0', 0x3F},
		{'1', 0x06},
		{'2', 0x5B},
		{'3', 0x4F},
		{'4', 0x66},
		{'5', 0x6D},
		{'6', 0x7D},
		{'7', 0x07},
		{'8', 0x7F},
		{'9', 0x6F},
	}
	for _, tt := range tests {
		got := Seg7[tt.r]
		if got != tt.want {
			t.Errorf("Seg7[%q] = 0x%02X, want 0x%02X", tt.r, got, tt.want)
		}
	}
}

func TestSeg7_Letters(t *testing.T) {
	tests := []struct {
		r    rune
		want byte
	}{
		{'A', 0x77},
		{'E', 0x79},
		{'H', 0x76},
		{'I', 0x30},
		{'w', 0x2A},
		{'m', 0x55},
		{'n', 0x54},
		{'q', 0x67},
		{' ', 0x00},
		{'-', 0x40},
	}
	for _, tt := range tests {
		got := Seg7[tt.r]
		if got != tt.want {
			t.Errorf("Seg7[%q] = 0x%02X, want 0x%02X", tt.r, got, tt.want)
		}
	}
}

func TestSeg7_DegreeSymbol(t *testing.T) {
	// Both '*' and '°' map to degree (0x63)
	if Seg7['*'] != 0x63 {
		t.Errorf("Seg7['*'] = 0x%02X, want 0x63", Seg7['*'])
	}
	if Seg7['°'] != 0x63 {
		t.Errorf("Seg7['°'] = 0x%02X, want 0x63", Seg7['°'])
	}
	if Seg7['*'] != Seg7['°'] {
		t.Error("expected '*' and '°' to map to the same segment value")
	}
}

func TestSeg7_AllLowercaseLetters(t *testing.T) {
	for r := 'a'; r <= 'z'; r++ {
		if _, ok := Seg7[r]; !ok {
			t.Errorf("Seg7 missing lowercase %q", r)
		}
	}
}

func TestSeg7_LowercaseFallbackMatchesUppercase(t *testing.T) {
	// Letters that should use uppercase representation
	fallbacks := []rune{'a', 'e', 'f', 'g', 'i', 'j', 'l', 'p', 's'}
	for _, r := range fallbacks {
		upper := r - 32 // 'a'-'A' = 32
		if Seg7[r] != Seg7[upper] {
			t.Errorf("Seg7[%q] = 0x%02X, want 0x%02X (same as %q)", r, Seg7[r], Seg7[upper], upper)
		}
	}
}

func TestSeg7_UnknownRune(t *testing.T) {
	got := Seg7['@']
	if got != 0 {
		t.Errorf("Seg7['@'] = 0x%02X, want 0x00 for unknown rune", got)
	}
}

func TestSeg14_Digits(t *testing.T) {
	tests := []struct {
		r    rune
		want uint16
	}{
		{'0', 0x003F},
		{'1', 0x0006},
		{'2', 0x00DB},
		{'8', 0x00FF},
	}
	for _, tt := range tests {
		got := Seg14[tt.r]
		if got != tt.want {
			t.Errorf("Seg14[%q] = 0x%04X, want 0x%04X", tt.r, got, tt.want)
		}
	}
}

func TestSeg14_Letters(t *testing.T) {
	tests := []struct {
		r    rune
		want uint16
	}{
		{'A', 0x00F7},
		{'W', 0x2836},
		{'I', 0x1209},
		{'T', 0x1201},
	}
	for _, tt := range tests {
		got := Seg14[tt.r]
		if got != tt.want {
			t.Errorf("Seg14[%q] = 0x%04X, want 0x%04X", tt.r, got, tt.want)
		}
	}
}

func TestSeg14_DegreeSymbol(t *testing.T) {
	if Seg14['°'] != 0x00E3 {
		t.Errorf("Seg14['°'] = 0x%04X, want 0x00E3", Seg14['°'])
	}
}

func TestSeg14_Asterisk(t *testing.T) {
	if Seg14['*'] != 0x2DC0 {
		t.Errorf("Seg14['*'] = 0x%04X, want 0x2DC0", Seg14['*'])
	}
}

func TestSeg14_Space(t *testing.T) {
	if Seg14[' '] != 0 {
		t.Errorf("Seg14[' '] = 0x%04X, want 0x0000", Seg14[' '])
	}
}

func TestSeg14_UnknownRune(t *testing.T) {
	got := Seg14['~']
	if got != 0 {
		t.Errorf("Seg14['~'] = 0x%04X, want 0x0000 for unknown rune", got)
	}
}

func TestEnc7(t *testing.T) {
	got := Enc7('0')
	if got != 0x3F {
		t.Errorf("Enc7('0') = 0x%04X, want 0x003F", got)
	}
	if Enc7('~') != 0 {
		t.Errorf("Enc7('~') = 0x%04X, want 0x0000", Enc7('~'))
	}
}

func TestEnc14(t *testing.T) {
	got := Enc14('A')
	if got != 0x00F7 {
		t.Errorf("Enc14('A') = 0x%04X, want 0x00F7", got)
	}
	if Enc14('~') != 0 {
		t.Errorf("Enc14('~') = 0x%04X, want 0x0000", Enc14('~'))
	}
}

func TestEncodeText(t *testing.T) {
	result := EncodeText(Enc7, "12")
	if len(result) != 2 {
		t.Fatalf("EncodeText(Enc7, \"12\") length = %d, want 2", len(result))
	}
	if result[0] != uint16(Seg7['1']) {
		t.Errorf("result[0] = 0x%04X, want 0x%04X", result[0], Seg7['1'])
	}
	if result[1] != uint16(Seg7['2']) {
		t.Errorf("result[1] = 0x%04X, want 0x%04X", result[1], Seg7['2'])
	}
}

func TestEncodeText_Degree(t *testing.T) {
	result := EncodeText(Enc7, "72°F")
	if len(result) != 4 {
		t.Fatalf("length = %d, want 4", len(result))
	}
	if result[2] != 0x63 {
		t.Errorf("degree segment = 0x%04X, want 0x0063", result[2])
	}
}

func TestEncodeText_Empty(t *testing.T) {
	result := EncodeText(Enc7, "")
	if len(result) != 0 {
		t.Errorf("EncodeText(Enc7, \"\") length = %d, want 0", len(result))
	}
}

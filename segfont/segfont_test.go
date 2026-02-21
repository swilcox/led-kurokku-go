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
	// Spot-check a few common letters
	if Seg7['A'] != 0x77 {
		t.Errorf("Seg7['A'] = 0x%02X, want 0x77", Seg7['A'])
	}
	if Seg7['E'] != 0x79 {
		t.Errorf("Seg7['E'] = 0x%02X, want 0x79", Seg7['E'])
	}
	if Seg7[' '] != 0x00 {
		t.Errorf("Seg7[' '] = 0x%02X, want 0x00", Seg7[' '])
	}
	if Seg7['-'] != 0x40 {
		t.Errorf("Seg7['-'] = 0x%02X, want 0x40", Seg7['-'])
	}
}

func TestSeg7_UnknownRune(t *testing.T) {
	got := Seg7['@']
	if got != 0 {
		t.Errorf("Seg7['@'] = 0x%02X, want 0x00 for unknown rune", got)
	}
}

func TestSeg14_Digits(t *testing.T) {
	// Verify all digits are non-zero
	for _, r := range "0123456789" {
		if Seg14[r] == 0 {
			t.Errorf("Seg14[%q] = 0, expected non-zero", r)
		}
	}
}

func TestSeg14_Letters(t *testing.T) {
	// Spot-check a few letters
	if Seg14['A'] == 0 {
		t.Error("Seg14['A'] should be non-zero")
	}
	if Seg14['Z'] == 0 {
		t.Error("Seg14['Z'] should be non-zero")
	}
	if Seg14[' '] != 0 {
		t.Errorf("Seg14[' '] = 0x%04X, want 0x0000", Seg14[' '])
	}
}

func TestSeg14_UnknownRune(t *testing.T) {
	got := Seg14['@']
	if got != 0 {
		t.Errorf("Seg14['@'] = 0x%04X, want 0x0000 for unknown rune", got)
	}
}

func TestEnc7(t *testing.T) {
	got := Enc7('0')
	if got != 0x3F {
		t.Errorf("Enc7('0') = 0x%04X, want 0x003F", got)
	}
	// Unknown rune returns 0
	if Enc7('~') != 0 {
		t.Errorf("Enc7('~') = 0x%04X, want 0x0000", Enc7('~'))
	}
}

func TestEnc14(t *testing.T) {
	got := Enc14('A')
	if got == 0 {
		t.Error("Enc14('A') should be non-zero")
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

func TestEncodeText_Empty(t *testing.T) {
	result := EncodeText(Enc7, "")
	if len(result) != 0 {
		t.Errorf("EncodeText(Enc7, \"\") length = %d, want 0", len(result))
	}
}

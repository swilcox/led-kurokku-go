package display

import (
	"testing"

	"periph.io/x/conn/v3/gpio"
)

// bitBangEvent records a single pin output.
type bitBangEvent struct {
	pin   string
	level gpio.Level
}

// bitBangRecorder captures events from both clk and dio pins.
type bitBangRecorder struct {
	events []bitBangEvent
}

// recordingPin implements gpioPin and records Out calls.
type recordingPin struct {
	name string
	rec  *bitBangRecorder
}

func (p *recordingPin) Out(l gpio.Level) error {
	p.rec.events = append(p.rec.events, bitBangEvent{p.name, l})
	return nil
}

// decodeTM1637Bytes extracts the bytes sent over the TM1637 bit-bang protocol.
// Returns groups of bytes, one group per start/stop frame.
func decodeTM1637Bytes(events []bitBangEvent) [][]byte {
	clk, dio := gpio.High, gpio.High

	var groups [][]byte
	var currentBytes []byte
	var currentByte byte
	var bitCount int
	inFrame := false

	for _, e := range events {
		prevClk, prevDio := clk, dio
		switch e.pin {
		case "clk":
			clk = e.level
		case "dio":
			dio = e.level
		}

		// Start condition: dio falls while clk is high
		if e.pin == "dio" && prevDio == gpio.High && dio == gpio.Low && clk == gpio.High {
			inFrame = true
			currentBytes = nil
			currentByte = 0
			bitCount = 0
			continue
		}

		// Stop condition: dio rises while clk is high
		if e.pin == "dio" && prevDio == gpio.Low && dio == gpio.High && clk == gpio.High {
			if inFrame {
				groups = append(groups, currentBytes)
			}
			inFrame = false
			continue
		}

		// Rising clock edge: sample data
		if inFrame && e.pin == "clk" && prevClk == gpio.Low && clk == gpio.High {
			if bitCount < 8 {
				if dio == gpio.High {
					currentByte |= 1 << uint(bitCount)
				}
				bitCount++
			} else {
				// ACK clock — byte complete
				currentBytes = append(currentBytes, currentByte)
				currentByte = 0
				bitCount = 0
			}
		}
	}

	return groups
}

func newTestTM1637(rec *bitBangRecorder) *TM1637 {
	return &TM1637{
		clk:    &recordingPin{name: "clk", rec: rec},
		dio:    &recordingPin{name: "dio", rec: rec},
		bright: 7,
	}
}

func TestTM1637_DisplayLength(t *testing.T) {
	tm := &TM1637{}
	if tm.DisplayLength() != 4 {
		t.Errorf("DisplayLength() = %d, want 4", tm.DisplayLength())
	}
}

func TestTM1637_SetBrightness(t *testing.T) {
	tests := []struct {
		name      string
		level     byte
		wantField byte
	}{
		{"zero", 0, 0},
		{"one", 1, 0},     // 1 >> 1 = 0
		{"two", 2, 1},     // 2 >> 1 = 1
		{"fourteen", 14, 7}, // 14 >> 1 = 7
		{"fifteen", 15, 7},  // 15 >> 1 = 7
		{"max", 255, 7},     // 255 >> 1 = 127, clamped to 7
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := &bitBangRecorder{}
			tm := newTestTM1637(rec)
			tm.SetBrightness(tt.level)

			if tm.bright != tt.wantField {
				t.Errorf("bright = %d, want %d", tm.bright, tt.wantField)
			}

			// Should send one frame with brightness command
			groups := decodeTM1637Bytes(rec.events)
			if len(groups) != 1 {
				t.Fatalf("expected 1 frame, got %d", len(groups))
			}
			if len(groups[0]) != 1 {
				t.Fatalf("expected 1 byte in frame, got %d", len(groups[0]))
			}
			want := tm1637Brightness | tt.wantField
			if groups[0][0] != want {
				t.Errorf("sent 0x%02X, want 0x%02X", groups[0][0], want)
			}
		})
	}
}

func TestTM1637_WriteSegments_Basic(t *testing.T) {
	rec := &bitBangRecorder{}
	tm := newTestTM1637(rec)

	segs := []uint16{0x3F, 0x06, 0x5B, 0x4F}
	tm.WriteSegments(segs, false)

	groups := decodeTM1637Bytes(rec.events)
	// 3 frames: auto-increment, data write, brightness
	if len(groups) != 3 {
		t.Fatalf("expected 3 frames, got %d", len(groups))
	}

	// Frame 0: auto-increment command
	if len(groups[0]) != 1 || groups[0][0] != tm1637AddrAuto {
		t.Errorf("frame 0: got %v, want [0x%02X]", groups[0], tm1637AddrAuto)
	}

	// Frame 1: start address + 4 data bytes
	if len(groups[1]) != 5 {
		t.Fatalf("frame 1: expected 5 bytes, got %d", len(groups[1]))
	}
	if groups[1][0] != tm1637StartAddr {
		t.Errorf("frame 1 start addr: got 0x%02X, want 0x%02X", groups[1][0], tm1637StartAddr)
	}
	for i, seg := range segs {
		want := byte(seg & 0xFF)
		if groups[1][i+1] != want {
			t.Errorf("frame 1 digit %d: got 0x%02X, want 0x%02X", i, groups[1][i+1], want)
		}
	}

	// Frame 2: brightness
	if len(groups[2]) != 1 || groups[2][0] != tm1637Brightness|7 {
		t.Errorf("frame 2: got %v, want [0x%02X]", groups[2], tm1637Brightness|7)
	}
}

func TestTM1637_WriteSegments_Colon(t *testing.T) {
	rec := &bitBangRecorder{}
	tm := newTestTM1637(rec)

	segs := []uint16{0x00, 0x06, 0x00, 0x00}
	tm.WriteSegments(segs, true)

	groups := decodeTM1637Bytes(rec.events)
	if len(groups) != 3 {
		t.Fatalf("expected 3 frames, got %d", len(groups))
	}

	// Digit 1 should have colon bit (0x80) OR'd in
	digit1 := groups[1][2] // index 0=startAddr, 1=digit0, 2=digit1
	if digit1 != 0x86 {    // 0x06 | 0x80
		t.Errorf("digit 1 with colon: got 0x%02X, want 0x86", digit1)
	}
}

func TestTM1637_WriteSegments_LowByteOnly(t *testing.T) {
	rec := &bitBangRecorder{}
	tm := newTestTM1637(rec)

	// High bytes should be masked off
	segs := []uint16{0xFF3F, 0xFF06, 0xFF5B, 0xFF4F}
	tm.WriteSegments(segs, false)

	groups := decodeTM1637Bytes(rec.events)
	// Data bytes should only contain low byte
	for i, want := range []byte{0x3F, 0x06, 0x5B, 0x4F} {
		if groups[1][i+1] != want {
			t.Errorf("digit %d: got 0x%02X, want 0x%02X", i, groups[1][i+1], want)
		}
	}
}

func TestTM1637_Clear(t *testing.T) {
	rec := &bitBangRecorder{}
	tm := newTestTM1637(rec)
	tm.Clear()

	groups := decodeTM1637Bytes(rec.events)
	if len(groups) != 3 {
		t.Fatalf("expected 3 frames, got %d", len(groups))
	}

	// Data frame should have all zeros
	for i := 1; i <= 4; i++ {
		if groups[1][i] != 0 {
			t.Errorf("clear digit %d: got 0x%02X, want 0x00", i-1, groups[1][i])
		}
	}
}

func TestTM1637_WriteSegments_FewerThan4(t *testing.T) {
	rec := &bitBangRecorder{}
	tm := newTestTM1637(rec)

	segs := []uint16{0x3F, 0x06}
	tm.WriteSegments(segs, false)

	groups := decodeTM1637Bytes(rec.events)
	// Data frame: start addr + 2 data bytes (loop limited by len(segments))
	if len(groups[1]) != 3 {
		t.Errorf("frame 1: expected 3 bytes, got %d", len(groups[1]))
	}
}

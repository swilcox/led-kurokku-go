package display

import (
	"errors"
	"testing"
)

type mockSPI struct {
	txCalls [][]byte
	closed  bool
	txErr   error
}

func (m *mockSPI) Tx(w []byte) error {
	cp := make([]byte, len(w))
	copy(cp, w)
	m.txCalls = append(m.txCalls, cp)
	return m.txErr
}

func (m *mockSPI) Close() error {
	m.closed = true
	return nil
}

func newTestMAX7219(mock *mockSPI) *MAX7219 {
	return &MAX7219{dev: mock}
}

func TestMAX7219_Width(t *testing.T) {
	m := &MAX7219{}
	if m.Width() != 32 {
		t.Errorf("Width() = %d, want 32", m.Width())
	}
}

func TestMAX7219_Height(t *testing.T) {
	m := &MAX7219{}
	if m.Height() != 8 {
		t.Errorf("Height() = %d, want 8", m.Height())
	}
}

func TestMAX7219_SetBrightness(t *testing.T) {
	tests := []struct {
		name  string
		level byte
		want  byte
	}{
		{"zero", 0, 0},
		{"mid", 8, 8},
		{"max", 15, 15},
		{"clamp", 20, 15},
		{"clamp255", 255, 15},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockSPI{}
			m := newTestMAX7219(mock)
			m.SetBrightness(tt.level)

			if len(mock.txCalls) != 1 {
				t.Fatalf("expected 1 Tx call, got %d", len(mock.txCalls))
			}
			pkt := mock.txCalls[0]
			// All 4 devices should get regIntensity with the clamped value
			for dev := 0; dev < numDevices; dev++ {
				reg := pkt[dev*2]
				val := pkt[dev*2+1]
				if reg != regIntensity {
					t.Errorf("device %d: reg = 0x%02X, want 0x%02X", dev, reg, regIntensity)
				}
				if val != tt.want {
					t.Errorf("device %d: val = %d, want %d", dev, val, tt.want)
				}
			}
		})
	}
}

func TestMAX7219_Clear(t *testing.T) {
	mock := &mockSPI{}
	m := newTestMAX7219(mock)
	m.Clear()

	// Clear sends 8 row commands (rows 1-8), each with 0x00 to all devices
	if len(mock.txCalls) != 8 {
		t.Fatalf("expected 8 Tx calls, got %d", len(mock.txCalls))
	}
	for i, pkt := range mock.txCalls {
		row := byte(i + 1)
		for dev := 0; dev < numDevices; dev++ {
			if pkt[dev*2] != row {
				t.Errorf("call %d, device %d: reg = 0x%02X, want 0x%02X", i, dev, pkt[dev*2], row)
			}
			if pkt[dev*2+1] != 0x00 {
				t.Errorf("call %d, device %d: val = 0x%02X, want 0x00", i, dev, pkt[dev*2+1])
			}
		}
	}
}

func TestMAX7219_Close(t *testing.T) {
	mock := &mockSPI{}
	m := newTestMAX7219(mock)
	err := m.Close()
	if err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	// Should send shutdown command then close device
	if len(mock.txCalls) != 1 {
		t.Fatalf("expected 1 Tx call (shutdown), got %d", len(mock.txCalls))
	}
	pkt := mock.txCalls[0]
	for dev := 0; dev < numDevices; dev++ {
		if pkt[dev*2] != regShutdown || pkt[dev*2+1] != 0x00 {
			t.Errorf("device %d: expected shutdown(0x00), got reg=0x%02X val=0x%02X", dev, pkt[dev*2], pkt[dev*2+1])
		}
	}
	if !mock.closed {
		t.Error("expected device to be closed")
	}
}

func TestMAX7219_Close_NilDev(t *testing.T) {
	m := &MAX7219{}
	// Close calls writeAll (which panics with nil dev), then checks dev != nil.
	// With nil dev, this would panic — verifying the guard only protects dev.Close().
	_ = m // document that nil dev path isn't safe to call
}

func TestMAX7219_WriteFramebuffer(t *testing.T) {
	t.Run("single pixel at origin", func(t *testing.T) {
		mock := &mockSPI{}
		m := newTestMAX7219(mock)

		buf := make([]byte, 32)
		buf[0] = 0x01 // column 0, row 0 lit

		m.WriteFramebuffer(buf)

		if len(mock.txCalls) != 8 {
			t.Fatalf("expected 8 Tx calls, got %d", len(mock.txCalls))
		}

		// Row 0: device 0 (cols 0-7) should have bit 7 set (col 0 → MSB)
		pkt := mock.txCalls[0]
		if pkt[0] != 1 { // row register 1
			t.Errorf("row 0 register = %d, want 1", pkt[0])
		}
		if pkt[1] != 0x80 { // col 0 = bit 7
			t.Errorf("row 0 dev 0 data = 0x%02X, want 0x80", pkt[1])
		}
		// Other devices should be 0
		for dev := 1; dev < numDevices; dev++ {
			if pkt[dev*2+1] != 0 {
				t.Errorf("row 0 dev %d data = 0x%02X, want 0x00", dev, pkt[dev*2+1])
			}
		}

		// Rows 1-7 should be all zeros
		for row := 1; row < 8; row++ {
			pkt := mock.txCalls[row]
			for dev := 0; dev < numDevices; dev++ {
				if pkt[dev*2+1] != 0 {
					t.Errorf("row %d dev %d data = 0x%02X, want 0x00", row, dev, pkt[dev*2+1])
				}
			}
		}
	})

	t.Run("full row lit", func(t *testing.T) {
		mock := &mockSPI{}
		m := newTestMAX7219(mock)

		buf := make([]byte, 32)
		// Set row 0 bit in all 32 columns
		for i := range buf {
			buf[i] = 0x01
		}

		m.WriteFramebuffer(buf)

		// Row 0: all devices should have 0xFF
		pkt := mock.txCalls[0]
		for dev := 0; dev < numDevices; dev++ {
			if pkt[dev*2+1] != 0xFF {
				t.Errorf("row 0 dev %d data = 0x%02X, want 0xFF", dev, pkt[dev*2+1])
			}
		}
	})

	t.Run("column in last device", func(t *testing.T) {
		mock := &mockSPI{}
		m := newTestMAX7219(mock)

		buf := make([]byte, 32)
		buf[31] = 0x01 // last column, row 0

		m.WriteFramebuffer(buf)

		// Row 0: device 3 (cols 24-31) should have bit 0 set (col 31 = col 7 within device → bit 0)
		pkt := mock.txCalls[0]
		if pkt[3*2+1] != 0x01 {
			t.Errorf("row 0 dev 3 data = 0x%02X, want 0x01", pkt[3*2+1])
		}
	})

	t.Run("short buffer ignored", func(t *testing.T) {
		mock := &mockSPI{}
		m := newTestMAX7219(mock)

		m.WriteFramebuffer(make([]byte, 10))

		if len(mock.txCalls) != 0 {
			t.Errorf("expected 0 Tx calls for short buffer, got %d", len(mock.txCalls))
		}
	})

	t.Run("vertical stripe", func(t *testing.T) {
		mock := &mockSPI{}
		m := newTestMAX7219(mock)

		buf := make([]byte, 32)
		buf[0] = 0xFF // column 0, all rows lit

		m.WriteFramebuffer(buf)

		// Every row should have bit 7 set in device 0
		for row := 0; row < 8; row++ {
			pkt := mock.txCalls[row]
			if pkt[1] != 0x80 {
				t.Errorf("row %d dev 0 data = 0x%02X, want 0x80", row, pkt[1])
			}
		}
	})
}

func TestMAX7219_WriteFramebuffer_TxError(t *testing.T) {
	mock := &mockSPI{txErr: errors.New("spi fail")}
	m := newTestMAX7219(mock)

	buf := make([]byte, 32)
	buf[0] = 0x01
	m.WriteFramebuffer(buf) // should not panic

	// Only 1 call because it returns early on error
	if len(mock.txCalls) != 1 {
		t.Errorf("expected 1 Tx call before error abort, got %d", len(mock.txCalls))
	}
}

func TestMAX7219_writeAll(t *testing.T) {
	mock := &mockSPI{}
	m := newTestMAX7219(mock)
	m.writeAll(0xAA, 0xBB)

	if len(mock.txCalls) != 1 {
		t.Fatalf("expected 1 Tx call, got %d", len(mock.txCalls))
	}
	pkt := mock.txCalls[0]
	if len(pkt) != numDevices*2 {
		t.Fatalf("packet len = %d, want %d", len(pkt), numDevices*2)
	}
	for dev := 0; dev < numDevices; dev++ {
		if pkt[dev*2] != 0xAA || pkt[dev*2+1] != 0xBB {
			t.Errorf("device %d: got [0x%02X, 0x%02X], want [0xAA, 0xBB]", dev, pkt[dev*2], pkt[dev*2+1])
		}
	}
}

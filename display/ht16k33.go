package display

import (
	"encoding/binary"
	"fmt"

	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"
)

const (
	ht16k33CmdOscillator = 0x21
	ht16k33CmdDisplayOn  = 0x81
	ht16k33CmdBrightness = 0xE0
)

// HT16K33 drives a 4-digit 14-segment display over I2C.
type HT16K33 struct {
	busName string
	addr    uint16
	layout  string
	bus     i2c.BusCloser
	dev     *i2c.Dev
	prev    [8]byte
}

// NewHT16K33 creates a new HT16K33 display.
// layout is "sequential" (default) or "adafruit".
func NewHT16K33(busName string, addr uint16, layout string) *HT16K33 {
	if layout == "" {
		layout = "sequential"
	}
	return &HT16K33{busName: busName, addr: addr, layout: layout}
}

func (h *HT16K33) Init() error {
	if _, err := host.Init(); err != nil {
		return fmt.Errorf("periph init: %w", err)
	}
	bus, err := i2creg.Open(h.busName)
	if err != nil {
		return fmt.Errorf("i2c open %q: %w", h.busName, err)
	}
	h.bus = bus
	h.dev = &i2c.Dev{Bus: bus, Addr: h.addr}

	// Turn on oscillator
	if err := h.dev.Tx([]byte{ht16k33CmdOscillator}, nil); err != nil {
		return fmt.Errorf("ht16k33 oscillator: %w", err)
	}
	// Display on, no blinking
	if err := h.dev.Tx([]byte{ht16k33CmdDisplayOn}, nil); err != nil {
		return fmt.Errorf("ht16k33 display on: %w", err)
	}
	// Default brightness
	if err := h.dev.Tx([]byte{ht16k33CmdBrightness | 0x0F}, nil); err != nil {
		return fmt.Errorf("ht16k33 brightness: %w", err)
	}

	h.Clear()
	return nil
}

func (h *HT16K33) Close() error {
	h.Clear()
	if h.bus != nil {
		return h.bus.Close()
	}
	return nil
}

func (h *HT16K33) Clear() {
	h.WriteSegments([]uint16{0, 0, 0, 0}, false)
}

func (h *HT16K33) DisplayLength() int { return 4 }

func (h *HT16K33) SetBrightness(level byte) {
	if level > 15 {
		level = 15
	}
	h.dev.Tx([]byte{ht16k33CmdBrightness | level}, nil)
}

func (h *HT16K33) WriteSegments(segments []uint16, colon bool) {
	var buf [8]byte // 4 digits × 2 bytes each (little-endian uint16)

	for i := 0; i < 4 && i < len(segments); i++ {
		seg := segments[i]
		if i == 1 && colon {
			seg |= 0x4000 // colon bit on digit 1
		}
		pos := h.digitPosition(i)
		binary.LittleEndian.PutUint16(buf[pos*2:], seg)
	}

	// Differential update: only write changed bytes
	if buf != h.prev {
		// Write the full display buffer starting at address 0
		data := make([]byte, 9)
		data[0] = 0x00 // starting register address
		copy(data[1:], buf[:])
		h.dev.Tx(data, nil)
		h.prev = buf
	}
}

// digitPosition maps logical digit index to physical buffer position.
func (h *HT16K33) digitPosition(digit int) int {
	if h.layout == "adafruit" {
		// Adafruit 14-seg backpack layout: 0, 1, (colon at 2), 3, 4
		// Maps logical digits 0-3 to positions 0, 1, 3, 4 (skipping colon position 2)
		switch digit {
		case 0:
			return 0
		case 1:
			return 1
		case 2:
			return 3
		case 3:
			return 4 // use position 3 since we only have 4 uint16 slots
		}
	}
	return digit
}

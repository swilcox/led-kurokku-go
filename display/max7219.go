package display

import (
	"github.com/swilcox/led-kurokku-go/spi"
)

// MAX7219 registers
const (
	regNoop        = 0x00
	regDecodeMode  = 0x09
	regIntensity   = 0x0A
	regScanLimit   = 0x0B
	regShutdown    = 0x0C
	regDisplayTest = 0x0F
)

const numDevices = 4 // 4-in-1 module

// MAX7219 drives a 4-in-1 MAX7219 32x8 LED matrix over SPI.
type MAX7219 struct {
	dev *spi.Device
}

// NewMAX7219 creates a new MAX7219 display using the given SPI bus name.
// Pass "" for the default bus.
func NewMAX7219(spiBus string) *MAX7219 {
	return &MAX7219{}
}

// Init opens the SPI bus and configures all four MAX7219 chips.
func (m *MAX7219) Init() error {
	dev, err := spi.Open("", 10_000_000)
	if err != nil {
		return err
	}
	m.dev = dev

	// Initialize all devices
	m.writeAll(regDisplayTest, 0x00) // normal operation
	m.writeAll(regDecodeMode, 0x00)  // no BCD decode, raw segments
	m.writeAll(regScanLimit, 0x07)   // scan all 8 digits (rows)
	m.writeAll(regShutdown, 0x01)    // normal operation (not shutdown)
	m.writeAll(regIntensity, 0x04)   // moderate brightness

	m.Clear()
	return nil
}

// Close shuts down the displays and releases the SPI bus.
func (m *MAX7219) Close() error {
	m.writeAll(regShutdown, 0x00)
	if m.dev != nil {
		return m.dev.Close()
	}
	return nil
}

// Clear blanks all pixels.
func (m *MAX7219) Clear() {
	for row := byte(1); row <= 8; row++ {
		m.writeAll(row, 0x00)
	}
}

// SetBrightness sets brightness across all devices (0-15).
func (m *MAX7219) SetBrightness(level byte) {
	if level > 15 {
		level = 15
	}
	m.writeAll(regIntensity, level)
}

// WriteFramebuffer writes a 32x8 framebuffer to the display.
// buf must be 32 bytes: 32 columns of 8 vertical bits each (LSB = top row).
// The 4-in-1 module is wired so device 0 is the rightmost 8 columns.
func (m *MAX7219) WriteFramebuffer(buf []byte) {
	if len(buf) < 32 {
		return
	}

	for row := byte(0); row < 8; row++ {
		// Build a packet: 2 bytes per device (register, data), devices in daisy-chain order.
		// Device 0 (rightmost) gets the last byte in the chain.
		packet := make([]byte, numDevices*2)
		for dev := 0; dev < numDevices; dev++ {
			// First bytes in the packet shift to the last device in the chain.
			// dev 0 (columns 0-7) → first in packet → farthest device (right side).
			idx := dev * 2
			packet[idx] = row + 1 // MAX7219 row registers are 1-indexed

			// Pack 8 columns into one byte for this device's row
			var rowByte byte
			baseCol := dev * 8
			for col := 0; col < 8; col++ {
				if buf[baseCol+col]&(1<<row) != 0 {
					rowByte |= 1 << (7 - col)
				}
			}
			packet[idx+1] = rowByte
		}
		m.dev.Tx(packet)
	}
}

// Width returns the pixel width.
func (m *MAX7219) Width() int { return 32 }

// Height returns the pixel height.
func (m *MAX7219) Height() int { return 8 }

// writeAll sends the same register/value pair to all daisy-chained devices.
func (m *MAX7219) writeAll(reg, value byte) {
	packet := make([]byte, numDevices*2)
	for i := 0; i < numDevices; i++ {
		packet[i*2] = reg
		packet[i*2+1] = value
	}
	m.dev.Tx(packet)
}

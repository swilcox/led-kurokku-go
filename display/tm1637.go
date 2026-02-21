package display

import (
	"fmt"
	"time"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/host/v3"
)

const (
	tm1637AddrAuto  = 0x40
	tm1637StartAddr = 0xC0
	tm1637Brightness = 0x88
)

// TM1637 drives a 4-digit 7-segment display via bit-bang protocol.
type TM1637 struct {
	clkName string
	dioName string
	clk     gpio.PinIO
	dio     gpio.PinIO
	bright  byte
}

// NewTM1637 creates a new TM1637 display using the given GPIO pin names.
func NewTM1637(clkPin, dioPin string) *TM1637 {
	return &TM1637{clkName: clkPin, dioName: dioPin, bright: 7}
}

func (t *TM1637) Init() error {
	if _, err := host.Init(); err != nil {
		return fmt.Errorf("periph init: %w", err)
	}
	t.clk = gpioreg.ByName(t.clkName)
	if t.clk == nil {
		return fmt.Errorf("GPIO pin %q not found", t.clkName)
	}
	t.dio = gpioreg.ByName(t.dioName)
	if t.dio == nil {
		return fmt.Errorf("GPIO pin %q not found", t.dioName)
	}
	t.clk.Out(gpio.High)
	t.dio.Out(gpio.High)
	return nil
}

func (t *TM1637) Close() error {
	t.clk.Out(gpio.Low)
	t.dio.Out(gpio.Low)
	return nil
}

func (t *TM1637) Clear() {
	t.WriteSegments([]uint16{0, 0, 0, 0}, false)
}

func (t *TM1637) DisplayLength() int { return 4 }

func (t *TM1637) SetBrightness(level byte) {
	// TM1637 brightness range is 0-7
	b := level >> 1 // map 0-15 → 0-7
	if b > 7 {
		b = 7
	}
	t.bright = b
	// Send brightness command
	t.start()
	t.writeByte(tm1637Brightness | b)
	t.stop()
}

func (t *TM1637) WriteSegments(segments []uint16, colon bool) {
	// Auto-increment mode
	t.start()
	t.writeByte(tm1637AddrAuto)
	t.stop()

	// Write data starting at address 0
	t.start()
	t.writeByte(tm1637StartAddr)
	for i := 0; i < 4 && i < len(segments); i++ {
		b := byte(segments[i] & 0xFF) // TM1637 uses low byte only
		if i == 1 && colon {
			b |= 0x80 // colon is bit 7 of digit 1
		}
		t.writeByte(b)
	}
	t.stop()

	// Brightness
	t.start()
	t.writeByte(tm1637Brightness | t.bright)
	t.stop()
}

func (t *TM1637) start() {
	t.dio.Out(gpio.High)
	t.clk.Out(gpio.High)
	t.bitDelay()
	t.dio.Out(gpio.Low)
	t.bitDelay()
	t.clk.Out(gpio.Low)
	t.bitDelay()
}

func (t *TM1637) stop() {
	t.clk.Out(gpio.Low)
	t.bitDelay()
	t.dio.Out(gpio.Low)
	t.bitDelay()
	t.clk.Out(gpio.High)
	t.bitDelay()
	t.dio.Out(gpio.High)
	t.bitDelay()
}

func (t *TM1637) writeByte(b byte) {
	for i := 0; i < 8; i++ {
		t.clk.Out(gpio.Low)
		t.bitDelay()
		if b&(1<<uint(i)) != 0 {
			t.dio.Out(gpio.High)
		} else {
			t.dio.Out(gpio.Low)
		}
		t.bitDelay()
		t.clk.Out(gpio.High)
		t.bitDelay()
	}

	// ACK
	t.clk.Out(gpio.Low)
	t.bitDelay()
	t.dio.Out(gpio.High) // release data line
	t.bitDelay()
	t.clk.Out(gpio.High)
	t.bitDelay()
	t.clk.Out(gpio.Low)
	t.bitDelay()
}

func (t *TM1637) bitDelay() {
	time.Sleep(time.Microsecond)
}

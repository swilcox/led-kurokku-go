package display

import (
	"fmt"
	"io"
	"strings"
)

// Terminal renders a 32x8 framebuffer to the terminal using block characters.
type Terminal struct {
	w      io.Writer
	width  int
	height int
}

// NewTerminal creates a terminal display writing to w with standard 32x8 dimensions.
func NewTerminal(w io.Writer) *Terminal {
	return &Terminal{w: w, width: 32, height: 8}
}

func (t *Terminal) Init() error          { return nil }
func (t *Terminal) Close() error         { return nil }
func (t *Terminal) Width() int           { return t.width }
func (t *Terminal) Height() int          { return t.height }
func (t *Terminal) SetBrightness(_ byte) {}

// Clear clears the terminal screen.
func (t *Terminal) Clear() {
	fmt.Fprint(t.w, "\033[2J\033[H")
}

// WriteFramebuffer renders buf to the terminal.
// buf is 32 bytes: one byte per column, 8 vertical bits (LSB = top row).
func (t *Terminal) WriteFramebuffer(buf []byte) {
	fmt.Fprint(t.w, "\033[H") // cursor home
	border := "+" + strings.Repeat("-", t.width) + "+"
	fmt.Fprintln(t.w, border)
	for row := 0; row < t.height; row++ {
		fmt.Fprint(t.w, "|")
		for col := 0; col < t.width && col < len(buf); col++ {
			if buf[col]&(1<<row) != 0 {
				fmt.Fprint(t.w, "█")
			} else {
				fmt.Fprint(t.w, " ")
			}
		}
		fmt.Fprintln(t.w, "|")
	}
	fmt.Fprintln(t.w, border)
}

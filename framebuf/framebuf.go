package framebuf

import "github.com/swilcox/led-kurokku-go/font"

// Frame represents a 32x8 LED matrix framebuffer.
// Each byte is one vertical column (bit 0 = top row, bit 7 = bottom row).
type Frame [32]byte

// Clear zeroes all pixels.
func (f *Frame) Clear() {
	*f = Frame{}
}

// SetPixel sets or clears the pixel at (x, y). x: 0–31, y: 0–7.
func (f *Frame) SetPixel(x, y int, on bool) {
	if x < 0 || x >= 32 || y < 0 || y >= 8 {
		return
	}
	if on {
		f[x] |= 1 << uint(y)
	} else {
		f[x] &^= 1 << uint(y)
	}
}

// GetPixel returns whether the pixel at (x, y) is lit.
func (f *Frame) GetPixel(x, y int) bool {
	if x < 0 || x >= 32 || y < 0 || y >= 8 {
		return false
	}
	return f[x]&(1<<uint(y)) != 0
}

// Bytes returns the frame data as a byte slice.
func (f *Frame) Bytes() []byte {
	return f[:]
}

// BlitText renders text into the frame at the given horizontal offset.
// Returns the total pixel width of the rendered text.
func BlitText(f *Frame, text string, offsetX int) int {
	cols := font.RenderText(text)
	for i, col := range cols {
		x := offsetX + i
		if x >= 0 && x < 32 {
			f[x] = col
		}
	}
	return len(cols)
}

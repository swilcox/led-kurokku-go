package display

// Display is the base interface that all display backends implement.
type Display interface {
	Init() error
	Close() error
	Clear()
	SetBrightness(level byte) // 0-15 for MAX7219, mapped for others
}

// PixelDisplay is for pixel matrix displays (MAX7219, terminal).
type PixelDisplay interface {
	Display
	WriteFramebuffer(buf []byte)
	Width() int
	Height() int
}

// SegmentDisplay is for segment displays (TM1637, HT16K33).
type SegmentDisplay interface {
	Display
	WriteSegments(segments []uint16, colon bool)
	DisplayLength() int // typically 4
}

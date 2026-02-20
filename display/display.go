package display

// Display is the interface that all display backends must implement.
type Display interface {
	Init() error
	Close() error
	Clear()
	SetBrightness(level byte) // 0-15 for MAX7219, mapped for others
	WriteFramebuffer(buf []byte)
	Width() int
	Height() int
}

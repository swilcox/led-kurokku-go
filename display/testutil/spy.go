package testutil

import "github.com/swilcox/led-kurokku-go/display"

var _ display.Display = (*SpyDisplay)(nil)

// SpyDisplay records display calls for use in tests.
type SpyDisplay struct {
	Frames     [][]byte
	Brightness []byte
	ClearCalls int
}

func (s *SpyDisplay) Init() error          { return nil }
func (s *SpyDisplay) Close() error         { return nil }
func (s *SpyDisplay) Width() int           { return 32 }
func (s *SpyDisplay) Height() int          { return 8 }

func (s *SpyDisplay) Clear() {
	s.ClearCalls++
}

func (s *SpyDisplay) SetBrightness(level byte) {
	s.Brightness = append(s.Brightness, level)
}

func (s *SpyDisplay) WriteFramebuffer(buf []byte) {
	frame := make([]byte, len(buf))
	copy(frame, buf)
	s.Frames = append(s.Frames, frame)
}

package testutil

import "github.com/swilcox/led-kurokku-go/display"

var _ display.PixelDisplay = (*SpyDisplay)(nil)

// SpyDisplay records pixel display calls for use in tests.
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

var _ display.SegmentDisplay = (*SpySegmentDisplay)(nil)

// SegmentCall records a single WriteSegments call.
type SegmentCall struct {
	Segments []uint16
	Colon    bool
}

// SpySegmentDisplay records segment display calls for use in tests.
type SpySegmentDisplay struct {
	Calls      []SegmentCall
	Brightness []byte
	ClearCalls int
	Length     int // defaults to 4
}

func (s *SpySegmentDisplay) Init() error  { return nil }
func (s *SpySegmentDisplay) Close() error { return nil }

func (s *SpySegmentDisplay) Clear() {
	s.ClearCalls++
}

func (s *SpySegmentDisplay) SetBrightness(level byte) {
	s.Brightness = append(s.Brightness, level)
}

func (s *SpySegmentDisplay) DisplayLength() int {
	if s.Length > 0 {
		return s.Length
	}
	return 4
}

func (s *SpySegmentDisplay) WriteSegments(segments []uint16, colon bool) {
	seg := make([]uint16, len(segments))
	copy(seg, segments)
	s.Calls = append(s.Calls, SegmentCall{Segments: seg, Colon: colon})
}

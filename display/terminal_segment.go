package display

import (
	"fmt"
	"io"
	"strings"
)

// SegmentType selects between 7-segment and 14-segment rendering.
type SegmentType int

const (
	Segment7 SegmentType = iota
	Segment14
)

// TerminalSegment renders a segment display to the terminal using ASCII art.
type TerminalSegment struct {
	w       io.Writer
	segType SegmentType
	length  int
}

// NewTerminalSegment creates a terminal segment display writing to w.
func NewTerminalSegment(w io.Writer, segType SegmentType) *TerminalSegment {
	return &TerminalSegment{w: w, segType: segType, length: 4}
}

func (t *TerminalSegment) Init() error          { return nil }
func (t *TerminalSegment) Close() error         { return nil }
func (t *TerminalSegment) SetBrightness(_ byte) {}
func (t *TerminalSegment) DisplayLength() int   { return t.length }

func (t *TerminalSegment) Clear() {
	fmt.Fprint(t.w, "\033[2J\033[H")
}

func (t *TerminalSegment) WriteSegments(segments []uint16, colon bool) {
	fmt.Fprint(t.w, "\033[H") // cursor home
	if t.segType == Segment7 {
		t.render7(segments, colon)
	} else {
		t.render14(segments, colon)
	}
}

// render7 draws 7-segment digits as 5-line ASCII art.
//
// bit 0 = a (top), bit 1 = b (upper-right), bit 2 = c (lower-right),
// bit 3 = d (bottom), bit 4 = e (lower-left), bit 5 = f (upper-left),
// bit 6 = g (middle)
//
//  ___
// |   |
// |___|
// |   |
// |___|
func (t *TerminalSegment) render7(segments []uint16, colon bool) {
	var lines [5]string

	for i, seg := range segments {
		b := byte(seg)
		a := b&0x01 != 0
		bSeg := b&0x02 != 0
		c := b&0x04 != 0
		d := b&0x08 != 0
		e := b&0x10 != 0
		f := b&0x20 != 0
		g := b&0x40 != 0

		// Top bar
		if a {
			lines[0] += " ___ "
		} else {
			lines[0] += "     "
		}

		// Upper sides
		left := ch(f, '|')
		right := ch(bSeg, '|')
		lines[1] += left + "   " + right

		// Middle bar
		leftM := ch(f, '|')
		rightM := ch(bSeg, '|')
		if g {
			lines[2] += leftM + "___" + rightM
		} else {
			lines[2] += leftM + "   " + rightM
		}

		// Lower sides
		leftL := ch(e, '|')
		rightL := ch(c, '|')
		lines[3] += leftL + "   " + rightL

		// Bottom bar
		leftB := ch(e, '|')
		rightB := ch(c, '|')
		if d {
			lines[4] += leftB + "___" + rightB
		} else {
			lines[4] += leftB + "   " + rightB
		}

		// Colon after digit 1
		if i == 1 && colon {
			lines[0] += " "
			lines[1] += "o"
			lines[2] += " "
			lines[3] += "o"
			lines[4] += " "
		} else if i < len(segments)-1 {
			lines[0] += " "
			lines[1] += " "
			lines[2] += " "
			lines[3] += " "
			lines[4] += " "
		}
	}

	border := "+" + strings.Repeat("-", len(lines[0])) + "+"
	fmt.Fprintln(t.w, border)
	for _, line := range lines {
		fmt.Fprintf(t.w, "|%s|\n", line)
	}
	fmt.Fprintln(t.w, border)
}

// render14 draws 14-segment digits as 7-line ASCII art.
//
// Bit layout matching Python HT16K33 driver:
// bit  0 = A   (top horizontal)
// bit  1 = B   (top right vertical)
// bit  2 = C   (bottom right vertical)
// bit  3 = D   (bottom horizontal)
// bit  4 = E   (bottom left vertical)
// bit  5 = F   (top left vertical)
// bit  6 = G1  (middle left horizontal)
// bit  7 = G2  (middle right horizontal)
// bit  8 = H   (top left diagonal)
// bit  9 = I   (top center vertical)
// bit 10 = J   (top right diagonal)
// bit 11 = K   (bottom left diagonal)
// bit 12 = L   (bottom center vertical)
// bit 13 = M   (bottom right diagonal)
//
//    ___
//   |\|/|
//    - -
//   |/|\|
//    ---
func (t *TerminalSegment) render14(segments []uint16, colon bool) {
	var lines [7]string

	for i, seg := range segments {
		a := seg&0x0001 != 0
		b := seg&0x0002 != 0
		c := seg&0x0004 != 0
		d := seg&0x0008 != 0
		e := seg&0x0010 != 0
		f := seg&0x0020 != 0
		g1 := seg&0x0040 != 0
		g2 := seg&0x0080 != 0
		h := seg&0x0100 != 0
		iSeg := seg&0x0200 != 0
		j := seg&0x0400 != 0
		k := seg&0x0800 != 0
		l := seg&0x1000 != 0
		m := seg&0x2000 != 0

		// Top bar
		if a {
			lines[0] += " _____ "
		} else {
			lines[0] += "       "
		}

		// Upper half - top row
		fCh := ch(f, '|')
		hCh := ch14(h, '\\')
		iCh := ch(iSeg, '|')
		jCh := ch14(j, '/')
		bCh := ch(b, '|')
		lines[1] += fCh + hCh + "  " + iCh + "  " + jCh + bCh

		// Upper half - bottom row
		lines[2] += fCh + " " + hCh + " " + iCh + " " + jCh + " " + bCh

		// Middle bar
		g1Ch := "  "
		g2Ch := "  "
		if g1 {
			g1Ch = "__"
		}
		if g2 {
			g2Ch = "__"
		}
		lines[3] += " " + g1Ch + " " + g2Ch + " "

		// Lower half - top row
		eCh := ch(e, '|')
		cCh := ch(c, '|')
		kCh := ch14(k, '/')
		lCh := ch(l, '|')
		mCh := ch14(m, '\\')
		lines[4] += eCh + " " + kCh + " " + lCh + " " + mCh + " " + cCh

		// Lower half - bottom row
		lines[5] += eCh + kCh + "  " + lCh + "  " + mCh + cCh

		// Bottom bar
		if d {
			lines[6] += " _____ "
		} else {
			lines[6] += "       "
		}

		// Colon after digit 1
		if i == 1 && colon {
			lines[0] += " "
			lines[1] += " "
			lines[2] += " "
			lines[3] += "o"
			lines[4] += " "
			lines[5] += "o"
			lines[6] += " "
		} else if i < len(segments)-1 {
			for k := range lines {
				lines[k] += " "
			}
		}
	}

	border := "+" + strings.Repeat("-", len(lines[0])) + "+"
	fmt.Fprintln(t.w, border)
	for _, line := range lines {
		fmt.Fprintf(t.w, "|%s|\n", line)
	}
	fmt.Fprintln(t.w, border)
}

func ch(on bool, c byte) string {
	if on {
		return string(c)
	}
	return " "
}

func ch14(on bool, c byte) string {
	if on {
		return string(c)
	}
	return " "
}

package display

import (
	"bytes"
	"strings"
	"testing"
)

func TestTerminalSegment7_Init(t *testing.T) {
	var buf bytes.Buffer
	ts := NewTerminalSegment(&buf, Segment7)
	if err := ts.Init(); err != nil {
		t.Fatal(err)
	}
	if ts.DisplayLength() != 4 {
		t.Errorf("DisplayLength() = %d, want 4", ts.DisplayLength())
	}
}

func TestTerminalSegment7_WriteSegments(t *testing.T) {
	var buf bytes.Buffer
	ts := NewTerminalSegment(&buf, Segment7)

	// Display "1" on first digit (segments b, c = 0x06)
	segments := []uint16{0x06, 0x00, 0x00, 0x00}
	ts.WriteSegments(segments, false)

	output := buf.String()
	if !strings.Contains(output, "|") {
		t.Error("expected segment output to contain pipe characters")
	}
	// Should have border lines
	if !strings.Contains(output, "+") {
		t.Error("expected border characters in output")
	}
}

func TestTerminalSegment7_WithColon(t *testing.T) {
	var buf bytes.Buffer
	ts := NewTerminalSegment(&buf, Segment7)

	// Display "12:30" → digits 1, 2, 3, 0 with colon
	segments := []uint16{0x06, 0x5B, 0x4F, 0x3F}
	ts.WriteSegments(segments, true)

	output := buf.String()
	if !strings.Contains(output, "o") {
		t.Error("expected colon 'o' characters in output")
	}
}

func TestTerminalSegment14_WriteSegments(t *testing.T) {
	var buf bytes.Buffer
	ts := NewTerminalSegment(&buf, Segment14)

	// Display some non-zero segments
	segments := []uint16{0x03CF, 0x4A2F, 0x00F3, 0x482F}
	ts.WriteSegments(segments, false)

	output := buf.String()
	if !strings.Contains(output, "+") {
		t.Error("expected border characters in output")
	}
	// 14-seg output should have 7 content lines + 2 borders = 9 lines minimum
	lines := strings.Split(strings.TrimSpace(output), "\n")
	// Remove ANSI escape at start
	filteredLines := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			filteredLines++
		}
	}
	if filteredLines < 9 {
		t.Errorf("expected at least 9 lines for 14-seg display, got %d", filteredLines)
	}
}

func TestTerminalSegment14_WithColon(t *testing.T) {
	var buf bytes.Buffer
	ts := NewTerminalSegment(&buf, Segment14)

	segments := []uint16{0x0001, 0x0001, 0x0001, 0x0001}
	ts.WriteSegments(segments, true)

	output := buf.String()
	if !strings.Contains(output, "o") {
		t.Error("expected colon 'o' characters in output")
	}
}

func TestTerminalSegment_Clear(t *testing.T) {
	var buf bytes.Buffer
	ts := NewTerminalSegment(&buf, Segment7)
	ts.Clear()
	output := buf.String()
	if !strings.Contains(output, "\033[2J") {
		t.Error("expected ANSI clear sequence")
	}
}

func TestTerminalSegment_SetBrightness(t *testing.T) {
	var buf bytes.Buffer
	ts := NewTerminalSegment(&buf, Segment7)
	ts.SetBrightness(10) // should be a no-op, no panic
}

func TestTerminalSegment_Close(t *testing.T) {
	var buf bytes.Buffer
	ts := NewTerminalSegment(&buf, Segment7)
	if err := ts.Close(); err != nil {
		t.Errorf("Close() returned error: %v", err)
	}
}

func TestTerminalSegment7_AllSegmentsLit(t *testing.T) {
	var buf bytes.Buffer
	ts := NewTerminalSegment(&buf, Segment7)
	// All 7 segments lit: 0x7F
	segments := []uint16{0x7F, 0x7F, 0x7F, 0x7F}
	ts.WriteSegments(segments, false)
	output := buf.String()
	// All segments lit should produce underscores (bars) and pipes (sides)
	if !strings.Contains(output, "___") {
		t.Error("expected bars for all-lit segments")
	}
	if !strings.Contains(output, "|") {
		t.Error("expected pipes for all-lit segments")
	}
}

func TestTerminalSegment7_NoSegmentsLit(t *testing.T) {
	var buf bytes.Buffer
	ts := NewTerminalSegment(&buf, Segment7)
	segments := []uint16{0x00, 0x00, 0x00, 0x00}
	ts.WriteSegments(segments, false)
	output := buf.String()
	// Should still have border but no segment characters
	if !strings.Contains(output, "+") {
		t.Error("expected border characters")
	}
}

func TestTerminalSegment7_ColonOff(t *testing.T) {
	var buf bytes.Buffer
	ts := NewTerminalSegment(&buf, Segment7)
	segments := []uint16{0x06, 0x5B, 0x4F, 0x3F}
	ts.WriteSegments(segments, false)
	output := buf.String()
	if strings.Contains(output, "o") {
		t.Error("expected no colon when colon=false")
	}
}

func TestTerminalSegment14_AllSegmentsLit(t *testing.T) {
	var buf bytes.Buffer
	ts := NewTerminalSegment(&buf, Segment14)
	// All 14 segments lit: 0x3FFF
	segments := []uint16{0x3FFF, 0x3FFF, 0x3FFF, 0x3FFF}
	ts.WriteSegments(segments, true)
	output := buf.String()
	if !strings.Contains(output, "\\") {
		t.Error("expected backslash for diagonal segments")
	}
	if !strings.Contains(output, "/") {
		t.Error("expected slash for diagonal segments")
	}
	if !strings.Contains(output, "o") {
		t.Error("expected colon dots")
	}
}

func TestTerminalSegment14_Diagonals(t *testing.T) {
	var buf bytes.Buffer
	ts := NewTerminalSegment(&buf, Segment14)
	// H (bit 8) = top-left diagonal, J (bit 10) = top-right diagonal
	segments := []uint16{0x0500}
	ts.WriteSegments(segments, false)
	output := buf.String()
	if !strings.Contains(output, "\\") {
		t.Error("expected backslash for H segment")
	}
	if !strings.Contains(output, "/") {
		t.Error("expected slash for J segment")
	}
}

func TestTerminalSegment_CursorHome(t *testing.T) {
	var buf bytes.Buffer
	ts := NewTerminalSegment(&buf, Segment7)
	segments := []uint16{0x00}
	ts.WriteSegments(segments, false)
	output := buf.String()
	if !strings.Contains(output, "\033[H") {
		t.Error("expected cursor-home ANSI sequence")
	}
}

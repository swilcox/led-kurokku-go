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

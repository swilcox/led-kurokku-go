package display_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/swilcox/led-kurokku-go/display"
)

func TestTerminal_Dimensions(t *testing.T) {
	var buf bytes.Buffer
	d := display.NewTerminal(&buf)
	if d.Width() != 32 {
		t.Errorf("Width: got %d, want 32", d.Width())
	}
	if d.Height() != 8 {
		t.Errorf("Height: got %d, want 8", d.Height())
	}
}

func TestTerminal_WriteFramebuffer_Borders(t *testing.T) {
	var buf bytes.Buffer
	d := display.NewTerminal(&buf)
	frame := make([]byte, 32)
	d.WriteFramebuffer(frame)

	output := buf.String()
	if !strings.Contains(output, "+") {
		t.Error("expected '+' corner characters in output")
	}
	if !strings.Contains(output, "|") {
		t.Error("expected '|' border characters in output")
	}
	if !strings.Contains(output, "-") {
		t.Error("expected '-' border characters in output")
	}
}

func TestTerminal_WriteFramebuffer_LitPixel(t *testing.T) {
	var buf bytes.Buffer
	d := display.NewTerminal(&buf)
	frame := make([]byte, 32)
	frame[0] = 0x01 // top-left pixel on (bit 0 = row 0)
	d.WriteFramebuffer(frame)

	output := buf.String()
	if !strings.Contains(output, "█") {
		t.Error("expected block character '█' for a lit pixel")
	}
}

func TestTerminal_WriteFramebuffer_AllOff(t *testing.T) {
	var buf bytes.Buffer
	d := display.NewTerminal(&buf)
	frame := make([]byte, 32) // all zeros
	d.WriteFramebuffer(frame)

	output := buf.String()
	if strings.Contains(output, "█") {
		t.Error("expected no lit pixels for an all-zero framebuffer")
	}
}

func TestTerminal_Init_Close(t *testing.T) {
	var buf bytes.Buffer
	d := display.NewTerminal(&buf)
	if err := d.Init(); err != nil {
		t.Errorf("Init: unexpected error: %v", err)
	}
	if err := d.Close(); err != nil {
		t.Errorf("Close: unexpected error: %v", err)
	}
}

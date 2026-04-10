package admin

import (
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/swilcox/led-kurokku-go/config"
	"github.com/swilcox/led-kurokku-go/framebuf"
	"github.com/swilcox/led-kurokku-go/segfont"
)

// enabledWidgets returns the config indices of enabled widgets.
func enabledWidgets(cfg *config.Config) []int {
	var indices []int
	for i, w := range cfg.Widgets {
		if w.Enabled {
			indices = append(indices, i)
		}
	}
	return indices
}

// previewDuration returns the number of seconds to display a widget preview.
// Uses 5s default for "0s" (run-to-completion) durations, with a 3s minimum.
func previewDuration(w config.WidgetConfig) int {
	secs := int(time.Duration(w.Duration).Seconds())
	if secs == 0 {
		return 5
	}
	if secs < 3 {
		return 3
	}
	return secs
}

// previewWidgetHTML renders a preview for a specific widget by config index.
// Returns the rendered HTML and a label string (e.g. "clock").
func previewWidgetHTML(cfg *config.Config, host string, port int, widgetIdx int) (template.HTML, string) {
	if widgetIdx < 0 || widgetIdx >= len(cfg.Widgets) {
		return "", ""
	}
	w := cfg.Widgets[widgetIdx]
	text := previewTextForWidget(w, cfg, host, port)

	if cfg.Display.IsSegment() {
		enc := segfont.Enc14
		is7 := cfg.Display.Type == config.DisplayTM1637 || cfg.Display.Type == config.DisplayTerminalSeg7
		if is7 {
			enc = segfont.Enc7
		}
		digits, colon := segmentTextForWidget(w, text)
		return segmentPreviewHTML(digits, enc, is7, colon), w.Type
	}
	return pixelPreviewHTML(text), w.Type
}

// previewTextForWidget returns the display text for a given widget.
func previewTextForWidget(w config.WidgetConfig, cfg *config.Config, host string, port int) string {
	switch w.Type {
	case "clock":
		now := time.Now()
		if w.Format24h != nil && *w.Format24h {
			return now.Format("15:04")
		}
		h := now.Hour() % 12
		if h == 0 {
			h = 12
		}
		return fmt.Sprintf("%2d:%02d", h, now.Minute())
	case "message":
		if w.DynamicSource != "" {
			val, err := FetchKey(host, port, w.DynamicSource)
			if err == nil && val != "" {
				return val
			}
			if w.Text != "" {
				return w.Text
			}
			return "[dynamic: " + w.DynamicSource + "]"
		}
		if w.Text != "" {
			return w.Text
		}
		return "[message]"
	case "alert":
		if w.Text != "" {
			return w.Text
		}
		return "[alert]"
	case "animation":
		atype := w.AnimationType
		if atype == "" {
			atype = "frames"
		}
		return "[animation: " + atype + "]"
	default:
		return "[" + w.Type + "]"
	}
}

// segmentTextForWidget converts preview text into segment digits and colon state.
func segmentTextForWidget(w config.WidgetConfig, text string) (digits string, colon bool) {
	if w.Type == "clock" {
		// Clock gets colon and no separator in digit string
		now := time.Now()
		if w.Format24h != nil && *w.Format24h {
			return now.Format("1504"), true
		}
		h := now.Hour() % 12
		if h == 0 {
			h = 12
		}
		return fmt.Sprintf("%2d%02d", h, now.Minute()), true
	}
	// Strip colons and fit to 4 chars
	t := strings.ReplaceAll(text, ":", "")
	if len(t) > 4 {
		t = t[:4]
	}
	for len(t) < 4 {
		t += " "
	}
	return t, false
}

// pixelPreviewHTML renders a 32x8 CSS grid of LED pixels.
func pixelPreviewHTML(text string) template.HTML {
	var f framebuf.Frame
	framebuf.BlitText(&f, text, 0)

	var b strings.Builder
	b.WriteString(`<div class="preview-pixel">`)
	for y := 0; y < 8; y++ {
		for x := 0; x < 32; x++ {
			cls := "off"
			if f.GetPixel(x, y) {
				cls = "on"
			}
			fmt.Fprintf(&b, `<div class="px %s"></div>`, cls)
		}
	}
	b.WriteString(`</div>`)
	return template.HTML(b.String())
}

// 7-segment polygon coordinates for one digit in a 30x52 viewBox.
// Bit order: 0=a, 1=b, 2=c, 3=d, 4=e, 5=f, 6=g.
var seg7Polys = [7]string{
	"7,1 23,1 21,5 9,5",                // a (bit 0) - top
	"24,3 27,6 27,23 24,26 22,23 22,6", // b (bit 1) - top right
	"24,27 27,30 27,47 24,50 22,47 22,30", // c (bit 2) - bottom right
	"9,48 21,48 23,52 7,52",            // d (bit 3) - bottom
	"3,30 6,27 8,30 8,47 6,50 3,47",    // e (bit 4) - bottom left
	"3,6 6,3 8,6 8,23 6,26 3,23",       // f (bit 5) - top left
	"7,25 9,23 21,23 23,25 21,27 9,27", // g (bit 6) - middle
}

// 14-segment polygon coordinates for one digit in a 30x52 viewBox.
// Bit order: 0=A, 1=B, 2=C, 3=D, 4=E, 5=F, 6=G1, 7=G2,
// 8=H, 9=I, 10=J, 11=K, 12=L, 13=M.
var seg14Polys = [14]string{
	"7,1 23,1 21,5 9,5",                   // A (bit 0) - top
	"24,3 27,6 27,23 24,26 22,23 22,6",    // B (bit 1) - top right
	"24,27 27,30 27,47 24,50 22,47 22,30", // C (bit 2) - bottom right
	"9,48 21,48 23,52 7,52",               // D (bit 3) - bottom
	"3,30 6,27 8,30 8,47 6,50 3,47",       // E (bit 4) - bottom left
	"3,6 6,3 8,6 8,23 6,26 3,23",          // F (bit 5) - top left
	"7,25 9,23 14,23 14,27 9,27",          // G1 (bit 6) - middle left
	"16,23 21,23 23,25 21,27 16,27",       // G2 (bit 7) - middle right
	"9,7 12,7 16,21 13,21",                // H (bit 8) - top-left diagonal
	"14,7 16,7 16,21 14,21",               // I (bit 9) - top center vertical
	"18,7 21,7 17,21 14,21",               // J (bit 10) - top-right diagonal
	"13,32 16,32 12,46 9,46",              // K (bit 11) - bottom-left diagonal
	"14,32 16,32 16,46 14,46",             // L (bit 12) - bottom center vertical
	"14,32 17,32 21,46 18,46",             // M (bit 13) - bottom-right diagonal
}

const (
	digitWidth   = 30
	digitHeight  = 53
	digitGap     = 8
	colonWidth   = 12
	litColor7    = "#ff3333"
	litColor14   = "#33ff33"
	unlitColor   = "#1c1c1c"
)

// segmentPreviewHTML renders an SVG-based segment display preview.
func segmentPreviewHTML(digits string, enc segfont.Encoder, is7 bool, colon bool) template.HTML {
	bitmasks := segfont.EncodeText(enc, digits)
	for len(bitmasks) < 4 {
		bitmasks = append(bitmasks, 0)
	}

	// Layout: [digit0][gap][digit1][gap/2][colon][gap/2][digit2][gap][digit3]
	totalWidth := 4*digitWidth + 2*digitGap + colonWidth + digitGap
	xOff := [4]int{
		0,
		digitWidth + digitGap,
		digitWidth + digitGap + digitWidth + digitGap/2 + colonWidth + digitGap/2,
		digitWidth + digitGap + digitWidth + digitGap/2 + colonWidth + digitGap/2 + digitWidth + digitGap,
	}
	colonX := digitWidth + digitGap + digitWidth + digitGap/2 + colonWidth/2

	litColor := litColor7
	if !is7 {
		litColor = litColor14
	}

	polys := seg7Polys[:]
	numSegs := 7
	if !is7 {
		polys = seg14Polys[:]
		numSegs = 14
	}

	var b strings.Builder
	fmt.Fprintf(&b, `<svg class="preview-segment" viewBox="0 0 %d %d" xmlns="http://www.w3.org/2000/svg">`, totalWidth, digitHeight)

	for d := 0; d < 4; d++ {
		mask := bitmasks[d]
		for s := 0; s < numSegs; s++ {
			fill := unlitColor
			if mask&(1<<uint(s)) != 0 {
				fill = litColor
			}
			points := offsetPolygon(polys[s], xOff[d], 0)
			fmt.Fprintf(&b, `<polygon points="%s" fill="%s"/>`, points, fill)
		}
	}

	// Colon dots
	if colon {
		fmt.Fprintf(&b, `<circle cx="%d" cy="17" r="3" fill="%s"/>`, colonX, litColor)
		fmt.Fprintf(&b, `<circle cx="%d" cy="36" r="3" fill="%s"/>`, colonX, litColor)
	}

	b.WriteString(`</svg>`)
	return template.HTML(b.String())
}

// offsetPolygon shifts all x,y coordinates in a polygon points string.
func offsetPolygon(points string, dx, dy int) string {
	pairs := strings.Split(points, " ")
	result := make([]string, len(pairs))
	for i, pair := range pairs {
		var x, y int
		fmt.Sscanf(pair, "%d,%d", &x, &y)
		result[i] = fmt.Sprintf("%d,%d", x+dx, y+dy)
	}
	return strings.Join(result, " ")
}

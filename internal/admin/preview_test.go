package admin

import (
	"strings"
	"testing"
	"time"

	"github.com/swilcox/led-kurokku-go/config"
	"github.com/swilcox/led-kurokku-go/segfont"
)

func TestEnabledWidgets(t *testing.T) {
	cfg := &config.Config{
		Widgets: []config.WidgetConfig{
			{Type: "clock", Enabled: true},
			{Type: "message", Enabled: false},
			{Type: "alert", Enabled: true},
		},
	}
	got := enabledWidgets(cfg)
	if len(got) != 2 {
		t.Fatalf("expected 2 enabled widgets, got %d", len(got))
	}
	if got[0] != 0 || got[1] != 2 {
		t.Errorf("indices = %v, want [0 2]", got)
	}
}

func TestEnabledWidgets_NoneEnabled(t *testing.T) {
	cfg := &config.Config{
		Widgets: []config.WidgetConfig{
			{Type: "clock", Enabled: false},
		},
	}
	got := enabledWidgets(cfg)
	if len(got) != 0 {
		t.Errorf("expected 0 enabled widgets, got %d", len(got))
	}
}

func TestPreviewDuration_Zero(t *testing.T) {
	w := config.WidgetConfig{Duration: config.Duration(0)}
	if got := previewDuration(w); got != 5 {
		t.Errorf("previewDuration(0s) = %d, want 5", got)
	}
}

func TestPreviewDuration_BelowMinimum(t *testing.T) {
	w := config.WidgetConfig{Duration: config.Duration(2 * time.Second)}
	if got := previewDuration(w); got != 3 {
		t.Errorf("previewDuration(2s) = %d, want 3", got)
	}
}

func TestPreviewDuration_Normal(t *testing.T) {
	w := config.WidgetConfig{Duration: config.Duration(10 * time.Second)}
	if got := previewDuration(w); got != 10 {
		t.Errorf("previewDuration(10s) = %d, want 10", got)
	}
}

func TestPreviewTextForWidget_Message(t *testing.T) {
	w := config.WidgetConfig{Type: "message", Text: "Hello"}
	got := previewTextForWidget(w, nil, "", 0)
	if got != "Hello" {
		t.Errorf("got %q, want %q", got, "Hello")
	}
}

func TestPreviewTextForWidget_MessageEmpty(t *testing.T) {
	w := config.WidgetConfig{Type: "message"}
	got := previewTextForWidget(w, nil, "", 0)
	if got != "[message]" {
		t.Errorf("got %q, want %q", got, "[message]")
	}
}

func TestPreviewTextForWidget_Alert(t *testing.T) {
	w := config.WidgetConfig{Type: "alert", Text: "Fire!"}
	got := previewTextForWidget(w, nil, "", 0)
	if got != "Fire!" {
		t.Errorf("got %q, want %q", got, "Fire!")
	}
}

func TestPreviewTextForWidget_AlertEmpty(t *testing.T) {
	w := config.WidgetConfig{Type: "alert"}
	got := previewTextForWidget(w, nil, "", 0)
	if got != "[alert]" {
		t.Errorf("got %q, want %q", got, "[alert]")
	}
}

func TestPreviewTextForWidget_Animation(t *testing.T) {
	w := config.WidgetConfig{Type: "animation", AnimationType: "bounce"}
	got := previewTextForWidget(w, nil, "", 0)
	if got != "[animation: bounce]" {
		t.Errorf("got %q, want %q", got, "[animation: bounce]")
	}
}

func TestPreviewTextForWidget_AnimationDefault(t *testing.T) {
	w := config.WidgetConfig{Type: "animation"}
	got := previewTextForWidget(w, nil, "", 0)
	if got != "[animation: frames]" {
		t.Errorf("got %q, want %q", got, "[animation: frames]")
	}
}

func TestPreviewTextForWidget_UnknownType(t *testing.T) {
	w := config.WidgetConfig{Type: "custom"}
	got := previewTextForWidget(w, nil, "", 0)
	if got != "[custom]" {
		t.Errorf("got %q, want %q", got, "[custom]")
	}
}

func TestSegmentTextForWidget_NonClock(t *testing.T) {
	w := config.WidgetConfig{Type: "message", Text: "Hi"}
	digits, colon := segmentTextForWidget(w, "Hi")
	if colon {
		t.Error("expected colon=false for message")
	}
	if digits != "Hi  " {
		t.Errorf("digits = %q, want %q", digits, "Hi  ")
	}
}

func TestSegmentTextForWidget_LongText(t *testing.T) {
	w := config.WidgetConfig{Type: "message"}
	digits, _ := segmentTextForWidget(w, "ABCDEF")
	if len(digits) != 4 {
		t.Errorf("expected 4 chars, got %d: %q", len(digits), digits)
	}
	if digits != "ABCD" {
		t.Errorf("digits = %q, want %q", digits, "ABCD")
	}
}

func TestSegmentTextForWidget_StripsColons(t *testing.T) {
	w := config.WidgetConfig{Type: "message"}
	digits, _ := segmentTextForWidget(w, "A:B")
	if strings.Contains(digits, ":") {
		t.Errorf("expected colons stripped, got %q", digits)
	}
}

func TestPixelPreviewHTML_ContainsDivs(t *testing.T) {
	html := pixelPreviewHTML("Hi")
	s := string(html)
	if !strings.Contains(s, `class="preview-pixel"`) {
		t.Error("expected preview-pixel class")
	}
	if !strings.Contains(s, `class="px on"`) {
		t.Error("expected at least one lit pixel")
	}
}

func TestSegmentPreviewHTML_7seg(t *testing.T) {
	html := segmentPreviewHTML("1234", segfont.Enc7, true, true)
	s := string(html)
	if !strings.Contains(s, "<svg") {
		t.Error("expected SVG element")
	}
	if !strings.Contains(s, "<circle") {
		t.Error("expected colon circles when colon=true")
	}
}

func TestSegmentPreviewHTML_14seg_NoColon(t *testing.T) {
	html := segmentPreviewHTML("ABCD", segfont.Enc14, false, false)
	s := string(html)
	if !strings.Contains(s, "<svg") {
		t.Error("expected SVG element")
	}
	if strings.Contains(s, "<circle") {
		t.Error("expected no colon circles when colon=false")
	}
}

func TestOffsetPolygon(t *testing.T) {
	got := offsetPolygon("0,0 10,5 20,10", 5, 3)
	if got != "5,3 15,8 25,13" {
		t.Errorf("offsetPolygon = %q, want %q", got, "5,3 15,8 25,13")
	}
}

func TestOffsetPolygon_Zero(t *testing.T) {
	input := "7,1 23,1"
	got := offsetPolygon(input, 0, 0)
	if got != input {
		t.Errorf("offsetPolygon(0,0) = %q, want %q", got, input)
	}
}

func TestPreviewWidgetHTML_OutOfRange(t *testing.T) {
	cfg := &config.Config{Widgets: []config.WidgetConfig{{Type: "clock", Enabled: true}}}
	html, label := previewWidgetHTML(cfg, "", 0, 5)
	if html != "" || label != "" {
		t.Error("expected empty result for out-of-range index")
	}
}

func TestPreviewWidgetHTML_NegativeIndex(t *testing.T) {
	cfg := &config.Config{Widgets: []config.WidgetConfig{{Type: "clock", Enabled: true}}}
	html, label := previewWidgetHTML(cfg, "", 0, -1)
	if html != "" || label != "" {
		t.Error("expected empty result for negative index")
	}
}

func TestPreviewTextForWidget_MessageDynamicSourceFallback(t *testing.T) {
	w := config.WidgetConfig{
		Type:          "message",
		DynamicSource: "some_key",
		Text:          "fallback text",
	}
	// No Redis available, host is empty — FetchKey will fail, should use Text fallback
	got := previewTextForWidget(w, nil, "", 0)
	if got != "fallback text" {
		t.Errorf("got %q, want %q", got, "fallback text")
	}
}

func TestPreviewTextForWidget_MessageDynamicSourceNoFallback(t *testing.T) {
	w := config.WidgetConfig{
		Type:          "message",
		DynamicSource: "some_key",
	}
	got := previewTextForWidget(w, nil, "", 0)
	if got != "[dynamic: some_key]" {
		t.Errorf("got %q, want %q", got, "[dynamic: some_key]")
	}
}

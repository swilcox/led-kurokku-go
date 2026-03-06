package config_test

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/swilcox/led-kurokku-go/config"
)

func TestDuration_MarshalUnmarshal(t *testing.T) {
	d := config.Duration(5 * time.Second)
	b, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	var d2 config.Duration
	if err := json.Unmarshal(b, &d2); err != nil {
		t.Fatal(err)
	}
	if d2 != d {
		t.Errorf("round-trip: got %v, want %v", d2, d)
	}
}

func TestDuration_Unwrap(t *testing.T) {
	d := config.Duration(500 * time.Millisecond)
	if d.Unwrap() != 500*time.Millisecond {
		t.Errorf("Unwrap: got %v, want %v", d.Unwrap(), 500*time.Millisecond)
	}
}

func TestDuration_UnmarshalInvalid(t *testing.T) {
	var d config.Duration
	if err := json.Unmarshal([]byte(`"notaduration"`), &d); err == nil {
		t.Error("expected error for invalid duration string")
	}
}

func TestDuration_UnmarshalNotString(t *testing.T) {
	var d config.Duration
	if err := json.Unmarshal([]byte(`123`), &d); err == nil {
		t.Error("expected error when duration is not a JSON string")
	}
}

func TestParse_Valid(t *testing.T) {
	data := []byte(`{
		"brightness": {"high": 15, "low": 1, "day_start": "07:00", "day_end": "22:00"},
		"widgets": [
			{"type": "clock", "enabled": true, "duration": "30s"}
		]
	}`)
	cfg, err := config.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Brightness.High != 15 {
		t.Errorf("brightness.high: got %d, want 15", cfg.Brightness.High)
	}
	if cfg.Brightness.DayStart != "07:00" {
		t.Errorf("day_start: got %q, want \"07:00\"", cfg.Brightness.DayStart)
	}
	if len(cfg.Widgets) != 1 {
		t.Fatalf("expected 1 widget, got %d", len(cfg.Widgets))
	}
	if cfg.Widgets[0].Type != "clock" {
		t.Errorf("widget type: got %q, want \"clock\"", cfg.Widgets[0].Type)
	}
}

func TestParse_WithLocation(t *testing.T) {
	data := []byte(`{
		"location": {"lat": 36.166, "lon": -86.784, "timezone": "America/Chicago"},
		"brightness": {"high": 15, "low": 1, "use_location": true},
		"widgets": []
	}`)
	cfg, err := config.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Location == nil {
		t.Fatal("expected location to be set")
	}
	if cfg.Location.Lat != 36.166 {
		t.Errorf("lat: got %v, want 36.166", cfg.Location.Lat)
	}
	if cfg.Location.Lon != -86.784 {
		t.Errorf("lon: got %v, want -86.784", cfg.Location.Lon)
	}
	if cfg.Location.Timezone != "America/Chicago" {
		t.Errorf("timezone: got %q, want \"America/Chicago\"", cfg.Location.Timezone)
	}
	if !cfg.Brightness.UseLocation {
		t.Error("expected use_location to be true")
	}
}

func TestParse_WithoutLocation(t *testing.T) {
	data := []byte(`{"brightness": {"high": 15, "low": 1}, "widgets": []}`)
	cfg, err := config.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Location != nil {
		t.Error("expected location to be nil when not specified")
	}
}

func TestParse_InvalidJSON(t *testing.T) {
	_, err := config.Parse([]byte(`not json`))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestParse_InvalidDuration(t *testing.T) {
	data := []byte(`{"brightness": {}, "widgets": [{"type": "clock", "enabled": true, "duration": "bad"}]}`)
	_, err := config.Parse(data)
	if err == nil {
		t.Error("expected error for invalid duration in widget")
	}
}

func TestIsSegment(t *testing.T) {
	tests := []struct {
		dtype config.DisplayType
		want  bool
	}{
		{config.DisplayTerminal, false},
		{config.DisplayMAX7219, false},
		{config.DisplayTM1637, true},
		{config.DisplayHT16K33, true},
		{config.DisplayTerminalSeg7, true},
		{config.DisplayTerminalSeg14, true},
		{"unknown", false},
	}
	for _, tc := range tests {
		dc := config.DisplayConfig{Type: tc.dtype}
		if got := dc.IsSegment(); got != tc.want {
			t.Errorf("IsSegment(%q) = %v, want %v", tc.dtype, got, tc.want)
		}
	}
}

func TestLoad_ValidFile(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/config.json"
	data := []byte(`{"brightness": {"high": 10, "low": 2}, "widgets": []}`)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Brightness.High != 10 {
		t.Errorf("brightness.high = %d, want 10", cfg.Brightness.High)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := config.Load("/nonexistent/path/config.json")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/bad.json"
	os.WriteFile(path, []byte(`{bad`), 0644)
	_, err := config.Load(path)
	if err == nil {
		t.Error("expected error for invalid JSON file")
	}
}

func TestParse_DisplayConfig(t *testing.T) {
	data := []byte(`{
		"display": {"type": "tm1637", "clk_pin": "GPIO23", "dio_pin": "GPIO24"},
		"brightness": {"high": 7},
		"widgets": []
	}`)
	cfg, err := config.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Display.Type != config.DisplayTM1637 {
		t.Errorf("display.type = %q, want %q", cfg.Display.Type, config.DisplayTM1637)
	}
	if cfg.Display.ClkPin != "GPIO23" {
		t.Errorf("clk_pin = %q, want %q", cfg.Display.ClkPin, "GPIO23")
	}
	if cfg.Display.DioPin != "GPIO24" {
		t.Errorf("dio_pin = %q, want %q", cfg.Display.DioPin, "GPIO24")
	}
}

func TestParse_WidgetConfigFields(t *testing.T) {
	f24 := true
	data := []byte(`{
		"brightness": {},
		"widgets": [{
			"type": "clock",
			"enabled": true,
			"duration": "30s",
			"format_24h": true,
			"cron": "*/5 * * * *"
		}]
	}`)
	cfg, err := config.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	w := cfg.Widgets[0]
	if w.Type != "clock" {
		t.Errorf("type = %q, want %q", w.Type, "clock")
	}
	if w.Format24h == nil || *w.Format24h != f24 {
		t.Error("expected format_24h = true")
	}
	if w.Cron != "*/5 * * * *" {
		t.Errorf("cron = %q, want %q", w.Cron, "*/5 * * * *")
	}
}

func TestValidate_Valid(t *testing.T) {
	cfg := &config.Config{
		Display:    config.DisplayConfig{Type: config.DisplayTerminal},
		Brightness: config.BrightnessConfig{High: 12, Low: 2, DayStart: "07:00", DayEnd: "22:00"},
		Widgets: []config.WidgetConfig{
			{Type: "clock", Enabled: true},
			{Type: "message", Enabled: true, Cron: "*/5 * * * *"},
		},
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("expected valid config, got: %v", err)
	}
}

func TestValidate_BadDisplayType(t *testing.T) {
	cfg := &config.Config{Display: config.DisplayConfig{Type: "nonexistent"}}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for unknown display type")
	}
}

func TestValidate_BrightnessOutOfRange(t *testing.T) {
	cfg := &config.Config{
		Display:    config.DisplayConfig{Type: config.DisplayTerminal},
		Brightness: config.BrightnessConfig{High: 20},
	}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for brightness > 15")
	}
}

func TestValidate_BadDayStart(t *testing.T) {
	cfg := &config.Config{
		Display:    config.DisplayConfig{Type: config.DisplayTerminal},
		Brightness: config.BrightnessConfig{DayStart: "7am"},
	}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for bad day_start format")
	}
}

func TestValidate_BadTimezone(t *testing.T) {
	cfg := &config.Config{
		Display: config.DisplayConfig{Type: config.DisplayTerminal},
		Location: &config.LocationConfig{
			Lat: 36.0, Lon: -86.0, Timezone: "Not/A/Zone",
		},
	}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for invalid timezone")
	}
}

func TestValidate_BadWidgetType(t *testing.T) {
	cfg := &config.Config{
		Display: config.DisplayConfig{Type: config.DisplayTerminal},
		Widgets: []config.WidgetConfig{{Type: "bogus", Enabled: true}},
	}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for unknown widget type")
	}
}

func TestValidate_BadCron(t *testing.T) {
	cfg := &config.Config{
		Display: config.DisplayConfig{Type: config.DisplayTerminal},
		Widgets: []config.WidgetConfig{{Type: "clock", Cron: "not a cron"}},
	}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for invalid cron expression")
	}
}

func TestValidate_EmptyDisplayType(t *testing.T) {
	cfg := &config.Config{
		Brightness: config.BrightnessConfig{High: 10, Low: 1},
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("empty display type should be valid (defaults at runtime), got: %v", err)
	}
}

func TestParse_AlertConfig(t *testing.T) {
	data := []byte(`{
		"brightness": {},
		"widgets": [{
			"type": "alert",
			"enabled": true,
			"duration": "10s",
			"alerts": [
				{"id": "a1", "message": "Test alert", "priority": 5, "display_duration": "3s", "delete_after_display": true}
			]
		}]
	}`)
	cfg, err := config.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	alerts := cfg.Widgets[0].Alerts
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(alerts))
	}
	if alerts[0].ID != "a1" {
		t.Errorf("alert id = %q, want %q", alerts[0].ID, "a1")
	}
	if alerts[0].Priority != 5 {
		t.Errorf("priority = %d, want 5", alerts[0].Priority)
	}
	if !alerts[0].DeleteAfterDisplay {
		t.Error("expected delete_after_display = true")
	}
}

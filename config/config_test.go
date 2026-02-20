package config_test

import (
	"encoding/json"
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

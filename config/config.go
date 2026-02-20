package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Duration wraps time.Duration for JSON string marshaling ("500ms", "5s").
type Duration time.Duration

func (d Duration) Unwrap() time.Duration {
	return time.Duration(d)
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	dur, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", s, err)
	}
	*d = Duration(dur)
	return nil
}

// Config is the top-level configuration.
type Config struct {
	Brightness BrightnessConfig `json:"brightness"`
	Widgets    []WidgetConfig   `json:"widgets"`
}

// BrightnessConfig controls time-of-day brightness.
type BrightnessConfig struct {
	High     byte   `json:"high"`
	Low      byte   `json:"low"`
	DayStart string `json:"day_start"`
	DayEnd   string `json:"day_end"`
}

// AlertConfig describes a single alert entry.
type AlertConfig struct {
	ID                string   `json:"id"`
	Message           string   `json:"message"`
	Priority          int      `json:"priority"`
	DisplayDuration   Duration `json:"display_duration"`
	DeleteAfterDisplay bool    `json:"delete_after_display"`
}

// FrameConfig describes a single animation frame.
type FrameConfig struct {
	Data     [32]byte `json:"data"`
	Duration Duration `json:"duration,omitempty"`
}

// WidgetConfig describes a single widget entry.
type WidgetConfig struct {
	Type     string   `json:"type"`
	Enabled  bool     `json:"enabled"`
	Duration Duration `json:"duration"`
	Cron     string   `json:"cron,omitempty"` // optional cron expression, e.g. "*/15 * * * *"
	// Clock
	Format24h *bool `json:"format_24h,omitempty"`
	// Message / Alert
	Text          string `json:"text,omitempty"`
	DynamicSource string `json:"dynamic_source,omitempty"`
	ScrollSpeed  Duration `json:"scroll_speed,omitempty"`
	Repeats      *int     `json:"repeats,omitempty"`
	SleepBetween Duration `json:"sleep_between,omitempty"`
	// Alert-specific
	Alerts []AlertConfig `json:"alerts,omitempty"`
	// Animation
	AnimationType string        `json:"animation_type,omitempty"`
	Frames        []FrameConfig `json:"frames,omitempty"`
	FrameDuration Duration      `json:"frame_duration,omitempty"`
}

// Parse parses JSON config data.
func Parse(data []byte) (*Config, error) {
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	return &cfg, nil
}

// Load reads and parses a JSON config file.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}
	return Parse(data)
}

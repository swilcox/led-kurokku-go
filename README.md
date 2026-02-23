# LED Kurokku Go

A Go application to drive LED displays on Raspberry Pi, with a widget-based display system and optional Redis integration for dynamic content.

## Supported Hardware

| Display | Type | Interface | Resolution |
|---------|------|-----------|------------|
| MAX7219 4-in-1 | Pixel matrix | SPI | 32x8 pixels |
| TM1637 | 7-segment | GPIO (bit-bang) | 4 digits |
| HT16K33 | 14-segment | I2C | 4 digits |

All three hardware types are driven from a single binary and config file. Terminal emulators are provided for development without hardware.

## Building

```bash
# Build for local machine (macOS dev)
make build

# Cross-compile for Raspberry Pi (linux/arm64)
make build-pi
```

## Usage

```bash
# Pixel matrix — terminal emulator
./kurokku -display terminal

# Pixel matrix — MAX7219 hardware
./kurokku -display max7219

# 7-segment — terminal emulator
./kurokku -display terminal_seg7

# 14-segment — terminal emulator
./kurokku -display terminal_seg14

# TM1637 hardware (requires GPIO pins in config)
./kurokku -display tm1637

# HT16K33 hardware (requires I2C config)
./kurokku -display ht16k33

# Custom config file
./kurokku -config my-config.json
```

### Flags

| Flag       | Default      | Description                         |
|------------|--------------|-------------------------------------|
| `-display` | *(from config, or `terminal`)* | Display type override |
| `-config`  | `config.json`| Path to JSON config file            |

The `-display` flag overrides the `display.type` field in the config file. If neither is set, it defaults to `terminal`.

## Configuration

All display behavior is driven by a JSON config file. Widgets cycle in order; each runs for its configured `duration` (use `"0s"` or omit for no timeout — the widget runs until done).

### Display Block

The `display` block selects the hardware backend and its settings:

```json
{
  "display": {
    "type": "terminal"
  }
}
```

**Pixel displays:**

```json
{ "display": { "type": "max7219" } }
```

**Segment displays:**

```json
{
  "display": {
    "type": "tm1637",
    "clk_pin": "GPIO23",
    "dio_pin": "GPIO24"
  }
}
```

```json
{
  "display": {
    "type": "ht16k33",
    "i2c_bus": "",
    "i2c_addr": 112,
    "layout": "adafruit"
  }
}
```

| Field      | Used By  | Description |
|------------|----------|-------------|
| `type`     | All      | `terminal`, `max7219`, `tm1637`, `ht16k33`, `terminal_seg7`, `terminal_seg14` |
| `clk_pin`  | TM1637   | GPIO clock pin (default `"GPIO23"`) |
| `dio_pin`  | TM1637   | GPIO data pin (default `"GPIO24"`) |
| `i2c_addr` | HT16K33  | I2C address (default `0x70` / `112`) |
| `i2c_bus`  | HT16K33  | I2C bus name (empty for default) |
| `layout`   | HT16K33  | `"sequential"` (default) or `"adafruit"` |

### Full Example (Pixel)

```json
{
  "display": { "type": "terminal" },
  "location": {
    "lat": 36.166,
    "lon": -86.784,
    "timezone": "America/Chicago"
  },
  "brightness": {
    "high": 12,
    "low": 2,
    "use_location": true
  },
  "widgets": [
    {
      "type": "clock",
      "enabled": true,
      "duration": "30s",
      "format_24h": false
    },
    {
      "type": "message",
      "enabled": true,
      "duration": "10s",
      "text": "Hello World!",
      "scroll_speed": "50ms",
      "repeats": 2,
      "sleep_between": "1s"
    },
    {
      "type": "alert",
      "enabled": true,
      "duration": "30s",
      "scroll_speed": "50ms",
      "alerts": [
        {
          "id": "weather",
          "message": "Heat advisory: 102F",
          "priority": 1,
          "display_duration": "10s",
          "delete_after_display": false
        }
      ]
    },
    {
      "type": "animation",
      "enabled": true,
      "duration": "10s",
      "animation_type": "rain"
    }
  ]
}
```

### Full Example (7-Segment)

```json
{
  "display": {
    "type": "tm1637",
    "clk_pin": "GPIO23",
    "dio_pin": "GPIO24"
  },
  "brightness": {
    "high": 12,
    "low": 2,
    "day_start": "07:00",
    "day_end": "22:00"
  },
  "widgets": [
    {
      "type": "clock",
      "enabled": true,
      "duration": "30s",
      "format_24h": true
    },
    {
      "type": "message",
      "enabled": true,
      "duration": "10s",
      "text": "HELLO",
      "scroll_speed": "300ms",
      "repeats": 2
    }
  ]
}
```

### Widget Types

| Type        | Pixel | Segment | Description |
|-------------|:-----:|:-------:|-------------|
| `clock`     | Yes | Yes | Time display with blinking colon. Set `format_24h` for 24-hour format. 12h PM uses double-blink pattern. |
| `message`   | Yes | Yes | Static or scrolling text. Supports `dynamic_source` for Redis-backed text. Pixel: 50ms scroll speed. Segment: 300ms per character. |
| `alert`     | Yes | Yes | Displays prioritized alerts. With Redis, fetches from `kurokku:alert:*` keys; without, uses the `alerts` array. |
| `animation` | Yes | Yes | Pixel: procedural (`rain`, `static`, `bounce`, `sine`, `scanner`, `life`) or custom `frames`. Segment: procedural (`rain`, `static`, `scanner`, `race`) or custom `segment_frames`. |

### Cron Scheduling

Any widget can include a `cron` field with a standard cron expression. The widget is skipped when the expression doesn't match the current time:

```json
{
  "type": "message",
  "enabled": true,
  "duration": "10s",
  "text": "Lunch time!",
  "cron": "0 12 * * *"
}
```

### Brightness

Brightness values are always specified in the **0-15 range**, regardless of display type. Displays with fewer hardware levels (e.g. TM1637 with 8 levels) map automatically.

Two modes:

**Location-based (recommended)** — set a top-level `location` block and enable `use_location`. Sunrise and sunset are computed automatically.

```json
{
  "location": { "lat": 36.166, "lon": -86.784, "timezone": "America/Chicago" },
  "brightness": { "high": 12, "low": 2, "use_location": true }
}
```

**Fixed times** — specify `day_start` and `day_end` in `HH:MM` (24-hour) format.

```json
{
  "brightness": { "high": 12, "low": 2, "day_start": "07:00", "day_end": "22:00" }
}
```

If `use_location` is true but no `location` block is present, or the timezone is invalid, the display defaults to `high` brightness and logs a warning.

### Message `dynamic_source`

When `dynamic_source` is set and Redis is connected, the message widget fetches its text from that Redis key each cycle. If the key doesn't exist or Redis is unavailable, it falls back to the `text` field.

```bash
redis-cli SET kurokku:weather:temp:spring_hill "72F"
```

## Redis Integration (Optional)

Redis provides dynamic alerts and message text at runtime. Everything works without Redis — the app degrades gracefully to JSON config values.

### Environment Variables

| Variable     | Description |
|--------------|-------------|
| `REDIS_URL`  | Full Redis URL (e.g. `redis://localhost:6379`). Checked first. |
| `REDIS_HOST` | Redis host. Used if `REDIS_URL` is not set. |
| `REDIS_PORT` | Redis port. Defaults to `6379`. |

If none are set, Redis is disabled entirely.

### Alert Keys

Alerts are stored as individual Redis keys with prefix `kurokku:alert:`:

```bash
# Add an alert
redis-cli SET kurokku:alert:weather '{"message":"Heat advisory","priority":1,"display_duration":"10s","delete_after_display":false}'

# Add a self-deleting alert
redis-cli SET kurokku:alert:reminder '{"message":"Take out trash","priority":5,"display_duration":"5s","delete_after_display":true}'

# Add an alert with TTL (auto-expires)
redis-cli SET kurokku:alert:temp-notice '{"message":"Brief notice","priority":3,"display_duration":"5s"}' EX 300

# Remove an alert
redis-cli DEL kurokku:alert:weather
```

The app uses Redis keyspace notifications to detect changes. When a new alert key is set, it **immediately interrupts** the current widget and displays all pending alerts sorted by priority.

### Alert JSON Fields

| Field                | Type   | Description |
|----------------------|--------|-------------|
| `message`            | string | Alert text to display |
| `priority`           | int    | Lower number = higher urgency |
| `display_duration`   | string | How long to show (e.g. `"5s"`) |
| `delete_after_display` | bool | Remove from Redis after showing |

## Hardware Wiring

### MAX7219 (SPI)

| MAX7219 Pin | RPi Pin | RPi GPIO | Description |
|-------------|---------|----------|-------------|
| VCC         | Pin 2   | 5V       | Power       |
| GND         | Pin 6   | GND      | Ground      |
| DIN         | Pin 19  | GPIO 10  | SPI0 MOSI   |
| CS          | Pin 24  | GPIO 8   | SPI0 CE0    |
| CLK         | Pin 23  | GPIO 11  | SPI0 SCLK   |

Enable SPI: `sudo raspi-config nonint do_spi 0`

### TM1637 (GPIO)

| TM1637 Pin | RPi Pin | Description |
|------------|---------|-------------|
| VCC        | Pin 1   | 3.3V Power  |
| GND        | Pin 9   | Ground      |
| CLK        | *(configurable)* | Clock (e.g. GPIO23) |
| DIO        | *(configurable)* | Data (e.g. GPIO24) |

### HT16K33 (I2C)

| HT16K33 Pin | RPi Pin | Description |
|-------------|---------|-------------|
| VIN         | Pin 1   | 3.3V Power  |
| GND         | Pin 9   | Ground      |
| SDA         | Pin 3   | I2C1 SDA    |
| SCL         | Pin 5   | I2C1 SCL    |

Enable I2C: `sudo raspi-config nonint do_i2c 0`

Default I2C address: `0x70`. Change with solder jumpers on the board.

## Project Structure

```
cmd/kurokku/main.go          Entry point, flag parsing, display creation
config/
  config.go                   Configuration types, DisplayConfig, JSON loading
display/
  display.go                  Display/PixelDisplay/SegmentDisplay interfaces
  terminal.go                 Terminal pixel emulator
  terminal_segment.go         Terminal segment emulator (7-seg & 14-seg ASCII art)
  max7219.go                  MAX7219 SPI driver
  tm1637.go                   TM1637 GPIO bit-bang driver
  ht16k33.go                  HT16K33 I2C driver
  testutil/spy.go             SpyDisplay + SpySegmentDisplay for tests
engine/
  engine.go                   Widget cycling loop, segment branching
font/
  font5x7.go                  5x7 bitmap font (pixel displays)
framebuf/
  framebuf.go                 32x8 framebuffer (pixel displays)
segfont/
  segfont.go                  7-seg and 14-seg character maps
redis/
  redis.go                    Optional Redis client
spi/
  spi.go                      SPI abstraction (periph.io)
widget/
  widget.go                   Widget interface, ScrollText, SleepOrCancel
  clock.go                    Pixel clock widget
  message.go                  Pixel message widget
  alert.go                    Pixel alert widget
  redis_alert.go              Redis-backed pixel alert
  redis_message.go            Redis-backed pixel message
  animation/                  Pixel animations (rain, static, bounce, sine, scanner, life)
  segment/
    clock.go                  Segment clock widget
    message.go                Segment message widget
    alert.go                  Segment alert widget
    animation.go              Segment frame animation
    redis_alert.go            Redis-backed segment alert
    redis_message.go          Redis-backed segment message
```

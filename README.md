# LED Kurokku Go

A Go application to drive a MAX7219 4-in-1 32x8 LED matrix on Raspberry Pi, with a widget-based display system and optional Redis integration for dynamic content.

## Raspberry Pi Wiring

Connect the MAX7219 module to the Raspberry Pi SPI0 pins:

| MAX7219 Pin | RPi Pin | RPi GPIO | Description |
|-------------|---------|----------|-------------|
| VCC         | Pin 2   | 5V       | Power       |
| GND         | Pin 6   | GND      | Ground      |
| DIN         | Pin 19  | GPIO 10  | SPI0 MOSI   |
| CS          | Pin 24  | GPIO 8   | SPI0 CE0    |
| CLK         | Pin 23  | GPIO 11  | SPI0 SCLK   |

> **Note:** The MAX7219 module requires 5V power. The logic pins are driven at 3.3V from the Pi's SPI GPIOs, which works reliably with most modules.

## Building

```bash
# Build for local machine (macOS dev)
make build

# Cross-compile for Raspberry Pi (linux/arm64)
make build-pi
```

## Usage

```bash
# Terminal display (for development/testing)
./kurokku -display terminal

# MAX7219 hardware display (on Raspberry Pi)
./kurokku -display max7219

# Custom config file
./kurokku -config my-config.json
```

### Flags

| Flag       | Default      | Description                         |
|------------|--------------|-------------------------------------|
| `-display` | `terminal`   | Display type: `terminal`, `max7219` |
| `-config`  | `config.json`| Path to JSON config file            |

## Configuration

All display behavior is driven by `config.json`. Widgets cycle in order; each runs for its configured `duration` (use `"0s"` or omit for no timeout — the widget runs until done).

```json
{
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
      "type": "message",
      "enabled": true,
      "duration": "10s",
      "text": "--\u00b0F",
      "dynamic_source": "kurokku:weather:temp:spring_hill",
      "scroll_speed": "50ms",
      "repeats": 1
    },
    {
      "type": "alert",
      "enabled": true,
      "duration": "0s",
      "scroll_speed": "50ms",
      "alerts": [
        {
          "id": "fallback",
          "message": "No alerts",
          "priority": 99,
          "display_duration": "3s",
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

### Widget Types

| Type        | Description |
|-------------|-------------|
| `clock`     | Displays time with blinking colon. Set `format_24h` for 24-hour format. |
| `message`   | Static or scrolling text. Supports `dynamic_source` for Redis-backed text. |
| `alert`     | Displays prioritized alerts. With Redis, fetches from `kurokku:alert:*` keys; without, uses the `alerts` array as fallback. |
| `animation` | Frame-based or procedural animations (`rain`, `random`, or custom `frames`). |

### Message `dynamic_source`

When `dynamic_source` is set and Redis is connected, the message widget fetches its text from that Redis key each cycle. If the key doesn't exist or Redis is unavailable, it falls back to the `text` field.

```json
{
  "type": "message",
  "enabled": true,
  "duration": "10s",
  "text": "--\u00b0F",
  "dynamic_source": "kurokku:weather:temp:spring_hill",
  "scroll_speed": "50ms"
}
```

```bash
redis-cli SET kurokku:weather:temp:spring_hill "72\u00b0F"
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

The app uses Redis keyspace notifications (`__keyspace@0__:kurokku:alert:*`) to detect changes. When a new alert key is set, it **immediately interrupts** the current widget and displays all pending alerts sorted by priority.

> **Note:** Redis must have keyspace notifications enabled. The app attempts to run `CONFIG SET notify-keyspace-events KEA` on startup. If Redis is configured to disallow this (e.g. managed Redis), enable it in your Redis config: `notify-keyspace-events KEA`.

### Alert JSON Fields

| Field                | Type   | Description |
|----------------------|--------|-------------|
| `message`            | string | Alert text to display |
| `priority`           | int    | Lower number = higher urgency |
| `display_duration`   | string | How long to show (e.g. `"5s"`) |
| `delete_after_display` | bool | Remove from Redis after showing |

## Enabling SPI on the Raspberry Pi

SPI is disabled by default. Enable it with:

```bash
sudo raspi-config nonint do_spi 0
```

Or add `dtparam=spi=on` to `/boot/firmware/config.txt` and reboot.

Verify the SPI device exists:

```bash
ls /dev/spidev0.*
```

## Project Structure

```
├── cmd/kurokku/main.go    # Entry point, flag parsing, Redis init
├── config/
│   └── config.go          # Configuration types and JSON loading
├── display/
│   ├── display.go         # Display interface
│   ├── max7219.go         # MAX7219 4-in-1 matrix driver
│   └── terminal.go        # Virtual terminal display
├── engine/
│   └── engine.go          # Widget cycling loop with interrupt support
├── font/
│   └── font5x7.go         # 5x7 bitmap font
├── framebuf/
│   └── framebuf.go        # 32x8 framebuffer
├── redis/
│   └── redis.go           # Redis client (alerts, messages, keyspace sub)
├── spi/
│   └── spi.go             # SPI abstraction (periph.io)
└── widget/
    ├── widget.go           # Widget interface
    ├── clock.go            # Clock widget
    ├── message.go          # Message widget
    ├── alert.go            # Alert widget
    ├── redis_alert.go      # Redis-backed alert wrapper
    ├── redis_message.go    # Redis-backed message wrapper
    └── animation/
        ├── animation.go    # Frame-based animation
        ├── procedural.go   # Animation registry
        ├── rain.go         # Rain animation
        └── random.go       # Random noise animation
```

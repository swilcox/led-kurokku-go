# LED Kurokku Go

## Overview

Go application driving LED displays on Raspberry Pi. Supports three hardware types from one codebase:

- **MAX7219** — 32x8 pixel matrix (SPI)
- **TM1637** — 4-digit 7-segment (GPIO bit-bang)
- **HT16K33** — 4-digit 14-segment (I2C)

Widget-based display system with optional Redis integration for dynamic content.

## Build & Run

```bash
make build              # local binary
make build-pi           # cross-compile for RPi (linux/arm64)
go build ./...          # verify all packages compile
go test ./...           # run full test suite
```

Run: `./kurokku -display terminal -config config.json`

Display types: `terminal`, `max7219`, `tm1637`, `ht16k33`, `terminal_seg7`, `terminal_seg14`

## Architecture

### Display Interface Hierarchy

- **display.Display** — base interface: `Init`, `Close`, `Clear`, `SetBrightness`
- **display.PixelDisplay** — extends Display: `WriteFramebuffer`, `Width`, `Height`
- **display.SegmentDisplay** — extends Display: `WriteSegments([]uint16, colon bool)`, `DisplayLength`

Widgets accept `display.Display` in `Run()` and type-assert to the sub-interface they need. The engine only uses the base `Display` for brightness/clear.

### Packages

- **config/** — JSON config parsing, custom `Duration` type, `DisplayConfig` with `IsSegment()` helper
- **display/** — Display interfaces + backends: terminal, terminal_segment, MAX7219, TM1637, HT16K33
- **engine/** — Widget cycling loop. Branches on `cfg.Display.IsSegment()` to build pixel or segment widgets. Supports Redis alert interrupts via keyspace notifications.
- **widget/** — `Widget` interface (`Name()`, `Run(ctx, disp)`). Pixel widgets: clock, message, alert, animation. Redis wrappers delegate to inner widgets.
- **widget/segment/** — Segment display widgets: clock, message, alert, animation, redis_alert, redis_message. Mirror pixel widget behavior adapted for character-based displays.
- **widget/animation/** — Procedural pixel animations (rain, random, bounce, sine, scanner, life) and frame-based animation.
- **segfont/** — 7-segment (`Seg7`) and 14-segment (`Seg14`) character maps. `Encoder` type with `Enc7`/`Enc14` implementations.
- **redis/** — Optional Redis client. Alerts via `kurokku:alert:*`, messages via `dynamic_source` key. Keyspace notifications for live alert interrupts.
- **framebuf/** — 32x8 column-based framebuffer (pixel displays only)
- **font/** — 5x7 bitmap font (ASCII 32-126, pixel displays only)
- **spi/** — SPI abstraction via periph.io

## Key Patterns

- Widget `duration: "0s"` means no timeout (widget runs to completion)
- Redis is fully optional — `redis.NewFromEnv()` returns `(nil, nil)` when no env vars set
- `RedisAlert`/`RedisMessage` wrappers are only constructed when Redis is connected
- Alert interrupt: keyspace notification cancels current widget, fetches+displays all alerts, then resumes cycle
- Graceful degradation: Redis fetch errors fall back to JSON config values
- Engine uses `segmentEncoder()` to select `Enc7` (TM1637/terminal_seg7) or `Enc14` (HT16K33/terminal_seg14)
- CLI `-display` flag overrides `config.display.type`; default is `terminal`

## Test Patterns

- `SpyDisplay` (pixel) and `SpySegmentDisplay` (segment) in `display/testutil/spy.go`
- Engine tests: white-box (`package engine`) to access unexported `nowFunc`, `updateBrightness`
- Widget tests: black-box (`package widget_test`, `package segment_test`)
- Time injection: `NowFunc func() time.Time` on Clock/Alert structs
- Engine inner loop MUST check `ctx.Err()` before cron `continue` to avoid busy-loop

## Redis Key Layout

- `kurokku:alert:<id>` — individual alert JSON
- `dynamic_source` field value — arbitrary key for message text override

## Dependencies

- `github.com/redis/go-redis/v9` — Redis client
- `periph.io/x/conn/v3`, `periph.io/x/host/v3` — SPI/GPIO/I2C for hardware displays
- `github.com/nathan-osman/go-sunrise` — sunrise/sunset calculations
- `github.com/robfig/cron/v3` — cron expression matching

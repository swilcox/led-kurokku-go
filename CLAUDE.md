# LED Kurokku Go

## Overview

Go application driving a MAX7219 32x8 LED matrix on Raspberry Pi. Widget-based display system with optional Redis integration for dynamic content.

## Build & Run

```bash
make build              # local binary
make build-pi           # cross-compile for RPi (linux/arm64)
go build ./...          # verify all packages compile
```

Run: `./kurokku -display terminal -config config.json`

## Architecture

- **config/** — JSON config parsing, custom `Duration` type (wraps `time.Duration` for JSON strings like `"5s"`)
- **display/** — `Display` interface with terminal (dev) and MAX7219 (hardware) backends
- **engine/** — Widget cycling loop. Runs widgets sequentially with per-widget timeouts. Supports Redis alert interrupts via keyspace notifications.
- **widget/** — `Widget` interface (`Name()`, `Run(ctx, disp)`). Types: clock, message, alert, animation. Redis wrappers (`RedisAlert`, `RedisMessage`) delegate to inner widgets.
- **redis/** — Optional Redis client. Alerts stored as individual keys (`kurokku:alert:*`), scanned with `SCAN`. Messages fetched by explicit key from `dynamic_source` config field. Keyspace notifications (`__keyspace@0__:kurokku:alert:*`) for live alert interrupts.
- **framebuf/** — 32x8 column-based framebuffer
- **font/** — 5x7 bitmap font (ASCII 32-126)
- **spi/** — SPI abstraction via periph.io

## Key Patterns

- Widget `duration: "0s"` means no timeout (widget runs to completion)
- Redis is fully optional — `redis.NewFromEnv()` returns `(nil, nil)` when no env vars set
- `RedisAlert`/`RedisMessage` wrappers are only constructed when Redis is connected (and for messages, only when `dynamic_source` is set)
- Alert interrupt: keyspace notification cancels current widget, fetches+displays all alerts, then resumes cycle
- Graceful degradation: Redis fetch errors fall back to JSON config values

## Redis Key Layout

- `kurokku:alert:<id>` — individual alert JSON
- `dynamic_source` field value — arbitrary key for message text override (e.g. `kurokku:weather:temp:spring_hill`)

## Dependencies

- `github.com/redis/go-redis/v9` — Redis client
- `periph.io/x/conn/v3`, `periph.io/x/host/v3` — SPI/GPIO for MAX7219

# Development Guide

## Prerequisites

- Go 1.25+ (uses latest toolchain features)
- No hardware required — terminal emulators are provided for all display types

## Quick Start

```bash
# Clone and build
git clone <repo-url>
cd led-kurokku-go
make build

# Run with pixel terminal display
./kurokku -display terminal -config config.json

# Run with 7-segment terminal display
./kurokku -display terminal_seg7 -config config.json

# Run with 14-segment terminal display
./kurokku -display terminal_seg14 -config config.json
```

## Project Layout

```
cmd/kurokku/       Entry point — flag parsing, display creation, engine startup
config/            JSON configuration types and parsing
display/           Display interfaces and all backends
  testutil/        SpyDisplay + SpySegmentDisplay for tests
engine/            Widget cycling loop and brightness control
font/              5x7 bitmap font for pixel displays
framebuf/          32x8 framebuffer for pixel displays
internal/cronutil/ Cron expression matching
redis/             Optional Redis client
segfont/           7-segment and 14-segment character maps
spi/               SPI abstraction layer
widget/            Pixel widget implementations
  animation/       Procedural and frame-based pixel animations
  segment/         Segment widget implementations
docs/              This documentation
```

## Testing

```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./widget/segment/...
go test ./segfont/...
go test ./engine/...

# Run with verbose output
go test -v ./widget/...

# Run a specific test
go test -run TestSegmentClock_24h ./widget/segment/...
```

### Test Utilities

The `display/testutil` package provides spy implementations for testing:

```go
// Pixel display spy
spy := &testutil.SpyDisplay{}
// spy.Frames — recorded [][]byte framebuffer writes
// spy.Brightness — recorded []byte brightness calls
// spy.ClearCalls — count of Clear() calls

// Segment display spy
spy := &testutil.SpySegmentDisplay{}
// spy.Calls — recorded []SegmentCall (Segments []uint16, Colon bool)
// spy.Brightness — recorded []byte brightness calls
// spy.ClearCalls — count of Clear() calls
// spy.Length — configurable display length (default 4)
```

### Test Patterns

**Time injection:** Clock and Alert widgets have a `NowFunc` field:

```go
clk := &widget.Clock{
    Format24h: true,
    NowFunc:   func() time.Time { return time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC) },
}
```

**Engine white-box tests:** Engine tests use `package engine` to access the unexported `nowFunc` field:

```go
e := New(spy, cfg, nil)
e.nowFunc = func() time.Time { return fixedTime }
e.updateBrightness()
```

**Context-based lifecycle:** Widgets run until their context is cancelled. Tests create short-lived contexts:

```go
ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
defer cancel()
widget.Run(ctx, spy)
// Assert on spy.Frames, spy.Calls, etc.
```

## Adding a New Widget

### Pixel Widget

1. Create `widget/mywidget.go`:

```go
package widget

import (
    "context"
    "github.com/swilcox/led-kurokku-go/display"
)

type MyWidget struct {
    // config fields
}

func (w *MyWidget) Name() string { return "mywidget" }

func (w *MyWidget) Run(ctx context.Context, disp display.Display) error {
    pd := disp.(display.PixelDisplay)
    // render loop using pd.WriteFramebuffer()
    // use SleepOrCancel(ctx, duration) between frames
    // return nil or ctx.Err()
}
```

2. Add a test in `widget/mywidget_test.go`
3. Add a case in `engine.buildWidgets()` for the pixel path
4. Add the widget type string to the config docs

### Segment Widget

1. Create `widget/segment/mywidget.go` — same pattern but type-assert to `display.SegmentDisplay`
2. Add a test using `SpySegmentDisplay`
3. Add a case in `engine.buildWidgets()` for the segment path (under the `isSeg` branch)

### Procedural Pixel Animation

1. Create `widget/animation/myanimation.go`:

```go
package animation

import (
    "context"
    "github.com/swilcox/led-kurokku-go/display"
    "github.com/swilcox/led-kurokku-go/framebuf"
)

type MyAnimation struct{}

func (a *MyAnimation) Name() string { return "myanimation" }

func (a *MyAnimation) Run(ctx context.Context, disp display.Display) error {
    pd := disp.(display.PixelDisplay)
    ticker := time.NewTicker(50 * time.Millisecond)
    defer ticker.Stop()
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-ticker.C:
        }
        var f framebuf.Frame
        // populate frame...
        pd.WriteFramebuffer(f.Bytes())
    }
}
```

2. Register it in `widget/animation/procedural.go`:

```go
var Registry = map[string]func() widget.Widget{
    // existing entries...
    "myanimation": func() widget.Widget { return &MyAnimation{} },
}
```

3. Add a test in `widget/animation/animation_test.go`

## Adding a New Display Backend

### Pixel Display

1. Create `display/mydriver.go` implementing `PixelDisplay`:
   - `Init() error`, `Close() error`, `Clear()`, `SetBrightness(byte)`
   - `WriteFramebuffer([]byte)`, `Width() int`, `Height() int`

2. Add a constructor `NewMyDriver(...)` and a case in `cmd/kurokku/main.go`'s `createDisplay()`

3. Add the display type constant to `config/config.go`

### Segment Display

1. Create `display/mydriver.go` implementing `SegmentDisplay`:
   - `Init() error`, `Close() error`, `Clear()`, `SetBrightness(byte)`
   - `WriteSegments([]uint16, bool)`, `DisplayLength() int`

2. Add the type to `config.IsSegment()` so the engine branches correctly

3. Add a constructor and case in `createDisplay()`

## Cross-Compilation

```bash
# Raspberry Pi (64-bit)
GOOS=linux GOARCH=arm64 go build -o kurokku ./cmd/kurokku

# Or use the Makefile
make build-pi
```

Deploy the binary to the Pi:

```bash
scp kurokku pi@raspberrypi:~/
scp config.json pi@raspberrypi:~/
ssh pi@raspberrypi ./kurokku -config config.json
```

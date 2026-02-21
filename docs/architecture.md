# Architecture

## High-Level Overview

LED Kurokku Go is a single binary that drives three types of LED displays from one widget-based engine. The architecture separates display concerns (how to render) from widget logic (what to render) through a layered interface hierarchy.

```mermaid
graph TB
    subgraph "Entry Point"
        main["cmd/kurokku/main.go"]
    end

    subgraph "Core"
        engine["engine/"]
        config["config/"]
    end

    subgraph "Display Layer"
        di["display.Display (base)"]
        pdi["display.PixelDisplay"]
        sdi["display.SegmentDisplay"]

        di --> pdi
        di --> sdi

        pdi --> term["Terminal"]
        pdi --> max["MAX7219"]
        sdi --> tseg["TerminalSegment"]
        sdi --> tm["TM1637"]
        sdi --> ht["HT16K33"]
    end

    subgraph "Widget Layer"
        wi["widget.Widget"]
        wi --> pw["Pixel Widgets"]
        wi --> sw["Segment Widgets"]

        pw --> pc["Clock"]
        pw --> pm["Message"]
        pw --> pa["Alert"]
        pw --> pan["Animation"]

        sw --> sc["segment.Clock"]
        sw --> sm["segment.Message"]
        sw --> sa["segment.Alert"]
        sw --> san["segment.FrameAnimation"]
    end

    subgraph "Support"
        redis["redis/"]
        font["font/"]
        framebuf["framebuf/"]
        segfont["segfont/"]
        spi["spi/"]
        cronutil["internal/cronutil/"]
    end

    main --> engine
    main --> config
    main --> di
    engine --> wi
    engine --> config
    engine --> redis
    pw --> pdi
    pw --> font
    pw --> framebuf
    sw --> sdi
    sw --> segfont
```

## Display Interface Hierarchy

The display system uses a slim base interface with specialized sub-interfaces. This allows the engine and widget infrastructure to work with any display type while widgets access the specific rendering methods they need via type assertion.

```mermaid
classDiagram
    class Display {
        <<interface>>
        +Init() error
        +Close() error
        +Clear()
        +SetBrightness(level byte)
    }

    class PixelDisplay {
        <<interface>>
        +WriteFramebuffer(buf []byte)
        +Width() int
        +Height() int
    }

    class SegmentDisplay {
        <<interface>>
        +WriteSegments(segments []uint16, colon bool)
        +DisplayLength() int
    }

    Display <|-- PixelDisplay
    Display <|-- SegmentDisplay

    class Terminal {
        -w io.Writer
        -width int
        -height int
    }

    class MAX7219 {
        -dev *spi.Device
    }

    class TerminalSegment {
        -w io.Writer
        -segType SegmentType
        -length int
    }

    class TM1637 {
        -clk gpio.PinIO
        -dio gpio.PinIO
        -bright byte
    }

    class HT16K33 {
        -bus i2c.BusCloser
        -dev *i2c.Dev
        -prev [8]byte
    }

    PixelDisplay <|.. Terminal
    PixelDisplay <|.. MAX7219
    SegmentDisplay <|.. TerminalSegment
    SegmentDisplay <|.. TM1637
    SegmentDisplay <|.. HT16K33
```

### Why Type Assertions?

Widgets receive `display.Display` (the base interface) in their `Run` method signature. This keeps the `Widget` interface uniform — all widgets share one `Run(ctx, disp)` signature regardless of which display type they target. Inside `Run`, each widget type-asserts to the sub-interface it needs:

```go
// Pixel widget
func (c *Clock) Run(ctx context.Context, disp display.Display) error {
    pd := disp.(display.PixelDisplay)
    // use pd.WriteFramebuffer(), pd.Width(), etc.
}

// Segment widget
func (c *segment.Clock) Run(ctx context.Context, disp display.Display) error {
    sd := disp.(display.SegmentDisplay)
    // use sd.WriteSegments(), sd.DisplayLength(), etc.
}
```

The engine guarantees the correct pairing: it checks `cfg.Display.IsSegment()` and constructs only the matching widget variant, so the type assertion never panics at runtime.

## Engine Flow

The engine is the central coordinator. It builds widgets from config, cycles through them, and handles Redis alert interrupts.

```mermaid
flowchart TD
    start([Engine.Run]) --> build[buildWidgets]
    build --> brightness[Start brightnessLoop goroutine]
    brightness --> subscribe[Subscribe Redis alerts]
    subscribe --> loop

    subgraph loop [Widget Cycle Loop]
        check{ctx cancelled?}
        check -- Yes --> done([Return nil])
        check -- No --> cron{Cron matches?}
        cron -- Skip --> check
        cron -- Match --> timeout[Create widget context with timeout]
        timeout --> run[Run widget in goroutine]
        run --> select{Select}
        select -- widget done --> cancel[Cancel context]
        select -- alert interrupt --> cancelw[Cancel widget]
        cancelw --> wait[Wait for widget goroutine]
        wait --> alerts[runInterruptAlerts]
        alerts --> check
        select -- ctx done --> canceld[Cancel + wait]
        canceld --> done
        cancel --> check
    end
```

### Widget Building

The engine's `buildWidgets()` method branches on `cfg.Display.IsSegment()` for every widget type:

```mermaid
flowchart LR
    wc[WidgetConfig] --> type{Widget Type?}
    type -- clock --> seg{Segment?}
    seg -- Yes --> sc[segment.Clock]
    seg -- No --> pc[widget.Clock]

    type -- message --> seg2{Segment?}
    seg2 -- Yes --> redis2{Redis + DynamicSource?}
    redis2 -- Yes --> srm[segment.RedisMessage]
    redis2 -- No --> sm[segment.Message]
    seg2 -- No --> redis3{Redis + DynamicSource?}
    redis3 -- Yes --> prm[widget.RedisMessage]
    redis3 -- No --> pm[widget.Message]

    type -- alert --> seg3{Segment?}
    seg3 -- Yes --> redis4{Redis?}
    redis4 -- Yes --> sra[segment.RedisAlert]
    redis4 -- No --> sa[segment.Alert]
    seg3 -- No --> redis5{Redis?}
    redis5 -- Yes --> pra[widget.RedisAlert]
    redis5 -- No --> pa[widget.Alert]

    type -- animation --> seg4{Segment?}
    seg4 -- Yes --> san[segment.FrameAnimation]
    seg4 -- No --> anim{animation_type?}
    anim -- frames/empty --> pfr[animation.FrameAnimation]
    anim -- named --> reg[animation.Registry lookup]
```

## Data Flow

### Pixel Rendering Path

```mermaid
sequenceDiagram
    participant W as Widget (e.g. Clock)
    participant F as framebuf.Frame
    participant Font as font.RenderText
    participant D as PixelDisplay

    W->>Font: RenderText("14:30")
    Font-->>W: []byte (column data)
    W->>F: BlitText(&frame, text, offset)
    W->>D: WriteFramebuffer(frame.Bytes())
    D->>D: Render to hardware/terminal
```

### Segment Rendering Path

```mermaid
sequenceDiagram
    participant W as segment.Clock
    participant SF as segfont
    participant D as SegmentDisplay

    W->>SF: EncodeText(Enc7, "1430")
    SF-->>W: []uint16 (segment bitmasks)
    W->>D: WriteSegments(segments, colon=true)
    D->>D: Render to hardware/terminal
```

## Redis Integration

Redis is entirely optional. The system is designed for graceful degradation at every level.

```mermaid
flowchart TD
    env{Redis env vars set?}
    env -- No --> noredis[Redis disabled — use config.json only]
    env -- Yes --> ping{Redis reachable?}
    ping -- No --> fallback[Log warning, continue without Redis]
    ping -- Yes --> connected[Redis connected]

    connected --> config[Try fetch config from Redis]
    config -- Found --> userc[Use Redis config]
    config -- Not found --> usefile[Use config.json]

    connected --> subscribe[Subscribe keyspace notifications]
    subscribe --> interrupt[Alert key changes trigger interrupt]

    connected --> fetch[Widget fetches dynamic data]
    fetch -- Error --> fbtext[Fall back to config.json value]
    fetch -- Success --> dyntext[Use Redis value]
```

### Alert Interrupt Sequence

```mermaid
sequenceDiagram
    participant R as Redis
    participant E as Engine
    participant CW as Current Widget
    participant AW as Alert Widget

    R->>E: Keyspace notification (alert key changed)
    E->>CW: Cancel context
    CW-->>E: Goroutine exits
    E->>R: FetchAlerts (SCAN kurokku:alert:*)
    R-->>E: []AlertConfig
    E->>AW: Run(alertCtx, disp)
    AW->>AW: Sort by priority, display each
    AW-->>E: Done
    E->>E: Resume widget cycle
```

## Brightness Control

The engine runs a background goroutine that adjusts brightness every 60 seconds.

```mermaid
flowchart TD
    start[updateBrightness] --> loc{use_location?}
    loc -- Yes --> hasloc{Location configured?}
    hasloc -- No --> high[Set high brightness + warn]
    hasloc -- Yes --> tz{Valid timezone?}
    tz -- No --> high
    tz -- Yes --> calc[Calculate sunrise/sunset]
    calc --> sun{Now between rise and set?}
    sun -- Yes --> day[Set high brightness]
    sun -- No --> night[Set low brightness]

    loc -- No --> parse{Parse day_start/day_end}
    parse -- Error --> high
    parse -- OK --> time{Now in day range?}
    time -- Yes --> day
    time -- No --> night
```

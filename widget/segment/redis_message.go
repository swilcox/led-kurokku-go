package segment

import (
	"context"
	"log"
	"time"

	"github.com/swilcox/led-kurokku-go/display"
	"github.com/swilcox/led-kurokku-go/segfont"
	"github.com/swilcox/led-kurokku-go/widget"
)

// RedisMessage wraps segment message display with Redis-backed text overrides.
type RedisMessage struct {
	Fetcher      widget.MessageTextFetcher
	Key          string
	FallbackText string
	ScrollSpeed  time.Duration
	Repeats      int
	SleepBetween time.Duration
	Encoder      segfont.Encoder
}

func (rm *RedisMessage) Name() string { return "segment-redis-message" }

func (rm *RedisMessage) Run(ctx context.Context, disp display.Display) error {
	text := rm.FallbackText

	if override, ok, err := rm.Fetcher.FetchMessageText(ctx, rm.Key); err != nil {
		log.Printf("redis message fetch %s failed, using fallback: %v", rm.Key, err)
	} else if ok {
		text = override
	}

	m := &Message{
		Text:         text,
		ScrollSpeed:  rm.ScrollSpeed,
		Repeats:      rm.Repeats,
		SleepBetween: rm.SleepBetween,
		Encoder:      rm.Encoder,
	}
	return m.Run(ctx, disp)
}

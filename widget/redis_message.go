package widget

import (
	"context"
	"log"
	"time"

	"github.com/swilcox/led-kurokku-go/display"
)

// MessageTextFetcher fetches text from a Redis key.
type MessageTextFetcher interface {
	FetchMessageText(ctx context.Context, key string) (string, bool, error)
}

// RedisMessage wraps message display with Redis-backed text overrides.
// On each Run it fetches text from the configured Redis key, falling back
// to the configured JSON text on error or key absence.
type RedisMessage struct {
	Fetcher      MessageTextFetcher
	Key          string
	FallbackText string
	ScrollSpeed  time.Duration
	Repeats      int
	SleepBetween time.Duration
}

func (rm *RedisMessage) Name() string { return "redis-message" }

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
	}
	return m.Run(ctx, disp)
}

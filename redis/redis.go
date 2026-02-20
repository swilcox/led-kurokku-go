package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
	"github.com/swilcox/led-kurokku-go/config"
)

const (
	alertKeyPrefix       = "kurokku:alert:"
	alertKeyspacePattern = "__keyspace@0__:" + alertKeyPrefix + "*"

	configKey             = "kurokku:config"
	configKeyspacePattern = "__keyspace@0__:" + configKey
)

// Client wraps a Redis connection for kurokku operations.
type Client struct {
	rdb *redis.Client
}

// NewFromEnv creates a Redis client from environment variables.
// Checks REDIS_URL first, then REDIS_HOST + REDIS_PORT (default 6379).
// Returns (nil, nil) if no environment variables are set.
func NewFromEnv() (*Client, error) {
	if url := os.Getenv("REDIS_URL"); url != "" {
		opts, err := redis.ParseURL(url)
		if err != nil {
			return nil, fmt.Errorf("parsing REDIS_URL: %w", err)
		}
		return &Client{rdb: redis.NewClient(opts)}, nil
	}

	host := os.Getenv("REDIS_HOST")
	if host == "" {
		return nil, nil
	}

	port := os.Getenv("REDIS_PORT")
	if port == "" {
		port = "6379"
	}

	return &Client{rdb: redis.NewClient(&redis.Options{
		Addr: host + ":" + port,
	})}, nil
}

// Ping checks connectivity.
func (c *Client) Ping(ctx context.Context) error {
	return c.rdb.Ping(ctx).Err()
}

// Close closes the Redis connection.
func (c *Client) Close() error {
	return c.rdb.Close()
}

// FetchAlerts scans for all keys matching kurokku:alert:* and returns their values.
// Each key stores a JSON AlertConfig. The alert ID is derived from the key suffix.
func (c *Client) FetchAlerts(ctx context.Context) ([]config.AlertConfig, error) {
	var alerts []config.AlertConfig
	iter := c.rdb.Scan(ctx, 0, alertKeyPrefix+"*", 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		raw, err := c.rdb.Get(ctx, key).Result()
		if err == redis.Nil {
			continue // expired between scan and get
		}
		if err != nil {
			return nil, fmt.Errorf("GET %s: %w", key, err)
		}
		var ac config.AlertConfig
		if err := json.Unmarshal([]byte(raw), &ac); err != nil {
			return nil, fmt.Errorf("unmarshal alert %s: %w", key, err)
		}
		// Derive ID from key suffix if not set in JSON.
		if ac.ID == "" {
			ac.ID = key[len(alertKeyPrefix):]
		}
		alerts = append(alerts, ac)
	}
	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("SCAN %s*: %w", alertKeyPrefix, err)
	}
	return alerts, nil
}

// DeleteAlert removes an alert by deleting its key (kurokku:alert:<id>).
func (c *Client) DeleteAlert(ctx context.Context, id string) error {
	return c.rdb.Del(ctx, alertKeyPrefix+id).Err()
}

// FetchMessageText returns the value of the given Redis key.
// Returns ("", false, nil) if the key does not exist.
func (c *Client) FetchMessageText(ctx context.Context, key string) (string, bool, error) {
	val, err := c.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("GET %s: %w", key, err)
	}
	return val, true, nil
}

// FetchConfig fetches the full config JSON stored at kurokku:config.
// Returns (nil, false, nil) when the key is absent.
func (c *Client) FetchConfig(ctx context.Context) (*config.Config, bool, error) {
	raw, err := c.rdb.Get(ctx, configKey).Result()
	if err == redis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("GET %s: %w", configKey, err)
	}
	cfg, err := config.Parse([]byte(raw))
	if err != nil {
		return nil, false, err
	}
	return cfg, true, nil
}

// SubscribeConfig enables Redis keyspace notifications and subscribes to
// changes on kurokku:config via __keyspace@0__:kurokku:config.
// Returns a buffered(1) signal channel that receives a value whenever the
// config key is set or deleted.
func (c *Client) SubscribeConfig(ctx context.Context) (<-chan struct{}, error) {
	if err := c.rdb.ConfigSet(ctx, "notify-keyspace-events", "KEA").Err(); err != nil {
		log.Printf("warning: could not set notify-keyspace-events: %v", err)
	}

	sub := c.rdb.PSubscribe(ctx, configKeyspacePattern)
	if _, err := sub.Receive(ctx); err != nil {
		return nil, fmt.Errorf("psubscribe %s: %w", configKeyspacePattern, err)
	}

	ch := make(chan struct{}, 1)
	go func() {
		defer sub.Close()
		msgCh := sub.Channel()
		for {
			select {
			case <-ctx.Done():
				return
			case _, ok := <-msgCh:
				if !ok {
					return
				}
				select {
				case ch <- struct{}{}:
				default:
				}
			}
		}
	}()
	return ch, nil
}

// SubscribeAlerts enables Redis keyspace notifications and subscribes to
// key changes on kurokku:alert:* via pattern __keyspace@0__:kurokku:alert:*.
// Returns a buffered(1) signal channel that receives a value whenever an
// alert key is set, deleted, or expires.
func (c *Client) SubscribeAlerts(ctx context.Context) (<-chan struct{}, error) {
	// Enable keyspace notifications (KEA = Keyspace + Keyevent + All standard events).
	if err := c.rdb.ConfigSet(ctx, "notify-keyspace-events", "KEA").Err(); err != nil {
		log.Printf("warning: could not set notify-keyspace-events: %v", err)
	}

	sub := c.rdb.PSubscribe(ctx, alertKeyspacePattern)
	// Wait for subscription confirmation.
	if _, err := sub.Receive(ctx); err != nil {
		return nil, fmt.Errorf("psubscribe %s: %w", alertKeyspacePattern, err)
	}

	ch := make(chan struct{}, 1)
	go func() {
		defer sub.Close()
		msgCh := sub.Channel()
		for {
			select {
			case <-ctx.Done():
				return
			case _, ok := <-msgCh:
				if !ok {
					return
				}
				// Non-blocking send; drop if already signaled.
				select {
				case ch <- struct{}{}:
				default:
				}
			}
		}
	}()
	return ch, nil
}

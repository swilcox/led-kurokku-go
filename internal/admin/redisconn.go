package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/swilcox/led-kurokku-go/config"
)

// slogRedisLogger adapts go-redis internal logging to slog.
type slogRedisLogger struct{}

func (slogRedisLogger) Printf(ctx context.Context, format string, v ...interface{}) {
	slog.Warn(fmt.Sprintf(format, v...), "source", "go-redis")
}

func init() {
	redis.SetLogger(slogRedisLogger{})
}

const configKey = "kurokku:config"

// dialRedis creates an ad-hoc Redis client for the given host and port.
func dialRedis(host string, port int) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", host, port),
	})
}

// TestConnection verifies Redis connectivity with a 3-second timeout.
func TestConnection(host string, port int) error {
	rdb := dialRedis(host, port)
	defer rdb.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return rdb.Ping(ctx).Err()
}

// FetchConfig reads the kurokku:config key and parses it.
// Returns (nil, false, nil) when the key is absent.
func FetchConfig(host string, port int) (*config.Config, bool, error) {
	rdb := dialRedis(host, port)
	defer rdb.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	raw, err := rdb.Get(ctx, configKey).Result()
	if err == redis.Nil {
		slog.Debug("redis config key not found", "host", host, "port", port, "key", configKey)
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("GET %s: %w", configKey, err)
	}
	cfg, err := config.Parse([]byte(raw))
	if err != nil {
		slog.Error("redis config parse failed", "host", host, "port", port, "error", err)
		return nil, false, err
	}
	slog.Debug("redis config fetched", "host", host, "port", port, "display_type", cfg.Display.Type, "widgets", len(cfg.Widgets))
	return cfg, true, nil
}

// FetchConfigJSON reads the raw JSON string of the kurokku:config key.
func FetchConfigJSON(host string, port int) (string, bool, error) {
	rdb := dialRedis(host, port)
	defer rdb.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	raw, err := rdb.Get(ctx, configKey).Result()
	if err == redis.Nil {
		slog.Debug("redis config key not found", "host", host, "port", port, "key", configKey)
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("GET %s: %w", configKey, err)
	}
	return raw, true, nil
}

// SaveConfig marshals the config and writes it to Redis.
func SaveConfig(host string, port int, cfg *config.Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	return SaveConfigJSON(host, port, string(data))
}

// FetchKey reads an arbitrary Redis key, returning its value.
// Returns ("", nil) when the key does not exist.
func FetchKey(host string, port int, key string) (string, error) {
	rdb := dialRedis(host, port)
	defer rdb.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	val, err := rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		slog.Debug("redis key not found", "host", host, "port", port, "key", key)
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("GET %s: %w", key, err)
	}
	return val, nil
}

// SaveConfigJSON writes raw JSON to the kurokku:config key.
func SaveConfigJSON(host string, port int, jsonStr string) error {
	rdb := dialRedis(host, port)
	defer rdb.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Set(ctx, configKey, jsonStr, 0).Err(); err != nil {
		return err
	}
	slog.Debug("redis config saved", "host", host, "port", port, "key", configKey, "bytes", len(jsonStr))
	return nil
}

// SendAlert writes an alert JSON to kurokku:alert:<id> with an optional TTL.
// A ttl of 0 means no expiration.
func SendAlert(host string, port int, id, alertJSON string, ttl time.Duration) error {
	rdb := dialRedis(host, port)
	defer rdb.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	key := "kurokku:alert:" + id
	if err := rdb.Set(ctx, key, alertJSON, ttl).Err(); err != nil {
		return err
	}
	slog.Debug("redis alert written", "host", host, "port", port, "key", key, "ttl", ttl)
	return nil
}

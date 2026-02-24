package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/swilcox/led-kurokku-go/config"
)

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

// FetchConfigJSON reads the raw JSON string of the kurokku:config key.
func FetchConfigJSON(host string, port int) (string, bool, error) {
	rdb := dialRedis(host, port)
	defer rdb.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	raw, err := rdb.Get(ctx, configKey).Result()
	if err == redis.Nil {
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

// SaveConfigJSON writes raw JSON to the kurokku:config key.
func SaveConfigJSON(host string, port int, jsonStr string) error {
	rdb := dialRedis(host, port)
	defer rdb.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return rdb.Set(ctx, configKey, jsonStr, 0).Err()
}

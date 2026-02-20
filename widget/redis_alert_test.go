package widget_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/swilcox/led-kurokku-go/config"
	"github.com/swilcox/led-kurokku-go/display/testutil"
	"github.com/swilcox/led-kurokku-go/widget"
)

type mockAlertFetcher struct {
	alerts []config.AlertConfig
	err    error
	deleted []string
}

func (m *mockAlertFetcher) FetchAlerts(_ context.Context) ([]config.AlertConfig, error) {
	return m.alerts, m.err
}

func (m *mockAlertFetcher) DeleteAlert(_ context.Context, id string) error {
	m.deleted = append(m.deleted, id)
	return nil
}

func TestRedisAlert_UsesRedisAlerts(t *testing.T) {
	spy := &testutil.SpyDisplay{}
	fetcher := &mockAlertFetcher{
		alerts: []config.AlertConfig{
			{ID: "r1", Message: "Hi", Priority: 1, DisplayDuration: config.Duration(time.Millisecond)},
		},
	}

	ra := &widget.RedisAlert{
		Fetcher:     fetcher,
		ScrollSpeed: 0,
	}

	ra.Run(context.Background(), spy) //nolint:errcheck

	if len(spy.Frames) == 0 {
		t.Error("expected frames from Redis alerts")
	}
}

func TestRedisAlert_FallbackOnFetchError(t *testing.T) {
	spy := &testutil.SpyDisplay{}
	fetcher := &mockAlertFetcher{err: errors.New("redis down")}

	ra := &widget.RedisAlert{
		Fetcher: fetcher,
		Fallback: []config.AlertConfig{
			{ID: "fb", Message: "Hi", Priority: 1, DisplayDuration: config.Duration(time.Millisecond)},
		},
		ScrollSpeed: 0,
	}

	ra.Run(context.Background(), spy) //nolint:errcheck

	if len(spy.Frames) == 0 {
		t.Error("expected frames from fallback alerts when Redis errors")
	}
}

func TestRedisAlert_EmptyWhenRedisReturnsNothing(t *testing.T) {
	spy := &testutil.SpyDisplay{}
	fetcher := &mockAlertFetcher{alerts: nil}

	ra := &widget.RedisAlert{
		Fetcher:     fetcher,
		Fallback:    nil,
		ScrollSpeed: 0,
	}

	ra.Run(context.Background(), spy) //nolint:errcheck

	if len(spy.Frames) != 0 {
		t.Errorf("expected no frames when Redis returns empty and no fallback, got %d", len(spy.Frames))
	}
}

func TestRedisAlert_DeleteCalledViaOnDelete(t *testing.T) {
	spy := &testutil.SpyDisplay{}
	fetcher := &mockAlertFetcher{
		alerts: []config.AlertConfig{
			{
				ID:                 "del1",
				Message:            "Hi",
				Priority:           1,
				DisplayDuration:    config.Duration(time.Millisecond),
				DeleteAfterDisplay: true,
			},
		},
	}

	ra := &widget.RedisAlert{
		Fetcher:     fetcher,
		ScrollSpeed: 0,
	}

	ra.Run(context.Background(), spy) //nolint:errcheck

	if len(fetcher.deleted) != 1 || fetcher.deleted[0] != "del1" {
		t.Errorf("expected DeleteAlert called with 'del1', got %v", fetcher.deleted)
	}
}

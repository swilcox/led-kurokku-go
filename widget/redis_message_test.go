package widget_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/swilcox/led-kurokku-go/display/testutil"
	"github.com/swilcox/led-kurokku-go/font"
	"github.com/swilcox/led-kurokku-go/framebuf"
	"github.com/swilcox/led-kurokku-go/widget"
)

type mockMessageFetcher struct {
	text string
	ok   bool
	err  error
}

func (m *mockMessageFetcher) FetchMessageText(_ context.Context, _ string) (string, bool, error) {
	return m.text, m.ok, m.err
}

func centeredFrame(text string) []byte {
	cols := font.RenderText(text)
	var f framebuf.Frame
	offset := (32 - len(cols)) / 2
	framebuf.BlitText(&f, text, offset)
	return f.Bytes()
}

func TestRedisMessage_UsesRedisText(t *testing.T) {
	spy := &testutil.SpyDisplay{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // static message; returns after one frame

	rm := &widget.RedisMessage{
		Fetcher:      &mockMessageFetcher{text: "Hi", ok: true},
		Key:          "some:key",
		FallbackText: "Lo",
		Repeats:      1,
	}
	rm.Run(ctx, spy) //nolint:errcheck

	if len(spy.Frames) == 0 {
		t.Fatal("expected at least one frame")
	}
	want := centeredFrame("Hi")
	if !bytes.Equal(spy.Frames[0], want) {
		t.Error("expected frame for Redis text 'Hi', not fallback 'Lo'")
	}
}

func TestRedisMessage_FallbackOnError(t *testing.T) {
	spy := &testutil.SpyDisplay{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	rm := &widget.RedisMessage{
		Fetcher:      &mockMessageFetcher{err: errors.New("redis down")},
		Key:          "some:key",
		FallbackText: "Hi",
		Repeats:      1,
	}
	rm.Run(ctx, spy) //nolint:errcheck

	if len(spy.Frames) == 0 {
		t.Fatal("expected at least one frame from fallback")
	}
	want := centeredFrame("Hi")
	if !bytes.Equal(spy.Frames[0], want) {
		t.Error("expected frame for fallback text 'Hi'")
	}
}

func TestRedisMessage_FallbackWhenKeyMissing(t *testing.T) {
	spy := &testutil.SpyDisplay{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	rm := &widget.RedisMessage{
		Fetcher:      &mockMessageFetcher{ok: false},
		Key:          "some:key",
		FallbackText: "Hi",
		Repeats:      1,
	}
	rm.Run(ctx, spy) //nolint:errcheck

	if len(spy.Frames) == 0 {
		t.Fatal("expected at least one frame from fallback")
	}
	want := centeredFrame("Hi")
	if !bytes.Equal(spy.Frames[0], want) {
		t.Error("expected frame for fallback text when key is missing")
	}
}

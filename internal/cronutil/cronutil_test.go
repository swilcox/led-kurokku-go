package cronutil

import (
	"testing"
	"time"
)

func TestMatchesNow_EveryMinute(t *testing.T) {
	// "* * * * *" matches every minute
	ts := time.Date(2024, 3, 15, 10, 30, 0, 0, time.UTC)
	if !MatchesNow("* * * * *", ts) {
		t.Error("expected every-minute cron to match")
	}
}

func TestMatchesNow_SpecificMinute_Match(t *testing.T) {
	// "30 10 * * *" matches at 10:30
	ts := time.Date(2024, 3, 15, 10, 30, 0, 0, time.UTC)
	if !MatchesNow("30 10 * * *", ts) {
		t.Error("expected 10:30 cron to match at 10:30")
	}
}

func TestMatchesNow_SpecificMinute_NoMatch(t *testing.T) {
	// "30 10 * * *" does not match at 10:31
	ts := time.Date(2024, 3, 15, 10, 31, 0, 0, time.UTC)
	if MatchesNow("30 10 * * *", ts) {
		t.Error("expected 10:30 cron NOT to match at 10:31")
	}
}

func TestMatchesNow_Midnight(t *testing.T) {
	ts := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
	if !MatchesNow("0 0 * * *", ts) {
		t.Error("expected midnight cron to match at 00:00")
	}
}

func TestMatchesNow_EveryFifteenMinutes(t *testing.T) {
	ts15 := time.Date(2024, 3, 15, 10, 15, 0, 0, time.UTC)
	ts16 := time.Date(2024, 3, 15, 10, 16, 0, 0, time.UTC)
	if !MatchesNow("*/15 * * * *", ts15) {
		t.Error("expected */15 to match at :15")
	}
	if MatchesNow("*/15 * * * *", ts16) {
		t.Error("expected */15 NOT to match at :16")
	}
}

func TestMatchesNow_DayOfWeek(t *testing.T) {
	// 2024-03-18 is a Monday
	monday := time.Date(2024, 3, 18, 9, 0, 0, 0, time.UTC)
	tuesday := time.Date(2024, 3, 19, 9, 0, 0, 0, time.UTC)
	if !MatchesNow("0 9 * * MON", monday) {
		t.Error("expected MON cron to match on Monday")
	}
	if MatchesNow("0 9 * * MON", tuesday) {
		t.Error("expected MON cron NOT to match on Tuesday")
	}
}

func TestMatchesNow_IgnoresSeconds(t *testing.T) {
	// Should match regardless of seconds value
	ts := time.Date(2024, 3, 15, 10, 30, 45, 123, time.UTC)
	if !MatchesNow("30 10 * * *", ts) {
		t.Error("expected cron to match regardless of seconds")
	}
}

func TestMatchesNow_InvalidExpression(t *testing.T) {
	ts := time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC)
	if MatchesNow("invalid cron", ts) {
		t.Error("expected invalid cron to return false")
	}
}

func TestMatchesNow_EmptyExpression(t *testing.T) {
	ts := time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC)
	if MatchesNow("", ts) {
		t.Error("expected empty cron to return false")
	}
}

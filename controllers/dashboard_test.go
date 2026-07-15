package controllers

import (
	"testing"
	"time"
)

func TestParseDateRange(t *testing.T) {
	now := time.Date(2026, time.July, 18, 14, 30, 0, 0, time.UTC)
	endOfDay := time.Date(2026, time.July, 18, 23, 59, 59, 0, time.UTC)

	tests := []struct {
		name, period, startParam, endParam string
		wantStart, wantEnd                 time.Time
	}{
		{name: "default seven days", wantStart: endOfDay.AddDate(0, 0, -7), wantEnd: endOfDay},
		{name: "today", period: "today", wantStart: now.Add(-24 * time.Hour), wantEnd: now},
		{name: "thirty days", period: "30d", wantStart: endOfDay.AddDate(0, 0, -30), wantEnd: endOfDay},
		{name: "current month", period: "month", wantStart: time.Date(2026, time.July, 1, 0, 0, 0, 0, time.UTC), wantEnd: endOfDay},
		{name: "custom", period: "custom", startParam: "2026-07-02", endParam: "2026-07-05", wantStart: time.Date(2026, time.July, 2, 0, 0, 0, 0, time.UTC), wantEnd: time.Date(2026, time.July, 5, 23, 59, 59, 0, time.UTC)},
		{name: "invalid custom", period: "custom", startParam: "bad", endParam: "2026-07-05", wantStart: endOfDay.AddDate(0, 0, -7), wantEnd: endOfDay},
		{name: "reversed custom", period: "custom", startParam: "2026-07-06", endParam: "2026-07-05", wantStart: endOfDay.AddDate(0, 0, -7), wantEnd: endOfDay},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			start, end := parseDateRangeAt(now, test.period, test.startParam, test.endParam)
			if !start.Equal(test.wantStart) || !end.Equal(test.wantEnd) {
				t.Fatalf("range = %s to %s, want %s to %s", start, end, test.wantStart, test.wantEnd)
			}
		})
	}
}

func TestBucketAndPreviousPeriodBoundaries(t *testing.T) {
	start := time.Date(2026, time.July, 1, 0, 0, 0, 0, time.UTC)

	if got := chooseBucket(start, start.Add(48*time.Hour)); got != "hour" {
		t.Fatalf("48 hour bucket = %q, want hour", got)
	}
	if got := chooseBucket(start, start.Add(48*time.Hour+time.Second)); got != "day" {
		t.Fatalf("over 48 hour bucket = %q, want day", got)
	}

	end := start.Add(7 * 24 * time.Hour)
	prevStart, prevEnd := previousPeriodRange(start, end)
	if !prevEnd.Equal(start.Add(-time.Second)) {
		t.Fatalf("previous end = %s, want %s", prevEnd, start.Add(-time.Second))
	}
	if prevEnd.Sub(prevStart) != end.Sub(start) {
		t.Fatalf("previous duration = %s, want %s", prevEnd.Sub(prevStart), end.Sub(start))
	}
}

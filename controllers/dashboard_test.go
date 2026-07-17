package controllers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"palantir/database"
	"palantir/models"
	"palantir/models/factories"
)

func TestDashboardIncludesBounceRateSeries(t *testing.T) {
	setTestConfig(t)
	db := controllerTestCluster.NewTestDB(t, database.Migrations, "migrations")
	ctx := context.Background()
	owner, err := factories.CreateUser(ctx, db.Executor())
	if err != nil {
		t.Fatal(err)
	}
	website, err := models.Website.Create(ctx, db.Executor(), models.CreateWebsiteData{UserID: owner.ID, Name: "Analytics", Domain: "analytics.example"})
	if err != nil {
		t.Fatal(err)
	}
	for _, visitor := range []string{"returning", "returning", "bounce"} {
		if _, err := models.Pageview.Create(ctx, db.Executor(), models.CreatePageviewData{WebsiteID: website.ID, URL: "/", VisitorHash: visitor}); err != nil {
			t.Fatal(err)
		}
	}

	e := authenticatedEcho(owner)
	e.GET("/websites/:id/dashboard", NewDashboard(db).Show)
	req := httptest.NewRequest(http.MethodGet, "/websites/"+website.ID.String()+"/dashboard?period=today", nil)
	req.Header.Set("X-Inertia", "true")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("dashboard status = %d; body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, `"BounceRateOverTime"`) || !strings.Contains(body, `"rate":50`) {
		t.Fatalf("dashboard props missing 50%% bounce-rate point: %s", body)
	}
	if !strings.Contains(body, `"auth":{"email":"`+owner.Email+`"}`) || strings.Contains(body, `"Password"`) {
		t.Fatalf("dashboard props did not expose only the authenticated email: %s", body)
	}
}

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

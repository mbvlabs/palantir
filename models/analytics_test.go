package models_test

import (
	"context"
	"testing"
	"time"

	"palantir/database"
	"palantir/models"
	"palantir/models/factories"
)

func TestCollectionDataAppearsInDashboardAggregates(t *testing.T) {
	db := modelTestCluster.NewTestDB(t, database.Migrations, "migrations")
	ctx := context.Background()
	owner, err := factories.CreateUser(ctx, db.Executor())
	if err != nil {
		t.Fatal(err)
	}
	website, err := models.Website.Create(ctx, db.Executor(), models.CreateWebsiteData{UserID: owner.ID, Name: "Analytics", Domain: "analytics.example"})
	if err != nil {
		t.Fatal(err)
	}

	_, err = models.Pageview.Create(ctx, db.Executor(), models.CreatePageviewData{
		WebsiteID: website.ID, URL: "/pricing", Browser: "Firefox", OS: "Linux", Device: "desktop",
		VisitorHash: "visitor-1", CountryCode: "NL", CountryName: "Netherlands", City: "Amsterdam",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = models.Event.Create(ctx, db.Executor(), models.CreateEventData{
		WebsiteID: website.ID, URL: "/pricing", EventName: "signup", VisitorHash: "visitor-1",
		CountryCode: "NL", CountryName: "Netherlands", City: "Amsterdam",
	})
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC()
	stats, err := models.DashboardStatsFor(ctx, db.Executor(), website.ID, now.Add(-time.Hour), now.Add(time.Hour), now.Add(-3*time.Hour), now.Add(-2*time.Hour), "hour")
	if err != nil {
		t.Fatal(err)
	}
	if stats.TotalPageviews != 1 || stats.TotalUniqueVisitors != 1 {
		t.Fatalf("totals = pageviews %d visitors %d", stats.TotalPageviews, stats.TotalUniqueVisitors)
	}
	if len(stats.TopEvents) != 1 || stats.TopEvents[0].Name != "signup" {
		t.Fatalf("events = %+v", stats.TopEvents)
	}
	if len(stats.Devices) != 1 || stats.Devices[0].Name != "desktop" {
		t.Fatalf("devices = %+v", stats.Devices)
	}
	if len(stats.TopCountries) != 1 || stats.TopCountries[0].Code != "NL" || len(stats.TopCities) != 1 {
		t.Fatalf("geo = countries %+v cities %+v", stats.TopCountries, stats.TopCities)
	}
	if len(stats.PageviewsOverTime) == 0 || len(stats.EventsOverTime) == 0 {
		t.Fatalf("series missing: pageviews=%+v events=%+v", stats.PageviewsOverTime, stats.EventsOverTime)
	}
}

package controllers

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"palantir/config"
	"palantir/database"
	"palantir/models"
	"palantir/models/factories"
	"palantir/services"

	"github.com/labstack/echo/v5"
)

type fixedGeoResolver struct{}

func (fixedGeoResolver) Resolve(string) (services.GeoResult, error) {
	return services.GeoResult{CountryCode: "NL", CountryName: "Netherlands", City: "Amsterdam", Region: "North Holland"}, nil
}

func TestCollectPageviewAndEventThroughHTTP(t *testing.T) {
	setTestConfig(t)
	db := controllerTestCluster.NewTestDB(t, database.Migrations, "migrations")
	ctx := context.Background()
	owner, err := factories.CreateUser(ctx, db.Executor())
	if err != nil {
		t.Fatal(err)
	}
	website, err := models.Website.Create(ctx, db.Executor(), models.CreateWebsiteData{UserID: owner.ID, Name: "Example", Domain: "example.com"})
	if err != nil {
		t.Fatal(err)
	}

	collect := NewCollect(db, fixedGeoResolver{}, config.NewConfig())
	e := echo.New()
	e.POST("/api/collect", collect.Create)

	for _, payload := range []string{
		`{"website_id":"` + website.ID.String() + `","type":"pageview","url":"/pricing","screen_width":1440,"language":"en"}`,
		`{"website_id":"` + website.ID.String() + `","type":"event","url":"/pricing","event_name":"signup","event_data":{"plan":"pro"}}`,
	} {
		req := httptest.NewRequest(http.MethodPost, "/api/collect", bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) Gecko/20100101 Firefox/128.0")
		req.RemoteAddr = "192.0.2.10:1234"
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("collect status = %d, want %d; body=%s", rec.Code, http.StatusNoContent, rec.Body.String())
		}
	}

	stats, err := models.DashboardStatsFor(ctx, db.Executor(), website.ID, time.Now().Add(-time.Hour), time.Now().Add(time.Hour), time.Now().Add(-3*time.Hour), time.Now().Add(-2*time.Hour), "hour")
	if err != nil {
		t.Fatal(err)
	}
	if stats.TotalPageviews != 1 || len(stats.TopEvents) != 1 || stats.TopEvents[0].Name != "signup" || len(stats.TopCities) != 1 || stats.TopCities[0].Name != "Amsterdam" {
		t.Fatalf("dashboard stats = %+v", stats)
	}

	dashboardEcho := authenticatedEcho(owner)
	dashboardEcho.GET("/websites/:id/dashboard", NewDashboard(db).Show)
	req := httptest.NewRequest(http.MethodGet, "/websites/"+website.ID.String()+"/dashboard", nil)
	req.Header.Set("X-Inertia", "true")
	rec := httptest.NewRecorder()
	dashboardEcho.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("dashboard status = %d; body=%s", rec.Code, rec.Body.String())
	}
	for _, expected := range []string{`"TotalPageviews":1`, `"name":"signup"`, `"name":"desktop"`, `"name":"Amsterdam"`} {
		if !strings.Contains(rec.Body.String(), expected) {
			t.Fatalf("dashboard props missing %s: %s", expected, rec.Body.String())
		}
	}
}

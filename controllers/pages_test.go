package controllers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"palantir/database"
	"palantir/models"
	"palantir/models/factories"
	"palantir/router/cookies"
	"palantir/router/routes"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/v5/session"
	"github.com/labstack/echo/v5"
)

func TestHomeSelectsLastViewedWebsite(t *testing.T) {
	setTestConfig(t)
	db := controllerTestCluster.NewTestDB(t, database.Migrations, "migrations")
	ctx := context.Background()
	owner, err := factories.CreateUser(ctx, db.Executor())
	if err != nil {
		t.Fatal(err)
	}
	first, err := models.Website.Create(ctx, db.Executor(), models.CreateWebsiteData{UserID: owner.ID, Name: "First", Domain: "first.example"})
	if err != nil {
		t.Fatal(err)
	}
	last, err := models.Website.Create(ctx, db.Executor(), models.CreateWebsiteData{UserID: owner.ID, Name: "Last", Domain: "last.example"})
	if err != nil {
		t.Fatal(err)
	}

	for _, test := range []struct {
		name        string
		lastWebsite uuid.UUID
		want        string
	}{
		{name: "last viewed", lastWebsite: last.ID, want: routes.WebsiteDashboard.URL(last.ID)},
		{name: "first website fallback", want: routes.WebsiteDashboard.URL(first.ID)},
		{name: "foreign website fallback", lastWebsite: uuid.New(), want: routes.WebsiteDashboard.URL(first.ID)},
	} {
		t.Run(test.name, func(t *testing.T) {
			e := echoWithUserAndLastWebsite(owner, test.lastWebsite)
			e.GET("/", NewPages(db).Home)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
			if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != test.want {
				t.Fatalf("status=%d location=%q, want 303 %q", rec.Code, rec.Header().Get("Location"), test.want)
			}
		})
	}
}

func TestHomeWithoutWebsitesOpensAddWebsite(t *testing.T) {
	setTestConfig(t)
	db := controllerTestCluster.NewTestDB(t, database.Migrations, "migrations")
	owner, err := factories.CreateUser(context.Background(), db.Executor())
	if err != nil {
		t.Fatal(err)
	}
	e := echoWithUserAndLastWebsite(owner, uuid.Nil)
	e.GET("/", NewPages(db).Home)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != routes.WebsiteNew.URL() {
		t.Fatalf("status=%d location=%q, want add website", rec.Code, rec.Header().Get("Location"))
	}
}

func TestHomeHeadRemainsAReadyCheck(t *testing.T) {
	e := echo.New()
	e.HEAD("/", NewPages(nil).Home)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodHead, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d, want 200", rec.Code)
	}
}

func echoWithUserAndLastWebsite(user models.UserEntity, websiteID uuid.UUID) *echo.Echo {
	e := echo.New()
	e.Use(session.Middleware(sessions.NewCookieStore([]byte("01234567890123456789012345678901"))))
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if err := cookies.CreateAppSession(c, user); err != nil {
				return err
			}
			if websiteID != uuid.Nil {
				if err := cookies.SetLastWebsite(c, websiteID); err != nil {
					return err
				}
			}
			return next(c)
		}
	})
	return e
}

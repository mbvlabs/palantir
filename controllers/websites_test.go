package controllers

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"palantir/database"
	"palantir/models"
	"palantir/models/factories"
	"palantir/router/cookies"
	"palantir/router/routes"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/v5/session"
	"github.com/labstack/echo/v5"
)

func TestWebsiteLifecycleThroughHTTP(t *testing.T) {
	setTestConfig(t)
	db := controllerTestCluster.NewTestDB(t, database.Migrations, "migrations")
	ctx := context.Background()
	owner, err := factories.CreateUser(ctx, db.Executor())
	if err != nil {
		t.Fatal(err)
	}
	websites := NewWebsites(db)
	e := authenticatedEcho(owner)
	e.GET("/websites", websites.Index)
	e.POST("/websites", websites.Create)
	e.PUT("/websites/:id", websites.Update)
	e.DELETE("/websites/:id", websites.Destroy)

	request := func(method, path, body string) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Inertia", "true")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		return rec
	}

	if rec := request(http.MethodPost, "/websites", `{"name":"Example","domain":"example.com"}`); rec.Code != http.StatusSeeOther {
		t.Fatalf("create status = %d body=%s", rec.Code, rec.Body.String())
	}
	owned, err := models.Website.AllOwned(ctx, db.Executor(), owner.ID)
	if err != nil || len(owned) != 1 {
		t.Fatalf("created websites = %+v, err=%v", owned, err)
	}
	website := owned[0]

	if rec := request(http.MethodGet, "/websites", ""); rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != routes.WebsiteDashboard.URL(website.ID) {
		t.Fatalf("index status=%d location=%q", rec.Code, rec.Header().Get("Location"))
	}
	if rec := request(http.MethodPut, "/websites/"+website.ID.String(), `{"name":"Updated","domain":"updated.example"}`); rec.Code != http.StatusSeeOther {
		t.Fatalf("update status=%d body=%s", rec.Code, rec.Body.String())
	}
	updated, err := models.Website.FindOwned(ctx, db.Executor(), website.ID, owner.ID)
	if err != nil || updated.Name != "Updated" {
		t.Fatalf("updated website=%+v err=%v", updated, err)
	}

	before := len(owned)
	if rec := request(http.MethodPost, "/websites", `{"name":"","domain":"not a domain"}`); rec.Code != http.StatusOK {
		t.Fatalf("invalid create status=%d body=%s", rec.Code, rec.Body.String())
	}
	owned, err = models.Website.AllOwned(ctx, db.Executor(), owner.ID)
	if err != nil || len(owned) != before {
		t.Fatalf("invalid create wrote row: websites=%d err=%v", len(owned), err)
	}

	if rec := request(http.MethodDelete, "/websites/"+website.ID.String(), ""); rec.Code != http.StatusSeeOther {
		t.Fatalf("delete status=%d body=%s", rec.Code, rec.Body.String())
	}
	if _, err := models.Website.FindOwned(ctx, db.Executor(), website.ID, owner.ID); err == nil {
		t.Fatal("website still exists after delete")
	}
}

func TestWebsiteHandlersHideForeignRecords(t *testing.T) {
	setTestConfig(t)
	db := controllerTestCluster.NewTestDB(t, database.Migrations, "migrations")
	ctx := context.Background()
	owner, err := factories.CreateUser(ctx, db.Executor())
	if err != nil {
		t.Fatal(err)
	}
	other, err := factories.CreateUser(ctx, db.Executor())
	if err != nil {
		t.Fatal(err)
	}
	website, err := models.Website.Create(ctx, db.Executor(), models.CreateWebsiteData{UserID: owner.ID, Name: "Private", Domain: "private.example"})
	if err != nil {
		t.Fatal(err)
	}

	websites := NewWebsites(db)
	dashboard := NewDashboard(db)
	e := authenticatedEcho(other)
	e.GET("/websites/:id", websites.Show)
	e.GET("/websites/:id/edit", websites.Edit)
	e.PUT("/websites/:id", websites.Update)
	e.DELETE("/websites/:id", websites.Destroy)
	e.GET("/websites/:id/dashboard", dashboard.Show)

	for _, test := range []struct{ method, suffix, body string }{
		{http.MethodGet, "", ""},
		{http.MethodGet, "/edit", ""},
		{http.MethodPut, "", `{"name":"Stolen","domain":"stolen.example"}`},
		{http.MethodDelete, "", ""},
		{http.MethodGet, "/dashboard", ""},
	} {
		req := httptest.NewRequest(test.method, "/websites/"+website.ID.String()+test.suffix, strings.NewReader(test.body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Inertia", "true")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("%s %s status=%d, want 404; body=%s", test.method, test.suffix, rec.Code, rec.Body.String())
		}
	}
}

func authenticatedEcho(user models.UserEntity) *echo.Echo {
	e := echo.New()
	store := sessions.NewCookieStore([]byte("01234567890123456789012345678901"))
	e.Use(session.Middleware(store))
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if err := cookies.CreateAppSession(c, user); err != nil {
				return err
			}
			return next(c)
		}
	})
	return e
}

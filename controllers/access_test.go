package controllers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"palantir/config"
	"palantir/controllers/api"
	"palantir/router"
	"palantir/services"
	"palantir/telemetry"
)

type emptyGeoResolver struct{}

func (emptyGeoResolver) Resolve(string) (services.GeoResult, error) { return services.GeoResult{}, nil }

func TestHTTPAccessMatrix(t *testing.T) {
	setTestConfig(t)
	cfg := config.NewConfig()
	r, err := router.New(cfg, &telemetry.Telemetry{})
	if err != nil {
		t.Fatal(err)
	}
	cache, err := NewCacheBuilder[string]().WithSize(2).Build()
	if err != nil {
		t.Fatal(err)
	}
	registrars := []func(*router.Router) error{
		NewPages(nil).RegisterRoutes,
		NewAssets(cache).RegisterRoutes,
		api.NewAPI(nil).RegisterRoutes,
		NewSessions(structIdentity()).RegisterRoutes,
		NewResetPasswords(structIdentity()).RegisterRoutes,
		NewWebsites(nil).RegisterRoutes,
		NewDashboard(nil).RegisterRoutes,
		NewCollect(nil, emptyGeoResolver{}, cfg).RegisterRoutes,
		NewTracking().RegisterRoutes,
	}
	for _, register := range registrars {
		if err := register(r); err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name, method, path string
		want               int
	}{
		{name: "sign in get public", method: http.MethodGet, path: "/users/sign-in", want: http.StatusOK},
		{name: "sign in head public", method: http.MethodHead, path: "/users/sign-in", want: http.StatusOK},
		{name: "sign in post public", method: http.MethodPost, path: "/users/sign-in", want: http.StatusForbidden},
		{name: "password request get public", method: http.MethodGet, path: "/users/password/new", want: http.StatusOK},
		{name: "password request post public", method: http.MethodPost, path: "/users/password", want: http.StatusForbidden},
		{name: "password token public", method: http.MethodGet, path: "/users/password/token/edit", want: http.StatusOK},
		{name: "root private", method: http.MethodGet, path: "/", want: http.StatusSeeOther},
		{name: "websites private", method: http.MethodGet, path: "/websites", want: http.StatusSeeOther},
		{name: "dashboard private", method: http.MethodGet, path: "/websites/00000000-0000-0000-0000-000000000000/dashboard", want: http.StatusSeeOther},
		{name: "sign out private", method: http.MethodDelete, path: "/users/sign-out", want: http.StatusSeeOther},
		{name: "registration removed", method: http.MethodGet, path: "/users/sign-up", want: http.StatusNotFound},
		{name: "confirmation removed", method: http.MethodGet, path: "/users/confirmation/new", want: http.StatusNotFound},
		{name: "collection public", method: http.MethodOptions, path: "/api/collect", want: http.StatusNoContent},
		{name: "tracking public", method: http.MethodGet, path: "/t/script.js", want: http.StatusOK},
		{name: "health public", method: http.MethodGet, path: "/api/health", want: http.StatusOK},
		{name: "asset public", method: http.MethodGet, path: "/robots.txt", want: http.StatusOK},
		{name: "unknown", method: http.MethodGet, path: "/missing", want: http.StatusNotFound},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(test.method, test.path, nil)
			rec := httptest.NewRecorder()
			r.Handler.ServeHTTP(rec, req)
			if rec.Code != test.want {
				t.Fatalf("status = %d, want %d; location=%q body=%s", rec.Code, test.want, rec.Header().Get("Location"), rec.Body.String())
			}
		})
	}
}

func structIdentity() services.Identity { return services.Identity{} }

func setTestConfig(t *testing.T) {
	t.Helper()
	values := map[string]string{
		"SESSION_KEY":            "36bdb5887212e20c32fd29344ac9ec15a84ede08b0f46af5013ca330442415d477306b11d0c20857cdbd24ae5779fa081b34c05149a4280ef8d9050dc06a5342",
		"SESSION_ENCRYPTION_KEY": "93b059b0787c043d21c57652cc45f8710d2d733467d7a7165e9cd1de23f65e44", "TOKEN_SIGNING_KEY": "test", "VISITOR_HASH_SALT": "test-salt", "PEPPER": "test",
		"CORS_ALLOWED_ORIGINS": "", "CSRF_TRUSTED_ORIGINS": "", "DB_PORT": "5432", "DB_HOST": "localhost", "DB_NAME": "test", "DB_USER": "test", "DB_PASSWORD": "test", "DB_KIND": "postgres", "DB_SSL_MODE": "disable",
		"AWS_SES_ACCESS_KEY_ID": "", "AWS_SES_SECRET_ACCESS_KEY": "", "AWS_SES_CONFIGURATION_SET": "",
	}
	for key, value := range values {
		t.Setenv(key, value)
	}
}

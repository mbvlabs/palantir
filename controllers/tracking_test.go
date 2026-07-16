package controllers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
)

func TestTrackingScriptSupportsDataAttributeClicks(t *testing.T) {
	e := echo.New()
	e.GET("/t/script.js", NewTracking().Script)

	req := httptest.NewRequest(http.MethodGet, "/t/script.js", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if contentType := rec.Header().Get("Content-Type"); contentType != "application/javascript; charset=utf-8" {
		t.Fatalf("Content-Type = %q", contentType)
	}
	if cacheControl := rec.Header().Get("Cache-Control"); cacheControl != "public, max-age=86400" {
		t.Fatalf("Cache-Control = %q", cacheControl)
	}
	for _, expected := range []string{"window.palantir", "[data-palantir-event]", ".closest("} {
		if !strings.Contains(rec.Body.String(), expected) {
			t.Fatalf("tracking script missing %q: %s", expected, rec.Body.String())
		}
	}
}

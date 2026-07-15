package controllers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"palantir/config"
	"palantir/internal/storage"
	"palantir/models"
	"palantir/router"
	"palantir/router/routes"
	"palantir/services"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/mssola/useragent"
)

type Collect struct {
	db   storage.Pool
	geo  services.GeoResolver
	salt string
}

func NewCollect(db storage.Pool, geo services.GeoResolver, cfg config.Config) Collect {
	return Collect{db: db, geo: geo, salt: cfg.App.VisitorHashSalt}
}

func (c Collect) RegisterRoutes(r *router.Router) error {
	var errs []error
	for _, route := range []echo.Route{
		{Method: http.MethodOptions, Path: routes.CollectCreate.Path(), Name: routes.CollectCreate.Name() + ".options", Handler: c.Preflight},
		{Method: http.MethodPost, Path: routes.CollectCreate.Path(), Name: routes.CollectCreate.Name(), Handler: c.Create},
	} {
		if _, err := r.AddRoute(route); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

type collectPayload struct {
	WebsiteID   string          `json:"website_id"`
	Type        string          `json:"type"`
	URL         string          `json:"url"`
	Referrer    string          `json:"referrer"`
	ScreenWidth int32           `json:"screen_width"`
	Language    string          `json:"language"`
	EventName   string          `json:"event_name"`
	EventData   json.RawMessage `json:"event_data"`
}

func collectCORS(c *echo.Context) {
	origin := c.Request().Header.Get("Origin")
	if origin == "" {
		origin = "*"
	} else {
		c.Response().Header().Add("Vary", "Origin")
	}
	c.Response().Header().Set("Access-Control-Allow-Origin", origin)
	c.Response().Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	c.Response().Header().Set("Access-Control-Allow-Headers", "Content-Type")
	c.Response().Header().Set("Access-Control-Max-Age", "300")
}

func (Collect) Preflight(c *echo.Context) error {
	collectCORS(c)
	return c.NoContent(http.StatusNoContent)
}

func (collector Collect) Create(c *echo.Context) error {
	collectCORS(c)
	request := c.Request()
	request.Body = http.MaxBytesReader(c.Response(), request.Body, 64<<10)
	var payload collectPayload
	decoder := json.NewDecoder(request.Body)
	if err := decoder.Decode(&payload); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return c.NoContent(http.StatusBadRequest)
	}
	websiteID, err := uuid.Parse(payload.WebsiteID)
	if err != nil || !validTrackedURL(payload.URL) || (payload.Type != "pageview" && payload.Type != "event") {
		return c.NoContent(http.StatusBadRequest)
	}
	if payload.Type == "event" && (strings.TrimSpace(payload.EventName) == "" || len(payload.EventName) > 255 || (len(payload.EventData) > 0 && !json.Valid(payload.EventData))) {
		return c.NoContent(http.StatusBadRequest)
	}
	if _, err := models.Website.Find(request.Context(), collector.db.Executor(), websiteID); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	ua := useragent.New(request.UserAgent())
	browser, _ := ua.Browser()
	ip := clientIP(request)
	visitorHash := computeVisitorHash(websiteID, ip, request.UserAgent(), collector.salt, time.Now().UTC())
	geo, _ := collector.geo.Resolve(ip)
	if payload.Type == "pageview" {
		_, err = models.Pageview.Create(request.Context(), collector.db.Executor(), models.CreatePageviewData{
			WebsiteID: websiteID, URL: payload.URL, Referrer: payload.Referrer, Browser: browser,
			OS: ua.OS(), Device: parseDevice(ua), Language: payload.Language, ScreenWidth: payload.ScreenWidth,
			VisitorHash: visitorHash, CountryCode: geo.CountryCode, CountryName: geo.CountryName, City: geo.City, Region: geo.Region,
		})
	} else {
		_, err = models.Event.Create(request.Context(), collector.db.Executor(), models.CreateEventData{
			WebsiteID: websiteID, URL: payload.URL, EventName: strings.TrimSpace(payload.EventName), EventData: payload.EventData,
			VisitorHash: visitorHash, CountryCode: geo.CountryCode, CountryName: geo.CountryName, City: geo.City, Region: geo.Region,
		})
	}
	if err != nil {
		slog.ErrorContext(request.Context(), "failed to collect analytics", "error", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusNoContent)
}

func validTrackedURL(value string) bool {
	if value == "" || len(value) > 2048 {
		return false
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return false
	}
	if parsed.IsAbs() {
		return (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != ""
	}
	return strings.HasPrefix(value, "/")
}

func parseDevice(ua *useragent.UserAgent) string {
	if ua.Bot() {
		return "bot"
	}
	if ua.Mobile() {
		return "mobile"
	}
	platform := strings.ToLower(ua.Platform())
	if strings.Contains(platform, "ipad") || strings.Contains(platform, "tablet") {
		return "tablet"
	}
	return "desktop"
}

func clientIP(request *http.Request) string {
	if forwarded := request.Header.Get("X-Forwarded-For"); forwarded != "" {
		return strings.TrimSpace(strings.SplitN(forwarded, ",", 2)[0])
	}
	host, _, err := net.SplitHostPort(request.RemoteAddr)
	if err == nil {
		return host
	}
	return request.RemoteAddr
}

func computeVisitorHash(websiteID uuid.UUID, ip, userAgent, salt string, day time.Time) string {
	hash := sha256.Sum256([]byte(websiteID.String() + ip + userAgent + day.UTC().Format("2006-01-02") + salt))
	return hex.EncodeToString(hash[:])
}

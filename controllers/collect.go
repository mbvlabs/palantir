package controllers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"palantir/internal/storage"
	"palantir/models"
	"palantir/services"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/mssola/useragent"
)

type Collect struct {
	db  storage.Pool
	geo services.GeoResolver
}

func NewCollect(db storage.Pool, geo services.GeoResolver) Collect {
	return Collect{db: db, geo: geo}
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

func setCollectCORSHeaders(etx *echo.Context) {
	headers := etx.Response().Header()
	origin := etx.Request().Header.Get("Origin")

	if origin == "" {
		headers.Set("Access-Control-Allow-Origin", "*")
	} else {
		headers.Set("Access-Control-Allow-Origin", origin)
		headers.Set("Vary", "Origin")
	}

	headers.Set("Access-Control-Allow-Credentials", "true")
	headers.Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	headers.Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept")
	headers.Set("Access-Control-Max-Age", "300")
}

func (c Collect) Preflight(etx *echo.Context) error {
	setCollectCORSHeaders(etx)
	return etx.NoContent(http.StatusNoContent)
}

func (c Collect) Create(etx *echo.Context) error {
	setCollectCORSHeaders(etx)

	var payload collectPayload
	if err := etx.Bind(&payload); err != nil {
		return etx.NoContent(http.StatusBadRequest)
	}

	websiteID, err := uuid.Parse(payload.WebsiteID)
	if err != nil {
		return etx.NoContent(http.StatusBadRequest)
	}

	if payload.URL == "" {
		return etx.NoContent(http.StatusBadRequest)
	}

	ctx := etx.Request().Context()

	_, err = models.FindWebsite(ctx, c.db.Conn(), websiteID)
	if err != nil {
		return etx.NoContent(http.StatusBadRequest)
	}

	ua := useragent.New(etx.Request().UserAgent())
	browserName, _ := ua.Browser()
	osInfo := ua.OS()
	device := parseDevice(ua)

	ip := clientIP(etx.Request())
	visitorHash := computeVisitorHash(websiteID, ip, etx.Request().UserAgent())

	geo, _ := c.geo.Resolve(ip)

	switch payload.Type {
	case "pageview":
		_, err = models.CreatePageview(ctx, c.db.Conn(), models.CreatePageviewData{
			WebsiteID:   websiteID,
			URL:         payload.URL,
			Referrer:    payload.Referrer,
			Browser:     browserName,
			OS:          osInfo,
			Device:      device,
			Language:    payload.Language,
			ScreenWidth: payload.ScreenWidth,
			VisitorHash: visitorHash,
			CountryCode: geo.CountryCode,
			CountryName: geo.CountryName,
			City:        geo.City,
			Region:      geo.Region,
		})
		if err != nil {
			slog.ErrorContext(ctx, "failed to create pageview", "error", err)
			return etx.NoContent(http.StatusInternalServerError)
		}

	case "event":
		if payload.EventName == "" {
			return etx.NoContent(http.StatusBadRequest)
		}
		_, err = models.CreateEvent(ctx, c.db.Conn(), models.CreateEventData{
			WebsiteID:   websiteID,
			URL:         payload.URL,
			EventName:   payload.EventName,
			EventData:   payload.EventData,
			VisitorHash: visitorHash,
			CountryCode: geo.CountryCode,
			CountryName: geo.CountryName,
			City:        geo.City,
			Region:      geo.Region,
		})
		if err != nil {
			slog.ErrorContext(ctx, "failed to create event", "error", err)
			return etx.NoContent(http.StatusInternalServerError)
		}

	default:
		return etx.NoContent(http.StatusBadRequest)
	}

	return etx.NoContent(http.StatusOK)
}

func parseDevice(ua *useragent.UserAgent) string {
	if ua.Mobile() {
		return "mobile"
	}
	if ua.Bot() {
		return "bot"
	}
	platform := strings.ToLower(ua.Platform())
	if strings.Contains(platform, "ipad") || strings.Contains(platform, "tablet") {
		return "tablet"
	}
	return "desktop"
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.SplitN(xff, ",", 2)
		return strings.TrimSpace(parts[0])
	}
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	return host
}

func computeVisitorHash(websiteID uuid.UUID, ip string, userAgent string) string {
	salt := os.Getenv("VISITOR_HASH_SALT")
	day := time.Now().UTC().Format("2006-01-02")
	h := sha256.Sum256([]byte(websiteID.String() + ip + userAgent + day + salt))
	return hex.EncodeToString(h[:])
}

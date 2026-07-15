package controllers

import (
	"net/http"
	"time"

	"palantir/internal/inertia"
	"palantir/internal/storage"
	"palantir/models"
	"palantir/router"
	"palantir/router/cookies"
	"palantir/router/middleware"
	"palantir/router/routes"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

type Dashboard struct{ db storage.Pool }

func NewDashboard(db storage.Pool) Dashboard { return Dashboard{db: db} }

func (d Dashboard) RegisterRoutes(r *router.Router) error {
	_, err := r.AddRoute(echo.Route{
		Method: http.MethodGet, Path: routes.WebsiteDashboard.Path(), Name: routes.WebsiteDashboard.Name(),
		Handler: d.Show, Middlewares: []echo.MiddlewareFunc{middleware.AuthOnly},
	})
	return err
}

func (d Dashboard) Show(c *echo.Context) error {
	websiteID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return notFound(c)
	}
	ownerID := cookies.ExtractFromCookieApp(c).UserID
	website, err := models.Website.FindOwned(c.Request().Context(), d.db.Executor(), websiteID, ownerID)
	if err != nil {
		return websiteError(c, err)
	}
	if err := cookies.SetLastWebsite(c, website.ID); err != nil {
		return internalError(c, err)
	}
	websites, err := models.Website.AllOwned(c.Request().Context(), d.db.Executor(), ownerID)
	if err != nil {
		return internalError(c, err)
	}
	period := c.QueryParam("period")
	if period == "" {
		period = "7d"
	}
	startParam, endParam := c.QueryParam("start"), c.QueryParam("end")
	start, end := parseDateRange(period, startParam, endParam)
	previousStart, previousEnd := previousPeriodRange(start, end)
	bucket := chooseBucket(start, end)
	stats, err := models.DashboardStatsFor(c.Request().Context(), d.db.Executor(), websiteID, start, end, previousStart, previousEnd, bucket)
	if err != nil {
		return internalError(c, err)
	}
	return inertia.Page(c, "Dashboard/Show", inertia.Props{
		"website": website, "websites": websites, "stats": stats, "period": period,
		"start": startParam, "end": endParam, "bucket": bucket,
	})
}

func parseDateRange(period, startParam, endParam string) (time.Time, time.Time) {
	return parseDateRangeAt(time.Now().UTC(), period, startParam, endParam)
}

func parseDateRangeAt(now time.Time, period, startParam, endParam string) (time.Time, time.Time) {
	now = now.UTC()
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, time.UTC)
	if period == "custom" {
		start, startErr := time.Parse("2006-01-02", startParam)
		end, endErr := time.Parse("2006-01-02", endParam)
		if startErr == nil && endErr == nil && !start.After(end) {
			return start, time.Date(end.Year(), end.Month(), end.Day(), 23, 59, 59, 0, time.UTC)
		}
	}
	switch period {
	case "today":
		return now.Add(-24 * time.Hour), now
	case "30d":
		return endOfDay.AddDate(0, 0, -30), endOfDay
	case "month":
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC), endOfDay
	default:
		return endOfDay.AddDate(0, 0, -7), endOfDay
	}
}

func chooseBucket(start, end time.Time) string {
	if end.Sub(start) <= 48*time.Hour {
		return "hour"
	}
	return "day"
}

func previousPeriodRange(start, end time.Time) (time.Time, time.Time) {
	previousEnd := start.Add(-time.Second)
	return previousEnd.Add(-end.Sub(start)), previousEnd
}

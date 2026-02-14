package controllers

import (
	"fmt"
	"math"
	"net/http"
	"time"

	"palantir/internal/hypermedia"
	"palantir/internal/storage"
	"palantir/models"
	"palantir/router/cookies"
	"palantir/views"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

type Dashboard struct {
	db storage.Pool
}

func NewDashboard(db storage.Pool) Dashboard {
	return Dashboard{db: db}
}

func (d Dashboard) Show(etx *echo.Context) error {
	websiteID, err := uuid.Parse(etx.Param("id"))
	if err != nil {
		return render(etx, views.BadRequest())
	}

	app := cookies.GetApp(etx)
	ctx := etx.Request().Context()

	website, err := models.FindWebsite(ctx, d.db.Conn(), websiteID)
	if err != nil {
		return render(etx, views.NotFound())
	}

	if website.UserID != app.UserID {
		return render(etx, views.NotFound())
	}

	period := etx.QueryParam("period")
	startParam := etx.QueryParam("start")
	endParam := etx.QueryParam("end")
	startDate, endDate := parseDateRange(period, startParam, endParam)
	prevStart, prevEnd := previousPeriodRange(startDate, endDate)

	bucket := chooseBucket(startDate, endDate)

	stats, err := models.GetDashboardStats(ctx, d.db.Conn(), websiteID, startDate, endDate, prevStart, prevEnd, bucket)
	if err != nil {
		return render(etx, views.InternalError())
	}

	return render(etx, views.DashboardShow(website, stats, period, startParam, endParam, bucket))
}

func (d Dashboard) Live(etx *echo.Context) error {
	websiteID, err := uuid.Parse(etx.Param("id"))
	if err != nil {
		return etx.NoContent(http.StatusBadRequest)
	}

	app := cookies.GetApp(etx)
	ctx := etx.Request().Context()

	website, err := models.FindWebsite(ctx, d.db.Conn(), websiteID)
	if err != nil {
		return etx.NoContent(http.StatusNotFound)
	}

	if website.UserID != app.UserID {
		return etx.NoContent(http.StatusNotFound)
	}

	period := etx.QueryParam("period")
	startParam := etx.QueryParam("start")
	endParam := etx.QueryParam("end")
	startDate, endDate := parseDateRange(period, startParam, endParam)
	prevStart, prevEnd := previousPeriodRange(startDate, endDate)
	bucket := chooseBucket(startDate, endDate)

	stats, err := models.GetDashboardStats(ctx, d.db.Conn(), websiteID, startDate, endDate, prevStart, prevEnd, bucket)
	if err != nil {
		return etx.NoContent(http.StatusInternalServerError)
	}

	return hypermedia.MarshalAndPatchSignals(etx, map[string]any{
		"dashboard": dashboardSignalsPayload(stats, bucket),
	})
}

func parseDateRange(period, startParam, endParam string) (time.Time, time.Time) {
	now := time.Now().UTC()
	endDate := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, time.UTC)

	if period == "custom" && startParam != "" && endParam != "" {
		start, err1 := time.Parse("2006-01-02", startParam)
		end, err2 := time.Parse("2006-01-02", endParam)
		if err1 == nil && err2 == nil {
			startDate := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC)
			endDate := time.Date(end.Year(), end.Month(), end.Day(), 23, 59, 59, 0, time.UTC)
			return startDate, endDate
		}
	}

	switch period {
	case "today":
		return now.Add(-24 * time.Hour), now
	case "30d":
		return endDate.AddDate(0, 0, -30), endDate
	case "month":
		startDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		return startDate, endDate
	default: // "7d" or empty
		return endDate.AddDate(0, 0, -7), endDate
	}
}

func chooseBucket(start, end time.Time) string {
	if end.Sub(start) <= 48*time.Hour {
		return "hour"
	}
	return "day"
}

func previousPeriodRange(start, end time.Time) (time.Time, time.Time) {
	duration := end.Sub(start)
	prevEnd := start.Add(-time.Second)
	prevStart := prevEnd.Add(-duration)
	return prevStart, prevEnd
}

func formatCompact(n int64) string {
	abs := n
	if abs < 0 {
		abs = -abs
	}
	switch {
	case abs >= 1_000_000:
		v := float64(n) / 1_000_000
		if math.Abs(v) >= 10 {
			return fmt.Sprintf("%.0fM", v)
		}
		return fmt.Sprintf("%.1fM", v)
	case abs >= 1_000:
		v := float64(n) / 1_000
		if math.Abs(v) >= 100 {
			return fmt.Sprintf("%.0fk", v)
		}
		if math.Abs(v) >= 10 {
			return fmt.Sprintf("%.1fk", v)
		}
		return fmt.Sprintf("%.1fk", v)
	default:
		return fmt.Sprintf("%d", n)
	}
}

func formatRate(v float64) string {
	if v == 0 {
		return "0%"
	}
	return fmt.Sprintf("%.1f%%", v)
}

func formatFloat1(v float64) string {
	if v == 0 {
		return "0"
	}
	return fmt.Sprintf("%.1f", v)
}

func dashboardSignalsPayload(stats models.DashboardStats, bucket string) map[string]any {
	return map[string]any{
		"totals": map[string]any{
			"visitors":          formatCompact(stats.TotalUniqueVisitors),
			"visitorsChange":    math.Round(stats.UniqueVisitorsChange),
			"pageviews":         formatCompact(stats.TotalPageviews),
			"pageviewsChange":   math.Round(stats.PageviewsChange),
			"viewsPerVisitor":   formatFloat1(stats.ViewsPerVisitor),
			"vpvChange":         math.Round(stats.ViewsPerVisitorChange),
			"bounceRate":        formatRate(stats.BounceRate),
			"bounceRateChange":  math.Round(stats.BounceRateChange),
		},
		"series": map[string]any{
			"pageviews": toSeriesPayload(stats.PageviewsOverTime, bucket),
			"visitors":  toSeriesPayload(stats.VisitorsOverTime, bucket),
			"events":    toSeriesPayload(stats.EventsOverTime, bucket),
		},
		"lastUpdated": time.Now().UTC().Format("15:04:05 UTC"),
	}
}

func toSeriesPayload(buckets []models.TimeBucket, bucketType string) map[string]any {
	labels := make([]string, len(buckets))
	values := make([]int64, len(buckets))

	for i, b := range buckets {
		if bucketType == "hour" {
			labels[i] = b.Time.Format("Jan 02 15:00")
		} else {
			labels[i] = b.Time.Format("Jan 02")
		}
		values[i] = b.Count
	}

	return map[string]any{
		"labels": labels,
		"values": values,
	}
}

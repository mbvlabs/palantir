package controllers

import (
	"time"

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

	bucket := chooseBucket(startDate, endDate)

	stats, err := models.GetDashboardStats(ctx, d.db.Conn(), websiteID, startDate, endDate, bucket)
	if err != nil {
		return render(etx, views.InternalError())
	}

	return render(etx, views.DashboardShow(website, stats, period, bucket))
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
		startDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		return startDate, endDate
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

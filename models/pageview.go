package models

import (
	"context"
	"database/sql"
	"time"

	"palantir/internal/storage"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type PageviewEntity struct {
	bun.BaseModel `bun:"table:pageviews,alias:pageview"`
	ID            uuid.UUID      `bun:"id,pk,type:uuid"`
	CreatedAt     time.Time      `bun:"created_at"`
	WebsiteID     uuid.UUID      `bun:"website_id,type:uuid"`
	Url           string         `bun:"url"`
	Referrer      sql.NullString `bun:"referrer"`
	Browser       sql.NullString `bun:"browser"`
	Os            sql.NullString `bun:"os"`
	Device        sql.NullString `bun:"device"`
	Country       sql.NullString `bun:"country"`
	Language      sql.NullString `bun:"language"`
	ScreenWidth   sql.NullInt32  `bun:"screen_width"`
	VisitorHash   sql.NullString `bun:"visitor_hash"`
	CountryCode   sql.NullString `bun:"country_code"`
	CountryName   sql.NullString `bun:"country_name"`
	City          sql.NullString `bun:"city"`
	Region        sql.NullString `bun:"region"`
}

type CreatePageviewData struct {
	WebsiteID                                             uuid.UUID
	URL, Referrer, Browser, OS, Device, Country, Language string
	ScreenWidth                                           int32
	VisitorHash, CountryCode, CountryName, City, Region   string
}

func (pageview) Create(ctx context.Context, db storage.Executor, data CreatePageviewData) (PageviewEntity, error) {
	entity := PageviewEntity{
		ID: uuid.New(), CreatedAt: time.Now().UTC(), WebsiteID: data.WebsiteID, Url: data.URL,
		Referrer: nullString(data.Referrer), Browser: nullString(data.Browser), Os: nullString(data.OS),
		Device: nullString(data.Device), Country: nullString(data.Country), Language: nullString(data.Language),
		ScreenWidth: sql.NullInt32{Int32: data.ScreenWidth, Valid: data.ScreenWidth > 0},
		VisitorHash: nullString(data.VisitorHash), CountryCode: nullString(data.CountryCode),
		CountryName: nullString(data.CountryName), City: nullString(data.City), Region: nullString(data.Region),
	}
	if _, err := db.NewInsert().Model(&entity).Exec(ctx); err != nil {
		return PageviewEntity{}, err
	}
	return entity, nil
}

func nullString(value string) sql.NullString {
	return sql.NullString{String: value, Valid: value != ""}
}

type TimeBucket struct {
	Time  time.Time `json:"time"`
	Count int64     `json:"count"`
}

type RateBucket struct {
	Time time.Time `json:"time" bun:"bucket_time"`
	Rate float64   `json:"rate" bun:"rate"`
}

type BreakdownItem struct {
	Name  string `json:"name" bun:"name"`
	Views int64  `json:"views" bun:"views"`
}

type GeoBreakdownItem struct {
	Name  string `json:"name" bun:"name"`
	Code  string `json:"code" bun:"code"`
	Views int64  `json:"views" bun:"views"`
}

type DashboardStats struct {
	TotalPageviews, TotalUniqueVisitors, BounceCount                               int64
	ViewsPerVisitor, BounceRate                                                    float64
	PageviewsChange, UniqueVisitorsChange, ViewsPerVisitorChange, BounceRateChange float64
	PageviewsOverTime, VisitorsOverTime, EventsOverTime                            []TimeBucket
	BounceRateOverTime                                                             []RateBucket
	TopPages, TopReferrers, Browsers, OSes, Devices, TopEvents                     []BreakdownItem
	TopCountries, TopCities                                                        []GeoBreakdownItem
}

func DashboardStatsFor(ctx context.Context, db storage.Executor, websiteID uuid.UUID, start, end, previousStart, previousEnd time.Time, bucket string) (DashboardStats, error) {
	if bucket != "hour" {
		bucket = "day"
	}
	total, unique, bounce, err := pageviewTotals(ctx, db, websiteID, start, end)
	if err != nil {
		return DashboardStats{}, err
	}
	previousTotal, previousUnique, previousBounce, err := pageviewTotals(ctx, db, websiteID, previousStart, previousEnd)
	if err != nil {
		return DashboardStats{}, err
	}
	pageviews, err := pageviewTimeBuckets(ctx, db, websiteID, start, end, bucket, false)
	if err != nil {
		return DashboardStats{}, err
	}
	visitors, err := pageviewTimeBuckets(ctx, db, websiteID, start, end, bucket, true)
	if err != nil {
		return DashboardStats{}, err
	}
	bounceRates, err := bounceRateTimeBuckets(ctx, db, websiteID, start, end, bucket)
	if err != nil {
		return DashboardStats{}, err
	}
	topPages, err := pageviewBreakdown(ctx, db, websiteID, start, end, "url", false, 10)
	if err != nil {
		return DashboardStats{}, err
	}
	topReferrers, err := pageviewBreakdown(ctx, db, websiteID, start, end, "referrer", true, 10)
	if err != nil {
		return DashboardStats{}, err
	}
	browsers, err := pageviewBreakdown(ctx, db, websiteID, start, end, "browser", false, 0)
	if err != nil {
		return DashboardStats{}, err
	}
	oses, err := pageviewBreakdown(ctx, db, websiteID, start, end, "os", false, 0)
	if err != nil {
		return DashboardStats{}, err
	}
	devices, err := pageviewBreakdown(ctx, db, websiteID, start, end, "device", false, 0)
	if err != nil {
		return DashboardStats{}, err
	}
	countries, err := geoBreakdown(ctx, db, websiteID, start, end, true)
	if err != nil {
		return DashboardStats{}, err
	}
	cities, err := geoBreakdown(ctx, db, websiteID, start, end, false)
	if err != nil {
		return DashboardStats{}, err
	}
	topEvents, err := Event.top(ctx, db, websiteID, start, end)
	if err != nil {
		return DashboardStats{}, err
	}
	events, err := Event.overTime(ctx, db, websiteID, start, end, bucket)
	if err != nil {
		return DashboardStats{}, err
	}

	viewsPerVisitor := ratio(total, unique)
	previousViewsPerVisitor := ratio(previousTotal, previousUnique)
	bounceRate := ratio(bounce*100, unique)
	previousBounceRate := ratio(previousBounce*100, previousUnique)
	return DashboardStats{
		TotalPageviews: total, TotalUniqueVisitors: unique, BounceCount: bounce,
		ViewsPerVisitor: viewsPerVisitor, BounceRate: bounceRate,
		PageviewsChange:       percentChange(float64(previousTotal), float64(total)),
		UniqueVisitorsChange:  percentChange(float64(previousUnique), float64(unique)),
		ViewsPerVisitorChange: percentChange(previousViewsPerVisitor, viewsPerVisitor),
		BounceRateChange:      -percentChange(previousBounceRate, bounceRate),
		PageviewsOverTime:     pageviews, VisitorsOverTime: visitors, EventsOverTime: events,
		BounceRateOverTime: bounceRates,
		TopPages:           topPages, TopReferrers: topReferrers, Browsers: browsers, OSes: oses,
		Devices: devices, TopCountries: countries, TopCities: cities, TopEvents: topEvents,
	}, nil
}

func pageviewTotals(ctx context.Context, db storage.Executor, websiteID uuid.UUID, start, end time.Time) (int64, int64, int64, error) {
	base := func() *bun.SelectQuery {
		return db.NewSelect().TableExpr("pageviews AS pageview").
			Where("pageview.website_id = ?", websiteID).
			Where("pageview.created_at BETWEEN ? AND ?", start, end)
	}
	var total, unique, bounce int64
	if err := base().ColumnExpr("count(*)").Scan(ctx, &total); err != nil {
		return 0, 0, 0, err
	}
	if err := base().ColumnExpr("count(DISTINCT pageview.visitor_hash)").Where("pageview.visitor_hash IS NOT NULL").Scan(ctx, &unique); err != nil {
		return 0, 0, 0, err
	}
	subquery := base().ColumnExpr("pageview.visitor_hash").Where("pageview.visitor_hash IS NOT NULL").GroupExpr("pageview.visitor_hash").Having("count(*) = 1")
	if err := db.NewSelect().TableExpr("(?) AS bounce_visitors", subquery).ColumnExpr("count(*)").Scan(ctx, &bounce); err != nil {
		return 0, 0, 0, err
	}
	return total, unique, bounce, nil
}

func pageviewTimeBuckets(ctx context.Context, db storage.Executor, websiteID uuid.UUID, start, end time.Time, bucket string, unique bool) ([]TimeBucket, error) {
	count := "count(*)"
	if unique {
		count = "count(DISTINCT pageview.visitor_hash)"
	}
	var rows []struct {
		Time  time.Time `bun:"bucket_time"`
		Count int64     `bun:"count"`
	}
	err := db.NewSelect().TableExpr("pageviews AS pageview").
		ColumnExpr("date_trunc(?, pageview.created_at) AS bucket_time", bucket).
		ColumnExpr(count+" AS count").
		Where("pageview.website_id = ?", websiteID).
		Where("pageview.created_at BETWEEN ? AND ?", start, end).
		GroupExpr("bucket_time").OrderExpr("bucket_time").Scan(ctx, &rows)
	if err != nil {
		return nil, err
	}
	items := make([]TimeBucket, len(rows))
	for i, row := range rows {
		items[i] = TimeBucket{Time: row.Time, Count: row.Count}
	}
	return fillTimeBuckets(items, start, end, bucket), nil
}

func bounceRateTimeBuckets(ctx context.Context, db storage.Executor, websiteID uuid.UUID, start, end time.Time, bucket string) ([]RateBucket, error) {
	visitors := db.NewSelect().TableExpr("pageviews AS pageview").
		ColumnExpr("date_trunc(?, pageview.created_at) AS bucket_time", bucket).
		ColumnExpr("pageview.visitor_hash").
		ColumnExpr("count(*) AS views").
		Where("pageview.website_id = ?", websiteID).
		Where("pageview.created_at BETWEEN ? AND ?", start, end).
		Where("pageview.visitor_hash IS NOT NULL").
		GroupExpr("bucket_time, pageview.visitor_hash")

	var sparse []RateBucket
	err := db.NewSelect().TableExpr("(?) AS bucket_visitors", visitors).
		ColumnExpr("bucket_time").
		ColumnExpr("count(*) FILTER (WHERE views = 1) * 100.0 / count(*) AS rate").
		GroupExpr("bucket_time").
		OrderExpr("bucket_time").
		Scan(ctx, &sparse)
	if err != nil {
		return nil, err
	}

	existing := make(map[int64]float64, len(sparse))
	for _, item := range sparse {
		existing[item.Time.Unix()] = item.Rate
	}
	buckets := fillTimeBuckets(nil, start, end, bucket)
	items := make([]RateBucket, len(buckets))
	for i, item := range buckets {
		items[i] = RateBucket{Time: item.Time, Rate: existing[item.Time.Unix()]}
	}
	return items, nil
}

func pageviewBreakdown(ctx context.Context, db storage.Executor, websiteID uuid.UUID, start, end time.Time, column string, nonEmpty bool, limit int) ([]BreakdownItem, error) {
	query := db.NewSelect().TableExpr("pageviews AS pageview").
		ColumnExpr("COALESCE(?, 'Unknown') AS name, count(*) AS views", bun.Ident("pageview."+column)).
		Where("pageview.website_id = ?", websiteID).
		Where("pageview.created_at BETWEEN ? AND ?", start, end).
		GroupExpr("?", bun.Ident("pageview."+column)).OrderExpr("views DESC")
	if nonEmpty {
		query = query.Where("? IS NOT NULL AND ? != ''", bun.Ident("pageview."+column), bun.Ident("pageview."+column))
	}
	if limit > 0 {
		query = query.Limit(limit)
	}
	items := make([]BreakdownItem, 0)
	if err := query.Scan(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func geoBreakdown(ctx context.Context, db storage.Executor, websiteID uuid.UUID, start, end time.Time, countries bool) ([]GeoBreakdownItem, error) {
	name := "city"
	if countries {
		name = "country_name"
	}
	items := make([]GeoBreakdownItem, 0)
	err := db.NewSelect().TableExpr("pageviews AS pageview").
		ColumnExpr("? AS name, pageview.country_code AS code, count(*) AS views", bun.Ident("pageview."+name)).
		Where("pageview.website_id = ?", websiteID).
		Where("pageview.created_at BETWEEN ? AND ?", start, end).
		Where("? IS NOT NULL AND ? != ''", bun.Ident("pageview."+name), bun.Ident("pageview."+name)).
		GroupExpr("?, pageview.country_code", bun.Ident("pageview."+name)).
		OrderExpr("views DESC").Limit(10).Scan(ctx, &items)
	return items, err
}

func fillTimeBuckets(sparse []TimeBucket, start, end time.Time, bucket string) []TimeBucket {
	existing := make(map[int64]int64, len(sparse))
	for _, item := range sparse {
		existing[item.Time.Unix()] = item.Count
	}
	cursor := truncateBucket(start, bucket)
	items := make([]TimeBucket, 0)
	for !cursor.After(end) {
		items = append(items, TimeBucket{Time: cursor, Count: existing[cursor.Unix()]})
		if bucket == "hour" {
			cursor = cursor.Add(time.Hour)
		} else {
			cursor = cursor.AddDate(0, 0, 1)
		}
	}
	return items
}

func truncateBucket(value time.Time, bucket string) time.Time {
	if bucket == "hour" {
		return value.Truncate(time.Hour)
	}
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, value.Location())
}

func ratio(numerator, denominator int64) float64 {
	if denominator == 0 {
		return 0
	}
	return float64(numerator) / float64(denominator)
}

func percentChange(previous, current float64) float64 {
	if previous == 0 {
		if current == 0 {
			return 0
		}
		return 100
	}
	return (current - previous) / previous * 100
}

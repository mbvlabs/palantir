package models

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"palantir/internal/storage"
	"palantir/models/internal/db"
)

type Pageview struct {
	ID          uuid.UUID
	CreatedAt   time.Time
	WebsiteID   uuid.UUID
	URL         string
	Referrer    string
	Browser     string
	OS          string
	Device      string
	Country     string
	Language    string
	ScreenWidth int32
	VisitorHash string
	CountryCode string
	CountryName string
	City        string
	Region      string
}

type CreatePageviewData struct {
	WebsiteID   uuid.UUID
	URL         string
	Referrer    string
	Browser     string
	OS          string
	Device      string
	Country     string
	Language    string
	ScreenWidth int32
	VisitorHash string
	CountryCode string
	CountryName string
	City        string
	Region      string
}

func CreatePageview(
	ctx context.Context,
	exec storage.Executor,
	data CreatePageviewData,
) (Pageview, error) {
	params := db.InsertPageviewParams{
		ID:          uuid.New(),
		WebsiteID:   data.WebsiteID,
		Url:         data.URL,
		Referrer:    pgtype.Text{String: data.Referrer, Valid: data.Referrer != ""},
		Browser:     pgtype.Text{String: data.Browser, Valid: data.Browser != ""},
		Os:          pgtype.Text{String: data.OS, Valid: data.OS != ""},
		Device:      pgtype.Text{String: data.Device, Valid: data.Device != ""},
		Country:     pgtype.Text{String: data.Country, Valid: data.Country != ""},
		Language:    pgtype.Text{String: data.Language, Valid: data.Language != ""},
		ScreenWidth: pgtype.Int4{Int32: data.ScreenWidth, Valid: data.ScreenWidth > 0},
		VisitorHash: pgtype.Text{String: data.VisitorHash, Valid: data.VisitorHash != ""},
		CountryCode: pgtype.Text{String: data.CountryCode, Valid: data.CountryCode != ""},
		CountryName: pgtype.Text{String: data.CountryName, Valid: data.CountryName != ""},
		City:        pgtype.Text{String: data.City, Valid: data.City != ""},
		Region:      pgtype.Text{String: data.Region, Valid: data.Region != ""},
	}
	row, err := queries.InsertPageview(ctx, exec, params)
	if err != nil {
		return Pageview{}, err
	}

	return rowToPageview(row), nil
}

type PageviewsPerDay struct {
	Date  time.Time
	Views int64
}

type TimeBucket struct {
	Time  time.Time
	Count int64
}

type BreakdownItem struct {
	Name  string
	Views int64
}

type GeoBreakdownItem struct {
	Name  string
	Code  string
	Views int64
}

type DashboardStats struct {
	TotalPageviews      int64
	TotalUniqueVisitors int64
	BounceCount         int64
	ViewsPerVisitor     float64
	BounceRate          float64

	// Percentage changes vs previous period
	PageviewsChange      float64
	UniqueVisitorsChange float64
	ViewsPerVisitorChange float64
	BounceRateChange     float64

	PageviewsOverTime []TimeBucket
	VisitorsOverTime  []TimeBucket
	TopPages          []BreakdownItem
	TopReferrers      []BreakdownItem
	Browsers          []BreakdownItem
	OSes              []BreakdownItem
	Devices           []BreakdownItem
	TopCountries      []GeoBreakdownItem
	TopCities         []GeoBreakdownItem
	TopEvents         []BreakdownItem
	EventsOverTime    []TimeBucket
}

func GetDashboardStats(
	ctx context.Context,
	exec storage.Executor,
	websiteID uuid.UUID,
	startDate time.Time,
	endDate time.Time,
	prevStartDate time.Time,
	prevEndDate time.Time,
	bucket string,
) (DashboardStats, error) {
	dateParams := func() (pgtype.Timestamptz, pgtype.Timestamptz) {
		return pgtype.Timestamptz{Time: startDate, Valid: true},
			pgtype.Timestamptz{Time: endDate, Valid: true}
	}
	start, end := dateParams()

	prevStart := pgtype.Timestamptz{Time: prevStartDate, Valid: true}
	prevEnd := pgtype.Timestamptz{Time: prevEndDate, Valid: true}

	total, err := queries.QueryTotalPageviews(ctx, exec, db.QueryTotalPageviewsParams{
		WebsiteID: websiteID,
		StartDate: start,
		EndDate:   end,
	})
	if err != nil {
		return DashboardStats{}, err
	}

	totalUnique, err := queries.QueryTotalUniqueVisitors(ctx, exec, db.QueryTotalUniqueVisitorsParams{
		WebsiteID: websiteID,
		StartDate: start,
		EndDate:   end,
	})
	if err != nil {
		return DashboardStats{}, err
	}

	bounceCount, err := queries.QueryBounceCount(ctx, exec, db.QueryBounceCountParams{
		WebsiteID: websiteID,
		StartDate: start,
		EndDate:   end,
	})
	if err != nil {
		return DashboardStats{}, err
	}

	// Previous period totals
	prevTotal, err := queries.QueryTotalPageviews(ctx, exec, db.QueryTotalPageviewsParams{
		WebsiteID: websiteID,
		StartDate: prevStart,
		EndDate:   prevEnd,
	})
	if err != nil {
		return DashboardStats{}, err
	}

	prevUnique, err := queries.QueryTotalUniqueVisitors(ctx, exec, db.QueryTotalUniqueVisitorsParams{
		WebsiteID: websiteID,
		StartDate: prevStart,
		EndDate:   prevEnd,
	})
	if err != nil {
		return DashboardStats{}, err
	}

	prevBounce, err := queries.QueryBounceCount(ctx, exec, db.QueryBounceCountParams{
		WebsiteID: websiteID,
		StartDate: prevStart,
		EndDate:   prevEnd,
	})
	if err != nil {
		return DashboardStats{}, err
	}

	// Compute derived metrics
	viewsPerVisitor := computeRatio(total, totalUnique)
	prevViewsPerVisitor := computeRatio(prevTotal, prevUnique)
	bounceRate := computeRatio(bounceCount*100, totalUnique)
	prevBounceRate := computeRatio(prevBounce*100, prevUnique)

	pvBucketRows, err := queries.QueryPageviewsTimeBucketed(ctx, exec, db.QueryPageviewsTimeBucketedParams{
		WebsiteID: websiteID,
		Bucket:    bucket,
		StartDate: start,
		EndDate:   end,
	})
	if err != nil {
		return DashboardStats{}, err
	}

	pvSparse := make([]TimeBucket, len(pvBucketRows))
	for i, row := range pvBucketRows {
		pvSparse[i] = TimeBucket{Time: row.BucketTime.Time, Count: row.Views}
	}
	pvOverTime := fillTimeBuckets(pvSparse, startDate, endDate, bucket)

	uvBucketRows, err := queries.QueryUniqueVisitorsTimeBucketed(ctx, exec, db.QueryUniqueVisitorsTimeBucketedParams{
		WebsiteID: websiteID,
		Bucket:    bucket,
		StartDate: start,
		EndDate:   end,
	})
	if err != nil {
		return DashboardStats{}, err
	}

	uvSparse := make([]TimeBucket, len(uvBucketRows))
	for i, row := range uvBucketRows {
		uvSparse[i] = TimeBucket{Time: row.BucketTime.Time, Count: row.Visitors}
	}
	uvOverTime := fillTimeBuckets(uvSparse, startDate, endDate, bucket)

	topPagesRows, err := queries.QueryTopPages(ctx, exec, db.QueryTopPagesParams{
		WebsiteID: websiteID,
		StartDate: start,
		EndDate:   end,
	})
	if err != nil {
		return DashboardStats{}, err
	}

	topPages := make([]BreakdownItem, len(topPagesRows))
	for i, row := range topPagesRows {
		topPages[i] = BreakdownItem{Name: row.Url, Views: row.Views}
	}

	topRefRows, err := queries.QueryTopReferrers(ctx, exec, db.QueryTopReferrersParams{
		WebsiteID: websiteID,
		StartDate: start,
		EndDate:   end,
	})
	if err != nil {
		return DashboardStats{}, err
	}

	topReferrers := make([]BreakdownItem, len(topRefRows))
	for i, row := range topRefRows {
		topReferrers[i] = BreakdownItem{Name: row.Referrer.String, Views: row.Views}
	}

	browserRows, err := queries.QueryBrowserBreakdown(ctx, exec, db.QueryBrowserBreakdownParams{
		WebsiteID: websiteID,
		StartDate: start,
		EndDate:   end,
	})
	if err != nil {
		return DashboardStats{}, err
	}

	browsers := make([]BreakdownItem, len(browserRows))
	for i, row := range browserRows {
		browsers[i] = BreakdownItem{Name: row.Browser.String, Views: row.Views}
	}

	osRows, err := queries.QueryOSBreakdown(ctx, exec, db.QueryOSBreakdownParams{
		WebsiteID: websiteID,
		StartDate: start,
		EndDate:   end,
	})
	if err != nil {
		return DashboardStats{}, err
	}

	oses := make([]BreakdownItem, len(osRows))
	for i, row := range osRows {
		oses[i] = BreakdownItem{Name: row.Os.String, Views: row.Views}
	}

	deviceRows, err := queries.QueryDeviceBreakdown(ctx, exec, db.QueryDeviceBreakdownParams{
		WebsiteID: websiteID,
		StartDate: start,
		EndDate:   end,
	})
	if err != nil {
		return DashboardStats{}, err
	}

	devices := make([]BreakdownItem, len(deviceRows))
	for i, row := range deviceRows {
		devices[i] = BreakdownItem{Name: row.Device.String, Views: row.Views}
	}

	countryRows, err := queries.QueryTopCountries(ctx, exec, db.QueryTopCountriesParams{
		WebsiteID: websiteID,
		StartDate: start,
		EndDate:   end,
	})
	if err != nil {
		return DashboardStats{}, err
	}

	countries := make([]GeoBreakdownItem, len(countryRows))
	for i, row := range countryRows {
		countries[i] = GeoBreakdownItem{Name: row.CountryName.String, Code: row.CountryCode.String, Views: row.Views}
	}

	cityRows, err := queries.QueryTopCities(ctx, exec, db.QueryTopCitiesParams{
		WebsiteID: websiteID,
		StartDate: start,
		EndDate:   end,
	})
	if err != nil {
		return DashboardStats{}, err
	}

	cities := make([]GeoBreakdownItem, len(cityRows))
	for i, row := range cityRows {
		cities[i] = GeoBreakdownItem{Name: row.City.String, Code: row.CountryCode.String, Views: row.Views}
	}

	topEventRows, err := queries.QueryTopEvents(ctx, exec, db.QueryTopEventsParams{
		WebsiteID: websiteID,
		StartDate: start,
		EndDate:   end,
	})
	if err != nil {
		return DashboardStats{}, err
	}

	topEvents := make([]BreakdownItem, len(topEventRows))
	for i, row := range topEventRows {
		topEvents[i] = BreakdownItem{Name: row.EventName, Views: row.EventCount}
	}

	eventBucketRows, err := queries.QueryEventsTimeBucketed(ctx, exec, db.QueryEventsTimeBucketedParams{
		WebsiteID: websiteID,
		Bucket:    bucket,
		StartDate: start,
		EndDate:   end,
	})
	if err != nil {
		return DashboardStats{}, err
	}

	eventsSparse := make([]TimeBucket, len(eventBucketRows))
	for i, row := range eventBucketRows {
		eventsSparse[i] = TimeBucket{Time: row.BucketTime.Time, Count: row.EventCount}
	}
	eventsOverTime := fillTimeBuckets(eventsSparse, startDate, endDate, bucket)

	return DashboardStats{
		TotalPageviews:        total,
		TotalUniqueVisitors:   totalUnique,
		BounceCount:           bounceCount,
		ViewsPerVisitor:       viewsPerVisitor,
		BounceRate:            bounceRate,
		PageviewsChange:       percentChange(prevTotal, total),
		UniqueVisitorsChange:  percentChange(prevUnique, totalUnique),
		ViewsPerVisitorChange: percentChangeFloat(prevViewsPerVisitor, viewsPerVisitor),
		BounceRateChange:      -percentChangeFloat(prevBounceRate, bounceRate), // negate: decrease is good
		PageviewsOverTime:     pvOverTime,
		VisitorsOverTime:      uvOverTime,
		TopPages:              topPages,
		TopReferrers:          topReferrers,
		Browsers:              browsers,
		OSes:                  oses,
		Devices:               devices,
		TopCountries:          countries,
		TopCities:             cities,
		TopEvents:             topEvents,
		EventsOverTime:        eventsOverTime,
	}, nil
}

// fillTimeBuckets generates a complete time series from startDate to endDate
// with the given bucket granularity, filling in zeros for missing buckets.
func fillTimeBuckets(sparse []TimeBucket, startDate, endDate time.Time, bucket string) []TimeBucket {
	// Build a lookup of existing data keyed by truncated time
	existing := make(map[int64]int64, len(sparse))
	for _, tb := range sparse {
		existing[tb.Time.Unix()] = tb.Count
	}

	// Truncate start to bucket boundary
	start := truncateToBucket(startDate, bucket)
	end := endDate

	var step time.Duration
	if bucket == "hour" {
		step = time.Hour
	} else {
		step = 24 * time.Hour
	}

	var result []TimeBucket
	for t := start; !t.After(end); t = t.Add(step) {
		count := existing[t.Unix()]
		result = append(result, TimeBucket{Time: t, Count: count})
	}

	return result
}

func truncateToBucket(t time.Time, bucket string) time.Time {
	if bucket == "hour" {
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location())
	}
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func percentChange(prev, current int64) float64 {
	if prev == 0 {
		if current == 0 {
			return 0
		}
		return 100
	}
	return (float64(current) - float64(prev)) / float64(prev) * 100
}

func percentChangeFloat(prev, current float64) float64 {
	if prev == 0 {
		if current == 0 {
			return 0
		}
		return 100
	}
	return (current - prev) / prev * 100
}

func computeRatio(numerator, denominator int64) float64 {
	if denominator == 0 {
		return 0
	}
	return float64(numerator) / float64(denominator)
}

func rowToPageview(row db.Pageview) Pageview {
	return Pageview{
		ID:          row.ID,
		CreatedAt:   row.CreatedAt.Time,
		WebsiteID:   row.WebsiteID,
		URL:         row.Url,
		Referrer:    row.Referrer.String,
		Browser:     row.Browser.String,
		OS:          row.Os.String,
		Device:      row.Device.String,
		Country:     row.Country.String,
		Language:    row.Language.String,
		ScreenWidth: row.ScreenWidth.Int32,
		VisitorHash: row.VisitorHash.String,
		CountryCode: row.CountryCode.String,
		CountryName: row.CountryName.String,
		City:        row.City.String,
		Region:      row.Region.String,
	}
}

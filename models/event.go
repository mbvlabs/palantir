package models

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"palantir/internal/storage"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type EventEntity struct {
	bun.BaseModel `bun:"table:events,alias:event"`
	ID            uuid.UUID       `bun:"id,pk,type:uuid"`
	CreatedAt     time.Time       `bun:"created_at"`
	WebsiteID     uuid.UUID       `bun:"website_id,type:uuid"`
	Url           string          `bun:"url"`
	EventName     string          `bun:"event_name"`
	EventData     json.RawMessage `bun:"event_data,type:jsonb"`
	VisitorHash   sql.NullString  `bun:"visitor_hash"`
	CountryCode   sql.NullString  `bun:"country_code"`
	CountryName   sql.NullString  `bun:"country_name"`
	City          sql.NullString  `bun:"city"`
	Region        sql.NullString  `bun:"region"`
}

type CreateEventData struct {
	WebsiteID                                           uuid.UUID
	URL, EventName                                      string
	EventData                                           json.RawMessage
	VisitorHash, CountryCode, CountryName, City, Region string
}

func (event) Create(ctx context.Context, db storage.Executor, data CreateEventData) (EventEntity, error) {
	entity := EventEntity{
		ID: uuid.New(), CreatedAt: time.Now().UTC(), WebsiteID: data.WebsiteID,
		Url: data.URL, EventName: data.EventName, EventData: data.EventData,
		VisitorHash: nullString(data.VisitorHash), CountryCode: nullString(data.CountryCode),
		CountryName: nullString(data.CountryName), City: nullString(data.City), Region: nullString(data.Region),
	}
	if len(entity.EventData) == 0 {
		entity.EventData = json.RawMessage(`{}`)
	}
	if _, err := db.NewInsert().Model(&entity).Exec(ctx); err != nil {
		return EventEntity{}, err
	}
	return entity, nil
}

func (event) top(ctx context.Context, db storage.Executor, websiteID uuid.UUID, start, end time.Time) ([]BreakdownItem, error) {
	items := make([]BreakdownItem, 0)
	err := db.NewSelect().TableExpr("events AS event").
		ColumnExpr("event.event_name AS name, count(*) AS views").
		Where("event.website_id = ?", websiteID).
		Where("event.created_at BETWEEN ? AND ?", start, end).
		GroupExpr("event.event_name").OrderExpr("views DESC").Limit(10).Scan(ctx, &items)
	return items, err
}

func (event) overTime(ctx context.Context, db storage.Executor, websiteID uuid.UUID, start, end time.Time, bucket string) ([]TimeBucket, error) {
	var rows []struct {
		Time  time.Time `bun:"bucket_time"`
		Count int64     `bun:"count"`
	}
	err := db.NewSelect().TableExpr("events AS event").
		ColumnExpr("date_trunc(?, event.created_at) AS bucket_time, count(*) AS count", bucket).
		Where("event.website_id = ?", websiteID).
		Where("event.created_at BETWEEN ? AND ?", start, end).
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

package models

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"palantir/internal/storage"
	"palantir/models/internal/db"
)

type Event struct {
	ID          uuid.UUID
	CreatedAt   time.Time
	WebsiteID   uuid.UUID
	URL         string
	EventName   string
	EventData   json.RawMessage
	VisitorHash string
	CountryCode string
	CountryName string
	City        string
	Region      string
}

type CreateEventData struct {
	WebsiteID   uuid.UUID
	URL         string
	EventName   string
	EventData   json.RawMessage
	VisitorHash string
	CountryCode string
	CountryName string
	City        string
	Region      string
}

func CreateEvent(
	ctx context.Context,
	exec storage.Executor,
	data CreateEventData,
) (Event, error) {
	params := db.InsertEventParams{
		ID:          uuid.New(),
		WebsiteID:   data.WebsiteID,
		Url:         data.URL,
		EventName:   data.EventName,
		EventData:   data.EventData,
		VisitorHash: pgtype.Text{String: data.VisitorHash, Valid: data.VisitorHash != ""},
		CountryCode: pgtype.Text{String: data.CountryCode, Valid: data.CountryCode != ""},
		CountryName: pgtype.Text{String: data.CountryName, Valid: data.CountryName != ""},
		City:        pgtype.Text{String: data.City, Valid: data.City != ""},
		Region:      pgtype.Text{String: data.Region, Valid: data.Region != ""},
	}
	row, err := queries.InsertEvent(ctx, exec, params)
	if err != nil {
		return Event{}, err
	}

	return rowToEvent(row), nil
}

func rowToEvent(row db.Event) Event {
	return Event{
		ID:          row.ID,
		CreatedAt:   row.CreatedAt.Time,
		WebsiteID:   row.WebsiteID,
		URL:         row.Url,
		EventName:   row.EventName,
		EventData:   row.EventData,
		VisitorHash: row.VisitorHash.String,
		CountryCode: row.CountryCode.String,
		CountryName: row.CountryName.String,
		City:        row.City.String,
		Region:      row.Region.String,
	}
}

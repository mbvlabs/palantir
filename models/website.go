package models

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"palantir/internal/storage"
	"palantir/models/internal/db"
)

type Website struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	UserID    uuid.UUID
	Name      string
	Domain    string
}

type CreateWebsiteData struct {
	UserID uuid.UUID
	Name   string `validate:"required,max=255"`
	Domain string `validate:"required,max=255"`
}

func CreateWebsite(
	ctx context.Context,
	exec storage.Executor,
	data CreateWebsiteData,
) (Website, error) {
	if err := Validate.Struct(data); err != nil {
		return Website{}, errors.Join(ErrDomainValidation, err)
	}

	params := db.InsertWebsiteParams{
		ID:     uuid.New(),
		UserID: data.UserID,
		Name:   data.Name,
		Domain: data.Domain,
	}
	row, err := queries.InsertWebsite(ctx, exec, params)
	if err != nil {
		return Website{}, err
	}

	return rowToWebsite(row), nil
}

type UpdateWebsiteData struct {
	ID     uuid.UUID
	Name   string `validate:"required,max=255"`
	Domain string `validate:"required,max=255"`
}

func UpdateWebsite(
	ctx context.Context,
	exec storage.Executor,
	data UpdateWebsiteData,
) (Website, error) {
	if err := Validate.Struct(data); err != nil {
		return Website{}, errors.Join(ErrDomainValidation, err)
	}

	params := db.UpdateWebsiteParams{
		ID:     data.ID,
		Name:   data.Name,
		Domain: data.Domain,
	}
	row, err := queries.UpdateWebsite(ctx, exec, params)
	if err != nil {
		return Website{}, err
	}

	return rowToWebsite(row), nil
}

func FindWebsite(
	ctx context.Context,
	exec storage.Executor,
	id uuid.UUID,
) (Website, error) {
	row, err := queries.QueryWebsiteByID(ctx, exec, id)
	if err != nil {
		return Website{}, err
	}

	return rowToWebsite(row), nil
}

func FindWebsitesByUserID(
	ctx context.Context,
	exec storage.Executor,
	userID uuid.UUID,
) ([]Website, error) {
	rows, err := queries.QueryWebsitesByUserID(ctx, exec, userID)
	if err != nil {
		return nil, err
	}

	websites := make([]Website, len(rows))
	for i, row := range rows {
		websites[i] = rowToWebsite(row)
	}
	return websites, nil
}

func DestroyWebsite(
	ctx context.Context,
	exec storage.Executor,
	id uuid.UUID,
) error {
	return queries.DeleteWebsite(ctx, exec, id)
}

func rowToWebsite(row db.Website) Website {
	return Website{
		ID:        row.ID,
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
		UserID:    row.UserID,
		Name:      row.Name,
		Domain:    row.Domain,
	}
}

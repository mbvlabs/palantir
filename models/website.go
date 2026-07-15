package models

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"strings"
	"time"

	"palantir/internal/storage"
	"palantir/internal/validation"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type WebsiteEntity struct {
	bun.BaseModel `bun:"table:websites,alias:website"`
	ID            uuid.UUID `bun:"id,pk,type:uuid"`
	CreatedAt     time.Time `bun:"created_at"`
	UpdatedAt     time.Time `bun:"updated_at"`
	UserID        uuid.UUID `bun:"user_id,type:uuid"`
	Name          string    `bun:"name"`
	Domain        string    `bun:"domain"`
}

func (e *WebsiteEntity) Validate() error {
	b := validation.NewBuilder()
	b.Required("name", e.Name)
	b.MaxLen("name", e.Name, 255)
	b.Required("domain", e.Domain)
	b.MaxLen("domain", e.Domain, 255)
	parsed, err := url.Parse("https://" + strings.TrimSpace(e.Domain))
	if err != nil || parsed.Hostname() == "" || parsed.Host != strings.TrimSpace(e.Domain) || strings.ContainsAny(e.Domain, " /?#") {
		b.Add("domain", "domain", "Domain must be a valid hostname")
	}
	return b.Err()
}

type CreateWebsiteData struct {
	UserID uuid.UUID
	Name   string
	Domain string
}

func (website) Create(ctx context.Context, db storage.Executor, data CreateWebsiteData) (WebsiteEntity, error) {
	now := time.Now().UTC()
	entity := WebsiteEntity{
		ID: uuid.New(), CreatedAt: now, UpdatedAt: now, UserID: data.UserID,
		Name: strings.TrimSpace(data.Name), Domain: strings.ToLower(strings.TrimSpace(data.Domain)),
	}
	if err := validation.Validate(&entity); err != nil {
		return WebsiteEntity{}, errors.Join(ErrDomainValidation, err)
	}
	if _, err := db.NewInsert().Model(&entity).Exec(ctx); err != nil {
		return WebsiteEntity{}, err
	}
	return entity, nil
}

type UpdateWebsiteData struct {
	ID     uuid.UUID
	Name   string
	Domain string
}

func (website) Find(ctx context.Context, db storage.Executor, id uuid.UUID) (WebsiteEntity, error) {
	var entity WebsiteEntity
	if err := db.NewSelect().Model(&entity).Where("website.id = ?", id).Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return WebsiteEntity{}, ErrNotFound
		}
		return WebsiteEntity{}, err
	}
	return entity, nil
}

func (website) FindOwned(ctx context.Context, db storage.Executor, id, ownerID uuid.UUID) (WebsiteEntity, error) {
	var entity WebsiteEntity
	if err := db.NewSelect().Model(&entity).
		Where("website.id = ?", id).
		Where("website.user_id = ?", ownerID).
		Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return WebsiteEntity{}, ErrNotFound
		}
		return WebsiteEntity{}, err
	}
	return entity, nil
}

func (website) AllOwned(ctx context.Context, db storage.Executor, ownerID uuid.UUID) ([]WebsiteEntity, error) {
	entities := make([]WebsiteEntity, 0)
	if err := db.NewSelect().Model(&entities).
		Where("website.user_id = ?", ownerID).
		OrderExpr("website.created_at ASC").
		Scan(ctx); err != nil {
		return nil, err
	}
	return entities, nil
}

func (w website) UpdateOwned(ctx context.Context, db storage.Executor, ownerID uuid.UUID, data UpdateWebsiteData) (WebsiteEntity, error) {
	current, err := w.FindOwned(ctx, db, data.ID, ownerID)
	if err != nil {
		return WebsiteEntity{}, err
	}
	current.Name = strings.TrimSpace(data.Name)
	current.Domain = strings.ToLower(strings.TrimSpace(data.Domain))
	current.UpdatedAt = time.Now().UTC()
	if err := validation.Validate(&current); err != nil {
		return WebsiteEntity{}, errors.Join(ErrDomainValidation, err)
	}
	if err := db.NewUpdate().Model(&current).
		Column("name", "domain", "updated_at").
		WherePK().Returning("*").Scan(ctx); err != nil {
		return WebsiteEntity{}, err
	}
	return current, nil
}

func (website) DestroyOwned(ctx context.Context, db storage.Executor, id, ownerID uuid.UUID) error {
	result, err := db.NewDelete().Model((*WebsiteEntity)(nil)).
		Where("id = ?", id).Where("user_id = ?", ownerID).Exec(ctx)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

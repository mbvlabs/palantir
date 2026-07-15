package seeds

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"palantir/internal/storage"
	"palantir/models"
	"palantir/models/factories"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

var demoWebsiteIDs = []uuid.UUID{
	uuid.MustParse("11111111-1111-4111-8111-111111111111"),
	uuid.MustParse("22222222-2222-4222-8222-222222222222"),
}

func seedPalantir(ctx context.Context, exec storage.Executor) error {
	admin, err := ensureSeedUser(ctx, exec, "admin@example.com", true)
	if err != nil {
		return err
	}
	if _, err := ensureSeedUser(ctx, exec, "user@example.com", false); err != nil {
		return err
	}

	sites := []struct{ name, domain string }{{"Palantir Demo Store", "store.demo.test"}, {"Palantir Demo Journal", "journal.demo.test"}}
	for index, site := range sites {
		entity := factories.BuildWebsite(admin.ID, factories.WithWebsitesName(site.name), factories.WithWebsitesDomain(site.domain))
		entity.ID = demoWebsiteIDs[index]
		entity.CreatedAt = time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
		entity.UpdatedAt = entity.CreatedAt
		if _, err := exec.NewInsert().Model(&entity).On("CONFLICT (id) DO UPDATE").
			Set("user_id = EXCLUDED.user_id").Set("name = EXCLUDED.name").Set("domain = EXCLUDED.domain").Set("updated_at = EXCLUDED.updated_at").Exec(ctx); err != nil {
			return fmt.Errorf("seed demo website: %w", err)
		}
	}
	if _, err := exec.NewDelete().Model((*models.PageviewEntity)(nil)).Where("website_id IN (?)", bun.In(demoWebsiteIDs)).Exec(ctx); err != nil {
		return err
	}
	if _, err := exec.NewDelete().Model((*models.EventEntity)(nil)).Where("website_id IN (?)", bun.In(demoWebsiteIDs)).Exec(ctx); err != nil {
		return err
	}

	day := time.Now().UTC()
	day = time.Date(day.Year(), day.Month(), day.Day(), 12, 0, 0, 0, time.UTC).AddDate(0, 0, -29)
	browsers := []string{"Firefox", "Chrome", "Safari"}
	oses := []string{"Linux", "Windows", "macOS", "iOS"}
	devices := []string{"desktop", "mobile", "tablet"}
	countries := []struct{ code, name, city, region string }{{"NL", "Netherlands", "Amsterdam", "North Holland"}, {"DE", "Germany", "Berlin", "Berlin"}, {"US", "United States", "New York", "New York"}}
	urls := []string{"/", "/pricing", "/docs", "/blog"}
	referrers := []string{"", "https://www.google.com/", "https://news.ycombinator.com/"}
	eventNames := []string{"signup", "purchase", "newsletter"}

	var pageviews []models.PageviewEntity
	var events []models.EventEntity
	for siteIndex, websiteID := range demoWebsiteIDs {
		for dayIndex := 0; dayIndex < 30; dayIndex++ {
			if dayIndex%9 == 0 {
				continue
			}
			at := day.AddDate(0, 0, dayIndex)
			for visitIndex := 0; visitIndex < 3+(dayIndex+siteIndex)%5; visitIndex++ {
				geo := countries[(dayIndex+visitIndex+siteIndex)%len(countries)]
				entity := factories.BuildPageview(websiteID)
				entity.ID = deterministicID(fmt.Sprintf("pageview:%d:%d:%d", siteIndex, dayIndex, visitIndex))
				entity.CreatedAt = at.Add(time.Duration(visitIndex) * time.Hour)
				entity.Url = urls[(dayIndex+visitIndex)%len(urls)]
				entity.Referrer = nullable(referrers[visitIndex%len(referrers)])
				entity.Browser = nullable(browsers[visitIndex%len(browsers)])
				entity.Os = nullable(oses[(dayIndex+visitIndex)%len(oses)])
				entity.Device = nullable(devices[visitIndex%len(devices)])
				entity.Country = nullable(geo.code)
				entity.Language = nullable("en")
				entity.ScreenWidth = sql.NullInt32{Int32: int32([]int{1440, 390, 1024}[visitIndex%3]), Valid: true}
				entity.VisitorHash = nullable(fmt.Sprintf("%s:%02d:visitor-%d", websiteID, dayIndex, visitIndex%3))
				entity.CountryCode, entity.CountryName, entity.City, entity.Region = nullable(geo.code), nullable(geo.name), nullable(geo.city), nullable(geo.region)
				pageviews = append(pageviews, entity)
				if visitIndex%2 == 0 {
					event := factories.BuildEvent(websiteID)
					event.ID = deterministicID(fmt.Sprintf("event:%d:%d:%d", siteIndex, dayIndex, visitIndex))
					event.CreatedAt, event.Url, event.EventName = entity.CreatedAt.Add(10*time.Minute), entity.Url, eventNames[(dayIndex+visitIndex)%len(eventNames)]
					event.EventData = json.RawMessage(fmt.Sprintf(`{"source":"demo","value":%d}`, dayIndex+visitIndex))
					event.VisitorHash, event.CountryCode, event.CountryName, event.City, event.Region = entity.VisitorHash, entity.CountryCode, entity.CountryName, entity.City, entity.Region
					events = append(events, event)
				}
			}
		}
	}
	if len(pageviews) > 0 {
		if _, err := exec.NewInsert().Model(&pageviews).Exec(ctx); err != nil {
			return fmt.Errorf("seed pageviews: %w", err)
		}
	}
	if len(events) > 0 {
		if _, err := exec.NewInsert().Model(&events).Exec(ctx); err != nil {
			return fmt.Errorf("seed events: %w", err)
		}
	}
	return nil
}

func ensureSeedUser(ctx context.Context, exec storage.Executor, email string, admin bool) (models.UserEntity, error) {
	password := factories.BuildUser(factories.WithDefaultPassword()).Password
	user, err := models.User.FindByEmail(ctx, exec, email)
	if errors.Is(err, models.ErrNotFound) {
		return factories.CreateUser(ctx, exec, factories.WithEmail(email), factories.WithIsAdmin(admin), factories.WithValidatedEmail(), factories.WithPassword(password))
	}
	if err != nil {
		return models.UserEntity{}, err
	}
	if _, err := exec.NewUpdate().Model(&user).
		Set("is_admin = ?", admin).
		Set("password = ?", password).
		Set("email_validated_at = COALESCE(email_validated_at, ?)", time.Now().UTC()).
		WherePK().Exec(ctx); err != nil {
		return models.UserEntity{}, err
	}
	user.IsAdmin = admin
	user.Password = password
	return user, nil
}

func deterministicID(value string) uuid.UUID {
	sum := sha256.Sum256([]byte(value))
	id, _ := uuid.FromBytes(sum[:16])
	return id
}

func nullable(value string) sql.NullString { return sql.NullString{String: value, Valid: value != ""} }

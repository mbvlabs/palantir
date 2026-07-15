package models_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"palantir/database"
	"palantir/internal/storage"
	"palantir/models"
	"palantir/models/factories"
)

var modelTestCluster *storage.TestCluster

func TestMain(m *testing.M) {
	ctx := context.Background()
	cluster, err := storage.NewTestCluster(ctx)
	if err != nil {
		panic(err)
	}
	modelTestCluster = cluster
	code := m.Run()
	_ = cluster.Close(ctx)
	os.Exit(code)
}

func TestWebsiteLifecycleValidationAndOwnerIsolation(t *testing.T) {
	db := modelTestCluster.NewTestDB(t, database.Migrations, "migrations")
	ctx := context.Background()
	owner, err := factories.CreateUser(ctx, db.Executor())
	if err != nil {
		t.Fatal(err)
	}
	other, err := factories.CreateUser(ctx, db.Executor())
	if err != nil {
		t.Fatal(err)
	}

	website, err := models.Website.Create(ctx, db.Executor(), models.CreateWebsiteData{
		UserID: owner.ID,
		Name:   "Example",
		Domain: "example.com",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	owned, err := models.Website.AllOwned(ctx, db.Executor(), owner.ID)
	if err != nil || len(owned) != 1 || owned[0].ID != website.ID {
		t.Fatalf("owned websites = %+v, err = %v", owned, err)
	}
	if _, err := models.Website.FindOwned(ctx, db.Executor(), website.ID, other.ID); !errors.Is(err, models.ErrNotFound) {
		t.Fatalf("foreign find error = %v, want ErrNotFound", err)
	}
	if _, err := models.Website.UpdateOwned(ctx, db.Executor(), other.ID, models.UpdateWebsiteData{ID: website.ID, Name: "Stolen", Domain: "stolen.example"}); !errors.Is(err, models.ErrNotFound) {
		t.Fatalf("foreign update error = %v, want ErrNotFound", err)
	}
	if err := models.Website.DestroyOwned(ctx, db.Executor(), website.ID, other.ID); !errors.Is(err, models.ErrNotFound) {
		t.Fatalf("foreign destroy error = %v, want ErrNotFound", err)
	}

	updated, err := models.Website.UpdateOwned(ctx, db.Executor(), owner.ID, models.UpdateWebsiteData{ID: website.ID, Name: "Updated", Domain: "updated.example"})
	if err != nil || updated.Name != "Updated" {
		t.Fatalf("update = %+v, err = %v", updated, err)
	}

	before, err := models.Website.AllOwned(ctx, db.Executor(), owner.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := models.Website.Create(ctx, db.Executor(), models.CreateWebsiteData{UserID: owner.ID, Name: "", Domain: "not a domain"}); !errors.Is(err, models.ErrDomainValidation) {
		t.Fatalf("invalid create error = %v, want ErrDomainValidation", err)
	}
	after, err := models.Website.AllOwned(ctx, db.Executor(), owner.ID)
	if err != nil || len(after) != len(before) {
		t.Fatalf("invalid create wrote a row: before=%d after=%d err=%v", len(before), len(after), err)
	}

	if err := models.Website.DestroyOwned(ctx, db.Executor(), website.ID, owner.ID); err != nil {
		t.Fatalf("destroy: %v", err)
	}
	if _, err := models.Website.FindOwned(ctx, db.Executor(), website.ID, owner.ID); !errors.Is(err, models.ErrNotFound) {
		t.Fatalf("find destroyed error = %v, want ErrNotFound", err)
	}
}

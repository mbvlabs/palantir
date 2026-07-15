package controllers

import (
	"context"
	"os"
	"testing"

	"palantir/internal/inertia"
	"palantir/internal/storage"
)

var controllerTestCluster *storage.TestCluster

func TestMain(m *testing.M) {
	ctx := context.Background()
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	if err := os.Chdir(".."); err != nil {
		panic(err)
	}
	if err := inertia.Init(); err != nil {
		panic(err)
	}
	if err := os.Chdir(cwd); err != nil {
		panic(err)
	}

	cluster, err := storage.NewTestCluster(ctx)
	if err != nil {
		panic(err)
	}
	controllerTestCluster = cluster
	code := m.Run()
	_ = cluster.Close(ctx)
	os.Exit(code)
}

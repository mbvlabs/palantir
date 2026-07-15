package main

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	mailclients "palantir/clients/email"
	"palantir/config"
	"palantir/internal/server"
)

func TestEmailSendersUseAwsSesOnlyInProduction(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		wantAwsSes  bool
	}{
		{name: "production", environment: server.ProdEnvironment, wantAwsSes: true},
		{name: "development", environment: server.DevEnvironment, wantAwsSes: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			transactional, marketing := newEmailSenders(config.Config{}, test.environment)
			_, transactionalIsAwsSes := transactional.(*mailclients.AwsSes)
			_, marketingIsAwsSes := marketing.(*mailclients.AwsSes)
			if transactionalIsAwsSes != test.wantAwsSes || marketingIsAwsSes != test.wantAwsSes {
				t.Fatalf("AWS SES senders = %t/%t, want %t", transactionalIsAwsSes, marketingIsAwsSes, test.wantAwsSes)
			}
		})
	}
}

func TestBackgroundLifecycleStopsOnceAndWaitsForExit(t *testing.T) {
	appCtx, cancelApp := context.WithCancel(context.Background())
	defer cancelApp()
	started := make(chan struct{})
	release := make(chan struct{})
	done := startInBackground(appCtx, "test worker", func(ctx context.Context) error {
		if ctx != appCtx {
			t.Errorf("worker context does not match application context")
		}
		close(started)
		<-release
		return nil
	})
	<-started

	var stopCalls atomic.Int32
	stop := func(context.Context) error {
		if stopCalls.Add(1) == 1 {
			close(release)
		}
		return nil
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := stopAndWait(shutdownCtx, stop, done); err != nil {
		t.Fatalf("stopAndWait: %v", err)
	}
	if got := stopCalls.Load(); got != 1 {
		t.Fatalf("stop calls = %d, want 1", got)
	}
	select {
	case <-done:
	default:
		t.Fatal("background worker still running after shutdown")
	}
}

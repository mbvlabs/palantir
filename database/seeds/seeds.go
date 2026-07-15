package seeds

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"palantir/internal/storage"
	"palantir/models/factories"
)

const Default = "development"

type Runner func(context.Context, storage.Executor) error

var Registry = map[string]Runner{
	"default":     Development,
	"development": Development,
	"test":        Test,
}

func Names() []string {
	names := make([]string, 0, len(Registry))
	for name := range Registry {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

func Run(ctx context.Context, exec storage.Executor, name string) error {
	if name == "" {
		name = Default
	}

	runner, ok := Registry[name]
	if !ok {
		return fmt.Errorf("unknown seed %q (available: %s)", name, strings.Join(Names(), ", "))
	}

	return runner(ctx, exec)
}

func Development(ctx context.Context, exec storage.Executor) error {
	return seedPalantir(ctx, exec)
}

func Test(ctx context.Context, exec storage.Executor) error {
	_, err := factories.CreateUser(ctx, exec,
		factories.WithEmail("test@example.com"),
		factories.WithValidatedEmail(),
	)
	if err != nil {
		return fmt.Errorf("failed to create test user: %w", err)
	}

	return nil
}

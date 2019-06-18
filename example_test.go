package golog_test

import (
    "context"

	"github.com/vardius/golog"
)

func ExampleDebug() {
    ctx := context.Background()
	logger := golog.New(golog.Debug)

	logger.Debug(ctx context.Context, "Hello %s!", "you")

	// Output:
	// DEBUG: Hello you!
}

func ExampleInfo() {
    ctx := context.Background()
	logger := golog.New(golog.Info)

	logger.Info(ctx context.Context, "Hello %s!", "you")

	// Output:
	// INFO: Hello you!
}

func ExampleWarning() {
    ctx := context.Background()
	logger := golog.New(golog.Warning)

	logger.Warning(ctx context.Context, "Hello %s!", "you")

	// Output:
	// WARN: Hello you!
}

func ExampleError() {
    ctx := context.Background()
	logger := golog.New(golog.Error)

	logger.Error(ctx context.Context, "Hello %s!", "you")

	// Output:
	// ERROR: Hello you!
}

func ExampleCritical() {
    ctx := context.Background()
	logger := golog.New(golog.Critical)

	logger.Critical(ctx context.Context, "Hello %s!", "you")

	// Output:
	// FATAL: Hello you!
}


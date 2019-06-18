package golog_test

import (
    "context"

	"github.com/vardius/golog"
)

func ExampleNewConsoleLogger() {
	ctx := context.Background()
	log := golog.NewConsoleLogger(Warning)

	// log.Info won't print to the console because it is below loglevel "warn"
	log.Info(ctx, "%s", "Info")
	// log.Warning and log.Error will both get printed
	log.Warning(ctx, "%s %d", "Warn", 1)
	log.Error(ctx, "%s %d", "Error", 666)
}

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


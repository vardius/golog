package golog_test

import (
	"context"

	"github.com/vardius/golog"
)

func ExampleNewConsoleLogger() {
	ctx := context.Background()
	logger := golog.New(golog.Warning)

	logger.SetFlags(0)

	// logger.Info won't print to the console because it is below loglevel "warn"
	logger.Info(ctx, "%s", "Info")
	// logger.Warning and log.Error will both get printed
	logger.Warning(ctx, "%s %d", "Warn", 1)
	logger.Error(ctx, "%s %d", "Error", 666)

	// Output:
	// [33;1mWARN: Warn 1
	// [31;1mERROR: Error 666
}

func ExampleNewConsoleLogger_second() {
	ctx := context.Background()

	logger := golog.NewConsoleLogger(golog.Verbose(2))

	logger.SetFlags(0)

	// logger.Info won't print to the console because it is below loglevel "warn"
	logger.Info(ctx, "%s", "Info")
	// logger.Warning and log.Error will both get printed
	logger.Warning(ctx, "%s %d", "Warn", 1)
	logger.Error(ctx, "%s %d", "Error", 666)

	// Output:
	// [33;1mWARN: Warn 1
	// [31;1mERROR: Error 666
}

func ExampleDebug() {
	ctx := context.Background()
	logger := golog.New(golog.Debug)

	logger.SetFlags(0)
	logger.Debug(ctx, "Hello %s!", "you")

	// Output:
	// [37;1mDEBUG: Hello you!
}

func ExampleInfo() {
	ctx := context.Background()
	logger := golog.New(golog.Info)

	logger.SetFlags(0)
	logger.Info(ctx, "Hello %s!", "you")

	// Output:
	// [36;1mINFO: Hello you!
}

func ExampleWarning() {
	ctx := context.Background()
	logger := golog.New(golog.Warning)

	logger.SetFlags(0)
	logger.Warning(ctx, "Hello %s!", "you")

	// Output:
	// [33;1mWARN: Hello you!
}

func ExampleError() {
	ctx := context.Background()
	logger := golog.New(golog.Error)

	logger.SetFlags(0)
	logger.Error(ctx, "Hello %s!", "you")

	// Output:
	// [31;1mERROR: Hello you!
}

func ExampleCritical() {
	ctx := context.Background()
	logger := golog.New(golog.Critical)

	logger.SetFlags(0)
	logger.Critical(ctx, "Hello %s!", "you")

	// Output:
	// [31;1mFATAL: Hello you!
}

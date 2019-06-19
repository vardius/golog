package golog_test

import (
	"context"

	"github.com/vardius/golog"
)

func ExampleNewConsoleLogger() {
	ctx := context.Background()
	log := golog.New(golog.Warning)

	// log.Info won't print to the console because it is below loglevel "warn"
	log.Info(ctx, "%s", "Info")
	// log.Warning and log.Error will both get printed
	log.Warning(ctx, "%s %d", "Warn", 1)
	log.Error(ctx, "%s %d", "Error", 666)
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
	// [31;1mEERROR: Hello you!
}

func ExampleCritical() {
	ctx := context.Background()
	logger := golog.New(golog.Critical)

	logger.SetFlags(0)
	logger.Critical(ctx, "Hello %s!", "you")

	// Output:
	// [31;1mFATAL: Hello you!
}

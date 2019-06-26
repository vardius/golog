package golog_test

import (
	"context"

	"github.com/vardius/golog"
)

func ExampleNewConsoleLogger() {
	ctx := context.Background()
	logger := golog.NewConsoleLogger()
	logger.SetVerbosity(golog.Warning | golog.Error)

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
	logger := golog.NewConsoleLogger()
	logger.SetVerbosity(golog.Warning | golog.Error)

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
	logger := golog.New()
	logger.SetVerbosity(golog.Debug)

	logger.SetFlags(0)
	logger.Debug(ctx, "Hello %s!", "you")

	// Output:
	// [37;1mDEBUG: Hello you!
}

func ExampleInfo() {
	ctx := context.Background()
	logger := golog.New()
	logger.SetVerbosity(golog.Info)

	logger.SetFlags(0)
	logger.Info(ctx, "Hello %s!", "you")

	// Output:
	// [36;1mINFO: Hello you!
}

func ExampleWarning() {
	ctx := context.Background()
	logger := golog.New()
	logger.SetVerbosity(golog.Warning)

	logger.SetFlags(0)
	logger.Warning(ctx, "Hello %s!", "you")

	// Output:
	// [33;1mWARN: Hello you!
}

func ExampleError() {
	ctx := context.Background()
	logger := golog.New()
	logger.SetVerbosity(golog.Error)

	logger.SetFlags(0)
	logger.Error(ctx, "Hello %s!", "you")

	// Output:
	// [31;1mERROR: Hello you!
}

func ExampleCritical() {
	ctx := context.Background()
	logger := golog.New()
	logger.SetVerbosity(golog.Critical)

	logger.SetFlags(0)
	logger.Critical(ctx, "Hello %s!", "you")

	// Output:
	// [31;1mFATAL: Hello you!
}

func ExampleDisable() {
	ctx := context.Background()
	logger := golog.New()
	logger.SetVerbosity(golog.Disabled)

	logger.SetFlags(0)
	logger.Debug(ctx, "Hello %s!", "you")
	logger.Info(ctx, "Hello %s!", "you")
	logger.Warning(ctx, "Hello %s!", "you")
	logger.Error(ctx, "Hello %s!", "you")
	logger.Critical(ctx, "Hello %s!", "you")

	// Output:
	//
}

func ExampleDisable_second() {
	ctx := context.Background()

	logger := golog.New()
	logger.SetVerbosity(golog.Debug | golog.Info | golog.Warning | golog.Error | golog.Critical | golog.Disabled)

	logger.SetFlags(0)
	logger.Debug(ctx, "Hello %s!", "you")
	logger.Info(ctx, "Hello %s!", "you")
	logger.Warning(ctx, "Hello %s!", "you")
	logger.Error(ctx, "Hello %s!", "you")
	logger.Critical(ctx, "Hello %s!", "you")

	// Output:
	//
}

func ExampleDisable_third() {
	ctx := context.Background()

	// do not disable logger
	logger := golog.New()
	logger.SetVerbosity(golog.Debug | golog.Info | golog.Warning | golog.Error | golog.Critical)

	logger.SetFlags(0)
	logger.Debug(ctx, "Hello %s!", "you")
	logger.Info(ctx, "Hello %s!", "you")
	logger.Warning(ctx, "Hello %s!", "you")
	logger.Error(ctx, "Hello %s!", "you")
	logger.Critical(ctx, "Hello %s!", "you")

	// Output:
	// [37;1mDEBUG: Hello you!
	// [36;1mINFO: Hello you!
	// [33;1mWARN: Hello you!
	// [31;1mERROR: Hello you!
	// [31;1mFATAL: Hello you!
}

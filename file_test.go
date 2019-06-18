package golog_test

import (
	"context"

	"github.com/vardius/golog"
)

func ExampleNewConsoleLogger() {
	ctx := context.Background()
	log := golog.NewConsoleLogger("warn")

	// log.Info won't print to the console because it is below loglevel "warn"
	log.Info(ctx, "%s", "Info")
	// log.Warning and log.Error will both get printed
	log.Warning(ctx, "%s %d", "Warn", 1)
	log.Error(ctx, "%s %d", "Error", 666)
}

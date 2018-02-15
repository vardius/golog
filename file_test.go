package golog_test

import "github.com/vardius/golog"
import "context"

func ExampleNewConsoleLogger() {
	log := golog.NewConsoleLogger("warn")
	ctx := context.TODO()

	// log.Info won't print to the console because it is below loglevel "warn"
	log.Info(ctx, "%s", "Info")
	// log.Warning and log.Error will both get printed
	log.Warning(ctx, "%s %d", "Warn", 1)
	log.Error(ctx, "%s %d", "Error", 666)
}

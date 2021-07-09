/*
Package golog provides simple logger
Basic example:
	package main

	import (
		"fmt"
		"context"

		"github.com/vardius/golog"
	)

	func main() {
		ctx := context.Background()
		logger := golog.New()

		logger.Debug(ctx context.Context, fmt.Sprintf("Hello %s!", "you"))
	}
As a package:
	package mylogger

	import (
		"context"

		"github.com/vardius/golog"
	)

	var logger golog.Logger

	func SetFlags(flag int) {
		logger.SetFlags(flag)
	}

	func SetVerbosity(verbosity golog.Verbose) {
		logger.SetVerbosity(verbosity)
	}

	func Debug(ctx context.Context, v string) {
		logger.Debug(ctx, v)
	}

	func Info(ctx context.Context, v string) {
		logger.Info(ctx, v)
	}

	func Warning(ctx context.Context, v string) {
		logger.Warning(ctx, v)
	}

	func Error(ctx context.Context, v string) {
		logger.Error(ctx, v)
	}

	func Critical(ctx context.Context, v string) {
		logger.Critical(ctx, v)
	}

	func init() {
		logger = golog.New()
	}
usage:
	package main

	import (
		"fmt"

		"mylogger"
	)

	func main() {
		mylogger.Debug(ctx context.Context, fmt.Sprintf("Hello %s!", "you"))
	}
*/
package golog

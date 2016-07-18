package golog

import (
	"golang.org/x/net/context"
)

type (
	Logger interface {
		Debug(ctx context.Context, format string, args ...interface{})
		Info(ctx context.Context, format string, args ...interface{})
		Warning(ctx context.Context, format string, args ...interface{})
		Critical(ctx context.Context, format string, args ...interface{})
	}

	loggerFactory func() Logger
)

var (
	New loggerFactory
)

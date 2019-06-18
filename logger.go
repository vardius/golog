package golog

import (
	"context"
	"log"
)

type Verbose int

const (
	DebugPrefix = "DEBUG: "
	InfoPrefix  = "INFO: "
	WarnPrefix  = "WARN: "
	ErrorPrefix = "ERROR: "
	FatalPrefix = "FATAL: "

	DefaultFlags = log.Ldate | log.Ltime | log.Lmicroseconds | log.LUTC

	Disabled Verbose = -1
	Critical Verbose = iota
	Error
	Warning
	Info
	Debug
)

type (
	Logger interface {
		Debug(ctx context.Context, format string, args ...interface{})
		Info(ctx context.Context, format string, args ...interface{})
		Warning(ctx context.Context, format string, args ...interface{})
		Error(ctx context.Context, format string, args ...interface{})
		Critical(ctx context.Context, format string, args ...interface{})
	}

	loggerFactory func(level Verbose) Logger
)

var New loggerFactory

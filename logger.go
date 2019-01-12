package golog

import (
	"context"
	"log"
)

const (
	DebugPrefix = "DEBUG: "
	InfoPrefix  = "INFO: "
	WarnPrefix  = "WARN: "
	ErrorPrefix = "ERROR: "
	FatalPrefix = "FATAL: "

	DebugLevel    = 4
	InfoLevel     = 3
	WarningLevel  = 2
	ErrorLevel    = 1
	CriticalLevel = 0

	DefaultFlags = log.Ldate | log.Ltime | log.Lmicroseconds | log.LUTC
)

type (
	Logger interface {
		Debug(ctx context.Context, format string, args ...interface{})
		Info(ctx context.Context, format string, args ...interface{})
		Warning(ctx context.Context, format string, args ...interface{})
		Error(ctx context.Context, format string, args ...interface{})
		Critical(ctx context.Context, format string, args ...interface{})
	}

	loggerFactory func(verbose string) Logger
)

var New loggerFactory

func parseVerboseLevel(verbose string) int {
	switch verbose {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warning":
		return WarningLevel
	case "error":
		return ErrorLevel
	case "critical":
		return CriticalLevel
	default:
		return -1 // logger disabled by default
	}
}

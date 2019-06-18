// +build !appengine,!appenginevm

package golog

import (
	"context"
	"log"
	"os"
)

type (
	fileLogger struct {
		verboseLevel Verbose

		debug    *log.Logger
		info     *log.Logger
		warning  *log.Logger
		error    *log.Logger
		critical *log.Logger
	}
)

func NewFileLogger(level Verbose, filePath string) Logger {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil
	}

	return &fileLogger{
		level,
		log.New(file, DebugPrefix, DefaultFlags),
		log.New(file, InfoPrefix, DefaultFlags),
		log.New(file, WarnPrefix, DefaultFlags),
		log.New(file, ErrorPrefix, DefaultFlags),
		log.New(file, FatalPrefix, DefaultFlags),
	}
}

func (logger *fileLogger) SetFlags(flag int) {
	logger.debug.SetFlags(flag)
	logger.info.SetFlags(flag)
	logger.warning.SetFlags(flag)
	logger.error.SetFlags(flag)
	logger.critical.SetFlags(flag)
}

func (logger *fileLogger) Debug(ctx context.Context, format string, args ...interface{}) {
	if logger.verboseLevel >= Debug {
		logger.debug.Printf(format, args...)
	}
}

func (logger *fileLogger) Info(ctx context.Context, format string, args ...interface{}) {
	if logger.verboseLevel >= Info {
		logger.info.Printf(format, args...)
	}
}

func (logger *fileLogger) Warning(ctx context.Context, format string, args ...interface{}) {
	if logger.verboseLevel >= Warning {
		logger.warning.Printf(format, args...)
	}
}

func (logger *fileLogger) Error(ctx context.Context, format string, args ...interface{}) {
	if logger.verboseLevel >= Error {
		logger.error.Printf(format, args...)
	}
}

func (logger *fileLogger) Critical(ctx context.Context, format string, args ...interface{}) {
	if logger.verboseLevel >= Critical {
		logger.critical.Printf(format, args...)
	}
}

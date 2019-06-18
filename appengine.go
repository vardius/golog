// +build appengine appenginevm

package golog

import (
	"context"

	"google.golang.org/appengine/log"
)

type appengineLogger struct {
	verboseLevel Verbose
}

func NewAppengineLogger(level Verbose) Logger {
	return &appengineLogger{level}
}

func (logger *appengineLogger) Debug(ctx context.Context, format string, args ...interface{}) {
	if logger.verboseLevel >= Debug {
		log.Debugf(ctx, format, args...)
	}
}

func (logger *appengineLogger) Info(ctx context.Context, format string, args ...interface{}) {
	if logger.verboseLevel >= Info {
		log.Infof(ctx, format, args...)
	}
}

func (logger *appengineLogger) Warning(ctx context.Context, format string, args ...interface{}) {
	if logger.verboseLevel >= Warning {
		log.Warningf(ctx, format, args...)
	}
}

func (logger *appengineLogger) Error(ctx context.Context, format string, args ...interface{}) {
	if logger.verboseLevel >= Error {
		log.Errorf(ctx, format, args...)
	}
}

func (logger *appengineLogger) Critical(ctx context.Context, format string, args ...interface{}) {
	if logger.verboseLevel >= Critical {
		log.Criticalf(ctx, format, args...)
	}
}

func init() {
	New = NewAppengineLogger
}

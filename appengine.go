// +build appengine appenginevm

package golog

import (
	"context"

	"google.golang.org/appengine/log"
)

type appengineLogger struct {
	verbosity Verbose
}

func NewAppengineLogger() Logger {
	return &appengineLogger{DefaultVerbosity}
}

func (l *appengineLogger) SetFlags(flag int) {}

func (l *appengineLogger) SetVerbosity(verbosity Verbose) {
	l.verbosity = verbosity
}

func (l *appengineLogger) Debug(ctx context.Context, format string, args ...interface{}) {
	if l.verbosity&Disabled != 0 {
		return
	}
	if l.verbosity&Debug != 0 {
		log.Debugf(ctx, format, args...)
	}
}

func (l *appengineLogger) Info(ctx context.Context, format string, args ...interface{}) {
	if l.verbosity&Disabled != 0 {
		return
	}
	if l.verbosity&Info != 0 {
		log.Infof(ctx, format, args...)
	}
}

func (l *appengineLogger) Warning(ctx context.Context, format string, args ...interface{}) {
	if l.verbosity&Disabled != 0 {
		return
	}
	if l.verbosity&Warning != 0 {
		log.Warningf(ctx, format, args...)
	}
}

func (l *appengineLogger) Error(ctx context.Context, format string, args ...interface{}) {
	if l.verbosity&Disabled != 0 {
		return
	}
	if l.verbosity&Error != 0 {
		log.Errorf(ctx, format, args...)
	}
}

func (l *appengineLogger) Critical(ctx context.Context, format string, args ...interface{}) {
	if l.verbosity&Disabled != 0 {
		return
	}
	if l.verbosity&Critical != 0 {
		log.Criticalf(ctx, format, args...)
	}
}

func init() {
	New = NewAppengineLogger
}

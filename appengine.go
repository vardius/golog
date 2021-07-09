// +build appengine appenginevm

package golog

import (
	"context"

	"google.golang.org/appengine/log"
)

type appEngineLogger struct {
	verbosity Verbose
}

func NewAppEngineLogger() Logger {
	return &appEngineLogger{DefaultVerbosity}
}

func (l *appEngineLogger) SetFlags(flag int) {}

func (l *appEngineLogger) SetVerbosity(verbosity Verbose) {
	l.verbosity = verbosity
}

func (l *appEngineLogger) Debug(ctx context.Context, v string) {
	if l.verbosity&Disabled != 0 {
		return
	}
	if l.verbosity&Debug != 0 {
		log.Debugf(ctx, "%s", v)
	}
}

func (l *appEngineLogger) Info(ctx context.Context, v string) {
	if l.verbosity&Disabled != 0 {
		return
	}
	if l.verbosity&Info != 0 {
		log.Infof(ctx, "%s", v)
	}
}

func (l *appEngineLogger) Warning(ctx context.Context, v string) {
	if l.verbosity&Disabled != 0 {
		return
	}
	if l.verbosity&Warning != 0 {
		log.Warningf(ctx, "%s", v)
	}
}

func (l *appEngineLogger) Error(ctx context.Context, v string) {
	if l.verbosity&Disabled != 0 {
		return
	}
	if l.verbosity&Error != 0 {
		log.Errorf(ctx, "%s", v)
	}
}

func (l *appEngineLogger) Critical(ctx context.Context, v string) {
	if l.verbosity&Disabled != 0 {
		return
	}
	if l.verbosity&Critical != 0 {
		log.Criticalf(ctx, "%s", v)
	}
}

func (l *appEngineLogger) Fatal(ctx context.Context, v string) {
	if l.verbosity&Disabled == 0 {
		log.Criticalf(ctx, "%s", v)
	}
	os.Exit(1)
}

func init() {
	New = NewAppEngineLogger
}

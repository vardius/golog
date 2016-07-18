// +build appengine appenginevm

package golog

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
)

type (
	appengineLogger struct{}
)

func NewAppengineLogger() Logger {
	return &appengineLogger{}
}

func (logger *appengineLogger) Debug(ctx context.Context, format string, args ...interface{}) {
	log.Debugf(ctx, format, args...)
}

func (logger *appengineLogger) Info(ctx context.Context, format string, args ...interface{}) {
	log.Infof(ctx, format, args...)
}

func (logger *appengineLogger) Warning(ctx context.Context, format string, args ...interface{}) {
	log.Warningf(ctx, format, args...)
}

func (logger *appengineLogger) Error(ctx context.Context, format string, args ...interface{}) {
	log.Errorf(ctx, format, args...)
}

func (logger *appengineLogger) Critical(ctx context.Context, format string, args ...interface{}) {
	log.Criticalf(ctx, format, args...)
}

func init() {
	New = NewAppengineLogger
}

// +build !appengine,!appenginevm

package golog

import (
	"context"
	"log"
	"os"
)

type consoleLogger struct {
	verbosity Verbose
	logger    *log.Logger
}

// Terminal colours.
const (
	CLR_0 = "\x1b[30;1m"
	CLR_R = "\x1b[31;1m"
	CLR_G = "\x1b[32;1m"
	CLR_Y = "\x1b[33;1m"
	CLR_B = "\x1b[34;1m"
	CLR_M = "\x1b[35;1m"
	CLR_C = "\x1b[36;1m"
	CLR_W = "\x1b[37;1m"
	CLR_N = "\x1b[0m"
)

// NewConsoleLogger returns a Logger that writes to the console.
func NewConsoleLogger() Logger {
	return &consoleLogger{
		verbosity: DefaultVerbosity,
		logger:    log.New(os.Stdout, "", DefaultFlags),
	}
}

func (l *consoleLogger) SetFlags(flag int) {
	l.logger.SetFlags(flag)
}

func (l *consoleLogger) SetVerbosity(verbosity Verbose) {
	l.verbosity = verbosity
}

func (l *consoleLogger) Debug(ctx context.Context, format string, args ...interface{}) {
	if l.verbosity&Disabled != 0 {
		return
	}
	if l.verbosity&Debug != 0 {
		l.logger.Printf("%s"+format, append([]interface{}{CLR_W + DebugPrefix}, args...)...)
	}
}

func (l *consoleLogger) Info(ctx context.Context, format string, args ...interface{}) {
	if l.verbosity&Disabled != 0 {
		return
	}
	if l.verbosity&Info != 0 {
		l.logger.Printf("%s"+format, append([]interface{}{CLR_C + InfoPrefix}, args...)...)
	}
}

func (l *consoleLogger) Warning(ctx context.Context, format string, args ...interface{}) {
	if l.verbosity&Disabled != 0 {
		return
	}
	if l.verbosity&Warning != 0 {
		l.logger.Printf("%s"+format, append([]interface{}{CLR_Y + WarnPrefix}, args...)...)
	}
}

func (l *consoleLogger) Error(ctx context.Context, format string, args ...interface{}) {
	if l.verbosity&Disabled != 0 {
		return
	}
	if l.verbosity&Error != 0 {
		l.logger.Printf("%s"+format, append([]interface{}{CLR_R + ErrorPrefix}, args...)...)
	}
}

func (l *consoleLogger) Critical(ctx context.Context, format string, args ...interface{}) {
	if l.verbosity&Disabled != 0 {
		return
	}
	if l.verbosity&Critical != 0 {
		l.logger.Printf("%s"+format, append([]interface{}{CLR_R + FatalPrefix}, args...)...)
	}
}

func init() {
	New = NewConsoleLogger
}

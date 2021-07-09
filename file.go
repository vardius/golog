// +build !appengine,!appenginevm

package golog

import (
	"context"
	"log"
	"os"
)

type (
	fileLogger struct {
		verbosity Verbose
		logger    *log.Logger
	}
)

// NewFileLogger returns a Logger that writes to the file.
func NewFileLogger(filePath string) Logger {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil
	}

	return &fileLogger{
		verbosity: DefaultVerbosity,
		logger:    log.New(file, "", DefaultFlags),
	}
}

func (l *fileLogger) SetFlags(flag int) {
	l.logger.SetFlags(flag)
}

func (l *fileLogger) SetVerbosity(verbosity Verbose) {
	l.verbosity = verbosity
}

func (l *fileLogger) Debug(_ context.Context, v string) {
	if l.verbosity&Disabled != 0 {
		return
	}
	if l.verbosity&Debug != 0 {
		l.logger.Printf("%s%s", DebugPrefix, v)
	}
}

func (l *fileLogger) Info(_ context.Context, v string) {
	if l.verbosity&Disabled != 0 {
		return
	}
	if l.verbosity&Info != 0 {
		l.logger.Printf("%s%s", InfoPrefix, v)
	}
}

func (l *fileLogger) Warning(_ context.Context, v string) {
	if l.verbosity&Disabled != 0 {
		return
	}
	if l.verbosity&Warning != 0 {
		l.logger.Printf("%s%s", WarnPrefix, v)
	}
}

func (l *fileLogger) Error(_ context.Context, v string) {
	if l.verbosity&Disabled != 0 {
		return
	}
	if l.verbosity&Error != 0 {
		l.logger.Printf("%s%s", ErrorPrefix, v)
	}
}

func (l *fileLogger) Critical(_ context.Context, v string) {
	if l.verbosity&Disabled != 0 {
		return
	}
	if l.verbosity&Critical != 0 {
		l.logger.Printf("%s%s", CriticalPrefix, v)
	}
}

func (l *fileLogger) Fatal(_ context.Context, v string) {
	if l.verbosity&Disabled == 0 {
		l.logger.Printf("%s%s", CriticalPrefix, v)
	}
	os.Exit(1)
}

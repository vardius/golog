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

func (l *fileLogger) Debug(ctx context.Context, format string, args ...interface{}) {
	if l.verbosity&Disabled != 0 {
		return
	}
	if l.verbosity&Debug != 0 {
		l.logger.Printf("%s"+format, append([]interface{}{DebugPrefix}, args...)...)
	}
}

func (l *fileLogger) Info(ctx context.Context, format string, args ...interface{}) {
	if l.verbosity&Disabled != 0 {
		return
	}
	if l.verbosity&Info != 0 {
		l.logger.Printf("%s"+format, append([]interface{}{InfoPrefix}, args...)...)
	}
}

func (l *fileLogger) Warning(ctx context.Context, format string, args ...interface{}) {
	if l.verbosity&Disabled != 0 {
		return
	}
	if l.verbosity&Warning != 0 {
		l.logger.Printf("%s"+format, append([]interface{}{WarnPrefix}, args...)...)
	}
}

func (l *fileLogger) Error(ctx context.Context, format string, args ...interface{}) {
	if l.verbosity&Disabled != 0 {
		return
	}
	if l.verbosity&Error != 0 {
		l.logger.Printf("%s"+format, append([]interface{}{ErrorPrefix}, args...)...)
	}
}

func (l *fileLogger) Critical(ctx context.Context, format string, args ...interface{}) {
	if l.verbosity&Disabled != 0 {
		return
	}
	if l.verbosity&Critical != 0 {
		l.logger.Printf("%s"+format, append([]interface{}{FatalPrefix}, args...)...)
	}
}

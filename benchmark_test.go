package golog_test

import (
	"context"
	"io"
	"log/slog"
	"strconv"
	"testing"

	"github.com/vardius/golog/v2"
)

// Every benchmark logs through a *slog.Logger to io.Discard, so what is measured is the Logger and the handler:
// the Logger's own cost (the time, the caller's PC, the Record) is the same for every handler.

func BenchmarkHandler(b *testing.B) {
	for _, n := range []int{0, 5, 10} {
		b.Run(strconv.Itoa(n)+" attrs", func(b *testing.B) {
			logger := slog.New(golog.NewHandler(io.Discard, nil))
			as := attrs(n)
			b.ReportAllocs()
			for b.Loop() {
				logger.LogAttrs(context.Background(), slog.LevelInfo, "request", as...)
			}
		})
	}
}

func BenchmarkHandlerGroups(b *testing.B) {
	logger := slog.New(golog.NewHandler(io.Discard, nil)).With("service", "api", "version", 2).WithGroup("http")
	group := slog.Attr{Key: "req", Value: slog.GroupValue(attrs(5)...)}
	b.ReportAllocs()
	for b.Loop() {
		logger.LogAttrs(context.Background(), slog.LevelInfo, "request", group)
	}
}

func BenchmarkHandlerColor(b *testing.B) {
	logger := slog.New(golog.NewHandler(io.Discard, &golog.Options{Color: true}))
	as := attrs(5)
	b.ReportAllocs()
	for b.Loop() {
		logger.LogAttrs(context.Background(), slog.LevelInfo, "request", as...)
	}
}

func BenchmarkHandlerSource(b *testing.B) {
	logger := slog.New(golog.NewHandler(io.Discard, &golog.Options{AddSource: true}))
	as := attrs(5)
	b.ReportAllocs()
	for b.Loop() {
		logger.LogAttrs(context.Background(), slog.LevelInfo, "request", as...)
	}
}

func BenchmarkHandlerDisabled(b *testing.B) {
	logger := slog.New(golog.NewHandler(io.Discard, nil))
	as := attrs(5)
	b.ReportAllocs()
	for b.Loop() {
		logger.LogAttrs(context.Background(), slog.LevelDebug, "request", as...)
	}
}

func BenchmarkHandlerParallel(b *testing.B) {
	logger := slog.New(golog.NewHandler(io.Discard, nil))
	as := attrs(5)
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.LogAttrs(context.Background(), slog.LevelInfo, "request", as...)
		}
	})
}

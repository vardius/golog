package golog_test

import (
	"context"
	"errors"
	"log/slog"
	"os"

	"github.com/vardius/golog/v2"
)

// withoutTime leaves the time out, so the output below doesn't change from run to run.
func withoutTime(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.TimeKey && groups == nil {
		return slog.Attr{}
	}
	return a
}

func ExampleNewHandler() {
	logger := slog.New(golog.NewHandler(os.Stdout, &golog.Options{ReplaceAttr: withoutTime}))

	logger.Info("server started", "addr", ":8080", "tls", false)
	logger.Error("payment failed", "order", 1042, "err", errors.New("card declined"))

	// Attributes from With are formatted once; a group's name qualifies the keys after it.
	requests := logger.With("service", "api").WithGroup("http")
	requests.Warn("slow request", "method", "GET", "path", "/search", "ms", 1204)
	// Output:
	// INFO  server started addr=:8080 tls=false
	// ERROR payment failed order=1042 err="card declined"
	// WARN  slow request service=api http.method=GET http.path=/search http.ms=1204
}

func ExampleOptions_level() {
	level := new(slog.LevelVar) // Info until it is set
	logger := slog.New(golog.NewHandler(os.Stdout, &golog.Options{Level: level, ReplaceAttr: withoutTime}))

	logger.Debug("not written")
	level.Set(slog.LevelDebug) // safe while other goroutines log
	logger.Debug("cache miss", "key", "user:42")
	// Output:
	// DEBUG cache miss key=user:42
}

func ExampleLevelCritical() {
	logger := slog.New(golog.NewHandler(os.Stdout, &golog.Options{ReplaceAttr: withoutTime}))

	logger.Log(context.Background(), golog.LevelCritical, "disk full", "path", "/var/lib/db", "free", 0)
	// Output:
	// CRITICAL disk full path=/var/lib/db free=0
}

func ExampleOptions_replaceAttr() {
	redact := func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == "password" {
			return slog.String(a.Key, "********")
		}
		return withoutTime(groups, a)
	}
	logger := slog.New(golog.NewHandler(os.Stdout, &golog.Options{ReplaceAttr: redact}))

	logger.Info("signed in", slog.Group("user", "name", "ada", "password", "hunter2"))
	// Output:
	// INFO  signed in user.name=ada user.password=********
}

// Command console logs a few records to the terminal, in colour when it is one. -debug logs Debug records too.
//
//	go run ./examples/console -debug
package main

import (
	"errors"
	"flag"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/vardius/golog/v2"
)

// newLogger logs to w, Debug records too if debug is set, and in colour if color is.
func newLogger(w io.Writer, debug, color bool) *slog.Logger {
	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}
	return slog.New(golog.NewHandler(w, &golog.Options{Level: level, Color: color}))
}

// isTerminal reports whether f is a character device, as a terminal is: colour codes in a file or a pipe are noise.
func isTerminal(f *os.File) bool {
	fi, err := f.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}

func serve(logger *slog.Logger) {
	logger.Debug("config loaded", "path", "config.toml")
	logger.Info("server started", "addr", ":8080")

	http := logger.WithGroup("http")
	http.Info("request", "method", "GET", "path", "/", "status", 200, "took", 3*time.Millisecond)
	http.Warn("slow request", "method", "GET", "path", "/search", "status", 200, "took", 1204*time.Millisecond)

	logger.Error("payment failed", "order", 1042, "err", errors.New("card declined"))
}

func main() {
	debug := flag.Bool("debug", false, "log Debug records too")
	flag.Parse()

	serve(newLogger(os.Stderr, *debug, isTerminal(os.Stderr)))
}

// Command file logs from one logger to two places: the terminal in colour, and a file in plain text with
// Debug records too.
//
//	go run ./examples/file app.log
package main

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"

	"github.com/vardius/golog/v2"
)

// openLog opens path for appending, and creates it if it isn't there.
func openLog(path string) (*os.File, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open log: %w", err)
	}
	return f, nil
}

// newLogger logs Info and above to console in colour, and everything to file in plain text.
func newLogger(console, file io.Writer) *slog.Logger {
	return slog.New(slog.NewMultiHandler(
		golog.NewHandler(console, &golog.Options{Color: true}),
		golog.NewHandler(file, &golog.Options{Level: slog.LevelDebug}),
	))
}

func run(path string, console io.Writer) error {
	f, err := openLog(path)
	if err != nil {
		return err
	}
	logger := newLogger(console, f)

	logger.Debug("cache warmed", "entries", 1200)
	logger.Info("server started", "addr", ":8080")

	return f.Close()
}

func main() {
	log.SetFlags(0)
	if len(os.Args) != 2 {
		log.Fatal("usage: file <path>")
	}
	if err := run(os.Args[1], os.Stderr); err != nil {
		log.Fatal(err)
	}
}

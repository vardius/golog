// Package golog is a log/slog handler that writes records as lines of text for people to read,
// optionally in colour, with no allocations per record.
//
//	logger := slog.New(golog.NewHandler(os.Stderr, &golog.Options{Color: true}))
//	logger.Info("server started", "addr", ":8080", "tls", false)
//
// prints
//
//	2026-10-24T09:41:07.312+11:00 INFO  server started addr=:8080 tls=false
//
// Use it with the standard *slog.Logger, or make it the default with slog.SetDefault.
//
// # The line
//
// Each record is one line: the time, the level, the file and line of the call if Options.AddSource is set,
// the message, and then the attributes as key=value. The time is RFC 3339 with milliseconds, in the record's
// time zone; a zero time is left out. Levels are DEBUG, INFO, WARN, ERROR and CRITICAL (LevelCritical),
// with an offset for levels in between, as in INFO+2, and the message is padded to line up after them.
//
// Attributes in a group are written with the group's name and a dot before the key, as in http.method=GET,
// whether the group comes from slog.Group or Logger.WithGroup. Keys and values are quoted as strconv.Quote
// quotes them when they are empty or hold a space of any kind, '=', '"', a control character, a character that
// doesn't print or invalid UTF-8. The message is written as it is, unless it holds a newline, another control
// character, a character that doesn't print, other than a space, or invalid UTF-8: then it is quoted too,
// so a record is always one line.
//
// Values are written by kind: numbers and booleans with strconv, durations as time.Duration.String does,
// and times like the record's. Other values are written with their AppendText or MarshalText method if they
// have one, an error as its Error(), a []byte quoted, and anything else with fmt's %+v.
//
// # Options
//
// Options.Level sets the minimum level, slog.LevelInfo if it is nil: pass a *slog.LevelVar to change it while
// the program runs. Options.ReplaceAttr rewrites or drops attributes, built-in ones too, with the contract of
// slog.HandlerOptions.ReplaceAttr. Options.Color adds ANSI colours for a terminal.
//
// # Concurrency
//
// A Handler is safe for concurrent use. Each record is formatted into a buffer of its own and written with one
// Write call under a lock that the handler shares with every handler derived from it, so lines never interleave.
// Handle returns the writer's error, wrapped.
//
// # Cost
//
// Handle allocates nothing once its buffers are warm: values are appended straight into a pooled buffer,
// attributes from WithAttrs are formatted once, the record's time is formatted in full once a second per buffer,
// and each call site's file and line are looked up once and kept. What can still allocate: fmt, for values with no
// faster way to be written; a long key that needs quoting inside a group; and, with ReplaceAttr, the built-in
// level and source handed to it. A *slog.Logger also allocates for a record with more than five attributes,
// whatever its handler.
package golog

# Changelog

## v2.0.0 (2026-10-24)

golog is now a handler for the standard library's `log/slog`: log with a `*slog.Logger`, and golog writes each
record as one line for people to read. See [MIGRATION.md](MIGRATION.md) for the before and after of every call.

### Added

- `NewHandler(w, opts)`, a `slog.Handler` that writes `time LEVEL message key=value` lines, with groups as
  `group.key`, values quoted only when they need it, and one `Write` call per record. It passes `testing/slogtest`.
- `Options`: `Level` (a `slog.Leveler`, so a `*slog.LevelVar` can change it while logging), `AddSource`,
  `ReplaceAttr` with the contract of `slog.HandlerOptions.ReplaceAttr`, and `Color` for terminals.
- `LevelCritical`, v1's Critical level, written as `CRITICAL`.
- Runnable examples in `examples/`, each with a test: a console logger, a terminal and a file at once, and a
  request ID taken from the context.

### Changed

- Module path is `github.com/vardius/golog/v2`; Go 1.26 or newer is required.
- Values are attributes, not a message formatted with `fmt.Sprintf`, and a record below the level is dropped
  before any of them is formatted.
- The default level is Info, as for slog's own handlers; v1 logged every level.
- The time is RFC 3339 with milliseconds in the record's time zone; v1 wrote microseconds in UTC.
- Colour is off unless `Options.Color` is set, and covers the level, time, source, keys and errors rather than the whole line.
- A message with a newline in it is quoted, so a record is always one line.
- No allocations per record, guarded by tests. Five attributes take 303 ns (310 ns in colour), against 447 ns for
  `slog.TextHandler` and 261 ns and 7 allocations for v1's console logger. v1 is quicker from one goroutine because a `*slog.Logger` reads the clock and
  looks up its caller on every record; from 16 goroutines at once, v2 takes 169 ns to v1's 204 ns. A record below the
  level costs 4 ns, where v1's callers paid 149 ns and 6 allocations to format it first.

### Fixed

- Changing the level while other goroutines log is safe: v1's `SetVerbosity` was a data race.
- A file that can't be opened is an error you handle: v1's `NewFileLogger` returned a nil `Logger` and dropped the error.
- `AddSource` reports the line that logged; with `log.Lshortfile`, v1 reported a line in its own `console.go`.
- Write errors are returned from `Handle`; v1 dropped them.

### Removed

- The `Logger` interface, `New`, `NewConsoleLogger`, `NewFileLogger`, `SetFlags` and `SetVerbosity`: use `*slog.Logger`
  and `NewHandler`.
- The `Verbose` level bits, `Disabled`, the prefix and colour constants, `DefaultFlags` and `DefaultVerbosity`.
- `Fatal`: log at `LevelCritical`, then call `os.Exit(1)`.
- The App Engine logger and the `google.golang.org/appengine` dependency. It didn't build in v1.2.0: `appengine.go`
  used `os` without importing it. Write JSON to standard output with `slog.NewJSONHandler` instead.

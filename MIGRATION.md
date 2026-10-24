# Migrating from v1 to v2

v2 is a handler for the standard library's `log/slog`, not a logger of its own. You log with a `*slog.Logger`
and golog writes the lines. The `Logger` interface, its level bits and its pre-formatted messages are gone:
values go in as attributes and come out as `key=value`. The changes to know about are listed below.

## 1. Update the import path

```sh
go get github.com/vardius/golog/v2
```

```go
import "github.com/vardius/golog/v2"
```

v2 needs Go 1.26 or newer, and depends on nothing outside the standard library.

## 2. Before and after

```go
// v1
logger := golog.New()
logger.SetVerbosity(golog.Info | golog.Warning | golog.Error | golog.Critical)

logger.Info(ctx, fmt.Sprintf("user %s signed in from %s", name, ip))
```

```go
// v2
logger := slog.New(golog.NewHandler(os.Stdout, &golog.Options{Color: true}))

logger.InfoContext(ctx, "user signed in", "user", name, "ip", ip)
```

The level is `slog.LevelInfo` by default, so the `SetVerbosity` line has no v2 counterpart here.
To make the logger the one `slog.Info` and friends use, call `slog.SetDefault(logger)`: that replaces a
package of your own that wrapped a `golog.Logger`.

## 3. Every call

| v1 | v2 |
| --- | --- |
| `golog.Logger` | `*slog.Logger` |
| `golog.New()`, `golog.NewConsoleLogger()` | `slog.New(golog.NewHandler(os.Stdout, &golog.Options{Color: true}))` |
| `golog.NewFileLogger(path)` | `os.OpenFile(path, os.O_CREATE\|os.O_WRONLY\|os.O_APPEND, 0o666)`, then `slog.New(golog.NewHandler(f, nil))`: see [files](#6-files) |
| `golog.NewAppEngineLogger()` | Removed: see [App Engine](#7-app-engine) |
| `logger.Debug(ctx, msg)` | `logger.DebugContext(ctx, msg, args...)` |
| `logger.Info(ctx, msg)` | `logger.InfoContext(ctx, msg, args...)` |
| `logger.Warning(ctx, msg)` | `logger.WarnContext(ctx, msg, args...)` |
| `logger.Error(ctx, msg)` | `logger.ErrorContext(ctx, msg, args...)` |
| `logger.Critical(ctx, msg)` | `logger.Log(ctx, golog.LevelCritical, msg, args...)` |
| `logger.Fatal(ctx, msg)` | `logger.Log(ctx, golog.LevelCritical, msg, args...)`, then `os.Exit(1)` |
| `logger.SetVerbosity(v)` | `Options.Level`; to change it later, a `*slog.LevelVar` and its `Set` method |
| `logger.SetFlags(flags)` | `Options.AddSource` for `log.Lshortfile` and `log.Llongfile`, `Options.ReplaceAttr` for the time |
| `golog.Debug`, `golog.Info`, `golog.Warning`, `golog.Error`, `golog.Critical` | `slog.LevelDebug`, `slog.LevelInfo`, `slog.LevelWarn`, `slog.LevelError`, `golog.LevelCritical` |
| `golog.Disabled` | `slog.DiscardHandler` in place of the handler |
| `golog.DefaultVerbosity` | `Options{Level: slog.LevelDebug}` |
| `golog.DefaultFlags` | The default: RFC 3339 with milliseconds, changed with `Options.ReplaceAttr` |
| `golog.Verbose` | `slog.Level`, or any `slog.Leveler` |
| `golog.DebugPrefix` … `golog.CriticalPrefix` | Removed: lines start with the time and the level's name |
| `golog.CLR_0` … `golog.CLR_W`, `golog.RESET` | Removed: `Options.Color` |

## 4. Levels

- **A minimum, not a set.** v1's verbosity was a set of bits, so `golog.Warning | golog.Error` logged warnings
  and errors but not critical records. slog's level is a minimum: `slog.LevelWarn` logs warnings and everything
  above. A set that isn't "this level and up" needs a handler of your own that wraps golog's and decides in `Enabled`.
- **Info by default.** v1 logged every level, Debug included, until you set the verbosity. v2 logs Info and up,
  as slog's own handlers do. Pass `Options{Level: slog.LevelDebug}` for v1's default.
- **Critical** is `golog.LevelCritical`, four above `slog.LevelError`. golog prints it as `CRITICAL`; slog's text
  and JSON handlers print it as `ERROR+4`. To have them print `CRITICAL` too:

  ```go
  ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
  	if a.Key == slog.LevelKey && groups == nil && a.Value.Any() == golog.LevelCritical {
  		return slog.String(a.Key, "CRITICAL")
  	}
  	return a
  },
  ```

- **No Fatal.** v2 doesn't exit for you: log at `golog.LevelCritical`, then call `os.Exit(1)`.
- **Changing the level while logging is safe.** v1's `SetVerbosity` wrote the level while other goroutines read it,
  a data race. A `*slog.LevelVar` is safe to `Set` at any time, and every handler derived from the logger follows it.

## 5. Messages and output

- **Attributes, not `fmt.Sprintf`.** Pass a fixed message and the values after it, as key and value pairs or
  `slog.Attr`s. They are written as `key=value` and quoted when they need it, and a record below the level is
  dropped before any of them is formatted. With v1, the `fmt.Sprintf` ran whether the record was logged or not.
- **The line.** v1 wrote `2026/10/23 22:41:07.312345 INFO: user ada signed in` in UTC.
  v2 writes `2026-10-24T09:41:07.312+11:00 INFO  user signed in user=ada`: RFC 3339 in the record's time zone,
  to the millisecond, with the level's name padded so messages line up. To log in UTC:

  ```go
  ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
  	if a.Key == slog.TimeKey && groups == nil {
  		return slog.Time(a.Key, a.Value.Time().UTC())
  	}
  	return a
  },
  ```

  Returning `slog.Attr{}` for the time leaves it out, as `SetFlags(0)` did.
- **Colour.** v1's console logger coloured the whole line, always. v2 colours only when `Options.Color` is set,
  and then only the level, with the time, source and keys dimmed and errors in red.
- **The writer is yours.** v1's console logger wrote to standard output. v2 writes to the `io.Writer` you pass:
  `os.Stdout` keeps v1's behaviour.
- **The source line is the caller's.** With `log.Lshortfile` set, v1 printed a line of its own `console.go`, where
  it called the `log` package. `Options.AddSource` prints the file and line that logged.
- **One line per record.** A message with a newline in it is quoted, as `"two\nlines"`, so it can't pass for two records.
- **Write errors.** v1 dropped them. v2's `Handle` returns them; `*slog.Logger` discards them, as it does for every handler.

## 6. Files

v1's `NewFileLogger` returned a nil `Logger` when it couldn't open the file, and dropped the error,
so the first record panicked. v2 doesn't open files: open the file yourself, handle the error, and pass it in.

```go
f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
if err != nil {
	return fmt.Errorf("open log: %w", err)
}
defer f.Close()

logger := slog.New(golog.NewHandler(f, nil))
```

These are v1's flags and permissions. To log to the terminal and a file at once, with a level for each,
combine two handlers with `slog.NewMultiHandler`, as the `examples/file` program does.

## 7. App Engine

v1's App Engine logger, chosen by the `appengine` and `appenginevm` build tags, wrote through
`google.golang.org/appengine/log`, the logging API of the App Engine standard environment. v2 drops it and the
dependency. It didn't build in v1.2.0 either: `appengine.go` called `os.Exit` without importing `os`.

v2 has no replacement. For a platform that collects structured logs from standard output, write JSON there with
the standard library, and rename fields with `ReplaceAttr` if the platform's documentation asks for other names:

```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
```

## 8. Context

v1 took a `context.Context` on every call and, except on App Engine, ignored it. In v2, `InfoContext` and the other
`...Context` methods hand it to the handler. golog's handler doesn't read it; to log values from it, such as a request ID,
wrap the handler, as the `examples/requestid` program does.

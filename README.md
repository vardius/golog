# 📟 golog

[![Test](https://github.com/vardius/golog/actions/workflows/test.yml/badge.svg)](https://github.com/vardius/golog/actions/workflows/test.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/vardius/golog/v2.svg)](https://pkg.go.dev/github.com/vardius/golog/v2)
[![Go Report Card](https://goreportcard.com/badge/github.com/vardius/golog/v2)](https://goreportcard.com/report/github.com/vardius/golog/v2)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE.md)

A handler for Go's `log/slog` that writes each record as one line people can read.

```txt
2026-10-24T09:41:07.312+11:00 INFO  server started addr=:8080 tls=false
2026-10-24T09:41:09.004+11:00 WARN  slow request http.method=GET http.path=/search http.took=1.204s
2026-10-24T09:41:12.871+11:00 ERROR payment failed order=1042 err="card declined"
```

- **Lines for people.** The time, a level that lines up, the message as it is, then `key=value`. Values are quoted only when they need to be, and a record is always one line.
- **Colour for terminals.** Levels in colour, the time and keys dimmed, errors in red. One option, off for files and pipes.
- **Standard slog.** A `slog.Handler`: log with `*slog.Logger`, set levels with `slog.LevelVar`, rewrite attributes with `ReplaceAttr`, add the caller with `AddSource`. It passes `testing/slogtest`.
- **No allocations per record.** Values go straight into a pooled buffer, `With` attributes are formatted once, and tests guard every allocation count.
- **One write per record,** under a lock shared by every handler made from it, so lines from many goroutines never mix.

📖 **[Documentation](https://vardius.github.io/golog/)** · [API reference](https://pkg.go.dev/github.com/vardius/golog/v2) · [Examples](examples) · [Migrating from v1](MIGRATION.md)

## Install

```sh
go get github.com/vardius/golog/v2
```

Requires Go 1.26 or newer. No dependencies beyond the standard library.

## Usage

```go
logger := slog.New(golog.NewHandler(os.Stderr, &golog.Options{Color: true}))
slog.SetDefault(logger) // optional: slog.Info and friends use it too

logger.Info("server started", "addr", ":8080", "tls", false)

http := logger.With("service", "api").WithGroup("http")
http.Warn("slow request", "method", "GET", "path", "/search", "took", took)
```

```txt
2026-10-24T09:41:07.312+11:00 INFO  server started addr=:8080 tls=false
2026-10-24T09:41:09.004+11:00 WARN  slow request service=api http.method=GET http.path=/search http.took=1.204s
```

## Options

```go
golog.NewHandler(w, &golog.Options{
	Level:       slog.LevelDebug, // or a *slog.LevelVar, to change it while running; nil is Info
	AddSource:   true,            // the file and line of the call, before the message
	ReplaceAttr: redact,          // rewrite or drop attributes, as for slog's own handlers
	Color:       true,            // ANSI colours, for a terminal
})
```

`golog.LevelCritical` is v1's Critical level, four above `slog.LevelError`, and is written as `CRITICAL`:

```go
logger.Log(ctx, golog.LevelCritical, "disk full", "path", "/var/lib/db")
```

## Output

- **Time:** RFC 3339 with milliseconds, in the record's time zone. A zero time is left out.
- **Level:** `DEBUG`, `INFO`, `WARN`, `ERROR`, `CRITICAL`, or an offset such as `INFO+2`, padded so messages line up.
- **Message:** as it is, unless it has a newline, another control character, a character that doesn't print (other than a space) or invalid UTF-8 in it: then it is quoted.
- **Attributes:** `key=value`; a group's name and a dot before its keys, as in `http.method=GET`. A key or value is quoted as `strconv.Quote` quotes it when it is empty or has a space, `=`, `"`, a control character, a character that doesn't print or invalid UTF-8 in it.

Each record is one call to the writer's `Write`. To log to a file, open it and pass it in; to log to the terminal and a file at once, combine two handlers with `slog.NewMultiHandler`, as [examples/file](examples/file/main.go) does.

## Benchmarks

Time and allocations per record, on an Apple M4 Max with Go 1.27.0, median of 5 runs. All ran on the same harness,
through a `*slog.Logger` with `LogAttrs`, writing to a writer that drops its input. v1 takes a string, so its benchmark
formats the same pairs into the message with `fmt.Sprintf`, as its callers had to; its console logger always colours,
so compare it with the last column. With 16 goroutines at once, the time is wall time per record across all of them,
so it measures throughput. Run golog's own suite with `go test -bench .`.

| Scenario | v1 | slog.TextHandler | slog.JSONHandler | v2 | v2 in colour |
| --- | ---: | ---: | ---: | ---: | ---: |
| Message only | 113 ns · 1 alloc | 259 ns · 0 allocs | 264 ns · 0 allocs | **165 ns · 0 allocs** | **174 ns · 0 allocs** |
| 5 attributes | 261 ns · 7 allocs | 447 ns · 0 allocs | 445 ns · 0 allocs | **303 ns · 0 allocs** | **310 ns · 0 allocs** |
| 10 attributes | 490 ns · 10 allocs | 837 ns · 2 allocs | 777 ns · 3 allocs | **509 ns · 1 alloc** | **522 ns · 1 alloc** |
| Groups: `With`, `WithGroup`, a group of 5 | 315 ns · 7 allocs | 531 ns · 0 allocs | 483 ns · 0 allocs | **334 ns · 0 allocs** | **344 ns · 0 allocs** |
| 5 attributes, 16 goroutines at once | 204 ns · 7 allocs | 206 ns · 0 allocs | 209 ns · 0 allocs | **169 ns · 0 allocs** | **170 ns · 0 allocs** |
| Below the level | 149 ns · 6 allocs | 3.2 ns · 0 allocs | 3.2 ns · 0 allocs | **3.6 ns · 0 allocs** | **3.6 ns · 0 allocs** |

A `*slog.Logger` reads the clock and looks up its caller on every record, whichever its handler: about 130 ns here.
v1 only reads the clock, which is why it is quicker for a short record from one goroutine. What golog adds on top
allocates nothing and costs less than what slog's own handlers add. Ten attributes cost one allocation, the `slog.Record`'s,
which holds five inline.

## Upgrading from v1

v2 is a handler for `log/slog` instead of a logger of its own:

```go
// v1
logger := golog.New()
logger.Info(ctx, fmt.Sprintf("user %s signed in from %s", name, ip))

// v2
logger := slog.New(golog.NewHandler(os.Stdout, &golog.Options{Color: true}))
logger.InfoContext(ctx, "user signed in", "user", name, "ip", ip)
```

[MIGRATION.md](MIGRATION.md) maps every v1 call to its v2 form and lists the behaviour changes.

## Contributing

Bug reports and pull requests are welcome: see [CONTRIBUTING.md](.github/CONTRIBUTING.md).

## License

MIT, see [LICENSE.md](LICENSE.md).

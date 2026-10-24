package golog_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/netip"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"testing/slogtest"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/vardius/golog/v2"
)

var _ slog.Handler = (*golog.Handler)(nil)

// at is the time of every record in these tests. Its zone is fixed, so the tests pass in any local time zone.
var at = time.Date(2026, 10, 24, 9, 41, 7, 312_987_654, time.FixedZone("AEDT", 11*60*60))

// line formats one record at time at through h and returns what h wrote.
func line(t *testing.T, opts *golog.Options, level slog.Level, msg string, attrs ...slog.Attr) string {
	t.Helper()
	var buf bytes.Buffer
	r := slog.NewRecord(at, level, msg, 0)
	r.AddAttrs(attrs...)
	if err := golog.NewHandler(&buf, opts).Handle(t.Context(), r); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	return buf.String()
}

// noTime drops the time, so a test can compare a whole line logged through a *slog.Logger.
func noTime(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.TimeKey && groups == nil {
		return slog.Attr{}
	}
	return a
}

func TestHandlerLine(t *testing.T) {
	tests := []struct {
		name  string
		level slog.Level
		msg   string
		attrs []slog.Attr
		want  string
	}{
		{"message only", slog.LevelInfo, "server started", nil, "2026-10-24T09:41:07.312+11:00 INFO  server started\n"},
		{"attrs", slog.LevelInfo, "server started", []slog.Attr{slog.String("addr", ":8080"), slog.Bool("tls", false)},
			"2026-10-24T09:41:07.312+11:00 INFO  server started addr=:8080 tls=false\n"},
		{"empty message", slog.LevelInfo, "", []slog.Attr{slog.Int("n", 1)}, "2026-10-24T09:41:07.312+11:00 INFO  n=1\n"},
		{"empty message, no attrs", slog.LevelWarn, "", nil, "2026-10-24T09:41:07.312+11:00 WARN\n"},
		{"message with spaces, = and quotes is written as it is", slog.LevelInfo, `a=b "c" d`, nil,
			"2026-10-24T09:41:07.312+11:00 INFO  a=b \"c\" d\n"},
		{"unicode message", slog.LevelInfo, "zażółć gęślą jaźń ✓", nil, "2026-10-24T09:41:07.312+11:00 INFO  zażółć gęślą jaźń ✓\n"},
		{"message with a newline is quoted", slog.LevelInfo, "two\nlines", nil, "2026-10-24T09:41:07.312+11:00 INFO  \"two\\nlines\"\n"},
		{"message with a tab is quoted", slog.LevelInfo, "a\tb", nil, "2026-10-24T09:41:07.312+11:00 INFO  \"a\\tb\"\n"},
		{"message with an escape code is quoted", slog.LevelInfo, "\x1b[2Jgone", nil, "2026-10-24T09:41:07.312+11:00 INFO  \"\\x1b[2Jgone\"\n"},
		{"message with invalid UTF-8 is quoted", slog.LevelInfo, "bad \xff", nil, "2026-10-24T09:41:07.312+11:00 INFO  \"bad \\xff\"\n"},
		{"message with a line separator is quoted", slog.LevelInfo, "a\u2028b", nil, "2026-10-24T09:41:07.312+11:00 INFO  \"a\\u2028b\"\n"},
		{"message with DEL is quoted", slog.LevelInfo, "a\x7fb", nil, "2026-10-24T09:41:07.312+11:00 INFO  \"a\\x7fb\"\n"},
		{"message with a bidi control is quoted", slog.LevelInfo, "a\u202eb", nil, "2026-10-24T09:41:07.312+11:00 INFO  \"a\\u202eb\"\n"},
		{"message with other spaces is written as it is", slog.LevelInfo, "Erreur\u00a0: サーバー\u3000起動", nil,
			"2026-10-24T09:41:07.312+11:00 INFO  Erreur\u00a0: サーバー\u3000起動\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := line(t, nil, tt.level, tt.msg, tt.attrs...); got != tt.want {
				t.Errorf("got  %q\nwant %q", got, tt.want)
			}
		})
	}
}

func TestHandlerLevels(t *testing.T) {
	tests := []struct {
		level slog.Level
		want  string // what follows the time
	}{
		{slog.LevelDebug - 2, "DEBUG-2 m"},
		{slog.LevelDebug, "DEBUG m"},
		{slog.LevelDebug + 1, "DEBUG+1 m"},
		{slog.LevelInfo, "INFO  m"},
		{slog.LevelInfo + 2, "INFO+2 m"},
		{slog.LevelWarn, "WARN  m"},
		{slog.LevelError, "ERROR m"},
		{slog.LevelError + 3, "ERROR+3 m"},
		{golog.LevelCritical, "CRITICAL m"},
		{golog.LevelCritical + 1, "CRITICAL+1 m"},
	}
	for _, tt := range tests {
		t.Run(tt.level.String(), func(t *testing.T) {
			got := line(t, &golog.Options{Level: slog.Level(-100)}, tt.level, "m")
			if want := "2026-10-24T09:41:07.312+11:00 " + tt.want + "\n"; got != want {
				t.Errorf("got  %q\nwant %q", got, want)
			}
		})
	}
}

type point struct{ X, Y int }

type textOnly string // has MarshalText but not AppendText

func (s textOnly) MarshalText() ([]byte, error) { return []byte("text:" + s), nil }

type badText struct{}

func (badText) AppendText([]byte) ([]byte, error) { return nil, errors.New("cannot append") }

type badMarshal struct{}

func (badMarshal) MarshalText() ([]byte, error) { return nil, errors.New("cannot marshal") }

type nilError struct{ msg string }

func (e *nilError) Error() string { return e.msg }

type panicky struct{}

func (panicky) String() string { panic("boom") }

type token string

func (token) LogValue() slog.Value { return slog.StringValue("REDACTED") }

type user struct{ id int }

func (u user) LogValue() slog.Value {
	return slog.GroupValue(slog.Int("id", u.id), slog.String("role", "admin"))
}

func TestHandlerValues(t *testing.T) {
	var typedNil *nilError
	tests := []struct {
		name string
		attr slog.Attr
		want string
	}{
		{"string", slog.String("k", "plain"), "k=plain"},
		{"string with a space", slog.String("k", "two words"), `k="two words"`},
		{"string with =", slog.String("k", "a=b"), `k="a=b"`},
		{"string with a quote", slog.String("k", `say "hi"`), `k="say \"hi\""`},
		{"string with a backslash", slog.String("k", `C:\dir`), `k=C:\dir`},
		{"empty string", slog.String("k", ""), `k=""`},
		{"string with a newline", slog.String("k", "a\nb"), `k="a\nb"`},
		{"string with DEL", slog.String("k", "a\x7fb"), `k="a\x7fb"`},
		{"unicode string", slog.String("k", "żółw"), "k=żółw"},
		{"string with a unicode space", slog.String("k", "a\u00a0b"), `k="a\u00a0b"`},
		{"string with invalid UTF-8", slog.String("k", "\xff"), `k="\xff"`},
		{"key with a space", slog.Int("my key", 1), `"my key"=1`},
		{"empty key", slog.Int("", 1), `""=1`},
		{"int", slog.Int("k", -42), "k=-42"},
		{"uint", slog.Uint64("k", math.MaxUint64), "k=18446744073709551615"},
		{"float", slog.Float64("k", 1.5), "k=1.5"},
		{"big float", slog.Float64("k", 1e21), "k=1e+21"},
		{"NaN", slog.Float64("k", math.NaN()), "k=NaN"},
		{"infinity", slog.Float64("k", math.Inf(-1)), "k=-Inf"},
		{"bool", slog.Bool("k", true), "k=true"},
		{"duration", slog.Duration("k", 1500*time.Millisecond), "k=1.5s"},
		{"long duration", slog.Duration("k", -(2540400*time.Hour + 10*time.Minute + 10*time.Second + time.Nanosecond)), "k=-2540400h10m10.000000001s"},
		{"zero duration", slog.Duration("k", 0), "k=0s"},
		{"time", slog.Time("k", time.Date(2026, 1, 2, 3, 4, 5, 999_999_999, time.UTC)), "k=2026-01-02T03:04:05.999Z"},
		{"time on a whole second", slog.Time("k", time.Date(2026, 1, 2, 3, 4, 5, 0, time.FixedZone("", -(3*60+30)*60))), "k=2026-01-02T03:04:05.000-03:30"},
		{"time past year 9999", slog.Time("k", time.Date(12026, 1, 2, 3, 4, 5, 6_000_000, time.UTC)), "k=12026-01-02T03:04:05.006Z"},
		{"error", slog.Any("err", errors.New("card declined")), `err="card declined"`},
		{"wrapped error", slog.Any("err", fmt.Errorf("pay: %w", errors.New("declined"))), `err="pay: declined"`},
		{"nil", slog.Any("err", nil), "err=<nil>"},
		{"nil pointer error", slog.Any("err", typedNil), "err=<nil>"},
		{"panicking String method", slog.Any("k", panicky{}), `k="%!v(PANIC=String method: boom)"`},
		{"AppendText", slog.Any("ip", netip.MustParseAddr("192.0.2.1")), "ip=192.0.2.1"},
		{"AppendText that needs quoting", slog.Any("ip", netip.Addr{}), `ip=""`},
		{"AppendText error", slog.Any("k", badText{}), `k="!ERROR:cannot append"`},
		{"MarshalText", slog.Any("k", textOnly("a b")), `k="text:a b"`},
		{"MarshalText error", slog.Any("k", badMarshal{}), `k="!ERROR:cannot marshal"`},
		{"bytes", slog.Any("k", []byte("hi")), `k="hi"`},
		{"struct", slog.Any("k", point{1, 2}), `k="{X:1 Y:2}"`},
		{"slice", slog.Any("k", []int{1, 2}), `k="[1 2]"`},
		{"LogValuer", slog.Any("token", token("secret")), "token=REDACTED"},
		{"LogValuer to a group", slog.Any("user", user{7}), "user.id=7 user.role=admin"},
		{"group", slog.Group("req", slog.String("method", "GET"), slog.Group("url", slog.String("path", "/"))),
			"req.method=GET req.url.path=/"},
		{"group with no key is inlined", slog.Group("", slog.Int("a", 1), slog.Int("b", 2)), "a=1 b=2"},
		{"group with a key that needs quoting", slog.Group("my group", slog.Int("a", 1)), `"my group.a"=1`},
		{"key that needs quoting in a group", slog.Group("g", slog.Int("a b", 1)), `"g.a b"=1`},
		{"group with a unicode name", slog.Group("żółw", slog.Int("a", 1)), "żółw.a=1"},
		{"group with a unicode space in its name", slog.Group("a\u00a0b", slog.Int("c", 1)), `"a\u00a0b.c"=1`},
		{"group with invalid UTF-8 in its name", slog.Group("\xff", slog.Int("c", 1)), `"\xff.c"=1`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := line(t, nil, slog.LevelInfo, "m", tt.attr)
			if want := "2026-10-24T09:41:07.312+11:00 INFO  m " + tt.want + "\n"; got != want {
				t.Errorf("got  %q\nwant %q", got, want)
			}
		})
	}
}

func TestHandlerDropsEmpty(t *testing.T) {
	tests := []struct {
		name string
		attr slog.Attr
	}{
		{"empty attr", slog.Attr{}},
		{"empty group", slog.Group("g")},
		{"group of empty attrs", slog.Group("g", slog.Attr{}, slog.Group("h"))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := line(t, nil, slog.LevelInfo, "m", slog.Int("a", 1), tt.attr, slog.Int("b", 2))
			if want := "2026-10-24T09:41:07.312+11:00 INFO  m a=1 b=2\n"; got != want {
				t.Errorf("got  %q\nwant %q", got, want)
			}
		})
	}
}

// Records reuse the time formatted for the record before them when it is in the same second and zone.
func TestHandlerRecordTimes(t *testing.T) {
	sydney := time.FixedZone("AEDT", 11*60*60)
	perth := time.FixedZone("AWST", 8*60*60)
	times := []time.Time{
		time.Date(2026, 10, 24, 9, 41, 7, 5_000_000, sydney),
		time.Date(2026, 10, 24, 9, 41, 7, 999_999_999, sydney),
		time.Date(2026, 10, 24, 6, 41, 7, 120_000_000, perth),  // the same second, in another zone
		time.Date(2026, 10, 24, 9, 41, 8, 0, sydney),           // the next second
		time.Date(2026, 10, 24, 9, 41, 7, 450_000_000, sydney), // and back
		time.Date(2027, 10, 24, 9, 41, 7, 450_000_000, sydney), // a year on, the same clock
	}
	var buf bytes.Buffer
	h := golog.NewHandler(&buf, nil)
	for _, tm := range times {
		if err := h.Handle(t.Context(), slog.NewRecord(tm, slog.LevelInfo, "m", 0)); err != nil {
			t.Fatal(err)
		}
	}
	want := `2026-10-24T09:41:07.005+11:00 INFO  m
2026-10-24T09:41:07.999+11:00 INFO  m
2026-10-24T06:41:07.120+08:00 INFO  m
2026-10-24T09:41:08.000+11:00 INFO  m
2026-10-24T09:41:07.450+11:00 INFO  m
2027-10-24T09:41:07.450+11:00 INFO  m
`
	if got := buf.String(); got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
}

func TestHandlerZeroTime(t *testing.T) {
	var buf bytes.Buffer
	r := slog.NewRecord(time.Time{}, slog.LevelInfo, "m", 0)
	if err := golog.NewHandler(&buf, nil).Handle(t.Context(), r); err != nil {
		t.Fatal(err)
	}
	if got, want := buf.String(), "INFO  m\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestHandlerWithAttrsAndGroups(t *testing.T) {
	var buf bytes.Buffer
	h := golog.NewHandler(&buf, &golog.Options{ReplaceAttr: noTime})
	base := slog.New(h).With("svc", "api")
	req := base.WithGroup("http").With("method", "GET")

	// Siblings derived from the same logger must not share their attributes.
	a := req.With("path", "/a")
	b := req.With("path", "/b")
	a.Info("a", "status", 200)
	b.Info("b", "status", 404)
	req.WithGroup("tls").Info("handshake", "version", "1.3")
	req.Info("no record attrs")
	base.WithGroup("empty").Info("group with no attrs")
	base.WithGroup("").With().Info("no group, no attrs")
	base.With("token", token("secret")).Info("resolved once")

	want := `INFO  a svc=api http.method=GET http.path=/a http.status=200
INFO  b svc=api http.method=GET http.path=/b http.status=404
INFO  handshake svc=api http.method=GET http.tls.version=1.3
INFO  no record attrs svc=api http.method=GET
INFO  group with no attrs svc=api
INFO  no group, no attrs svc=api
INFO  resolved once svc=api token=REDACTED
`
	if got := buf.String(); got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
}

func TestHandlerSameHandler(t *testing.T) {
	h := golog.NewHandler(&bytes.Buffer{}, nil)
	if h.WithGroup("") != slog.Handler(h) {
		t.Error("WithGroup(\"\") returned a new handler, want the receiver")
	}
	if h.WithAttrs(nil) != slog.Handler(h) {
		t.Error("WithAttrs(nil) returned a new handler, want the receiver")
	}
}

func TestHandlerEnabled(t *testing.T) {
	tests := []struct {
		name  string
		level slog.Leveler
		want  map[slog.Level]bool
	}{
		{"nil is Info", nil, map[slog.Level]bool{slog.LevelDebug: false, slog.LevelInfo: true, slog.LevelError: true}},
		{"Debug", slog.LevelDebug, map[slog.Level]bool{slog.LevelDebug - 1: false, slog.LevelDebug: true}},
		{"Critical", golog.LevelCritical, map[slog.Level]bool{slog.LevelError: false, golog.LevelCritical: true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := golog.NewHandler(&bytes.Buffer{}, &golog.Options{Level: tt.level})
			for l, want := range tt.want {
				if got := h.Enabled(t.Context(), l); got != want {
					t.Errorf("Enabled(%v) = %v, want %v", l, got, want)
				}
			}
		})
	}
}

// v1's SetVerbosity wrote the level without synchronisation while other goroutines read it to log.
// A *slog.LevelVar changes the level atomically, and the handler reads it on every record.
func TestHandlerLevelVarChangesWhileLogging(t *testing.T) {
	var level slog.LevelVar
	var buf bytes.Buffer
	logger := slog.New(golog.NewHandler(&buf, &golog.Options{Level: &level, ReplaceAttr: noTime}))

	var wg sync.WaitGroup
	for range 4 {
		wg.Go(func() {
			for range 100 {
				logger.Debug("d")
			}
		})
	}
	wg.Go(func() {
		for i := range 100 {
			level.Set(slog.Level(i%2) * slog.LevelDebug)
		}
	})
	wg.Wait()

	buf.Reset()
	level.Set(slog.LevelWarn)
	logger.Info("dropped")
	logger.Warn("kept")
	level.Set(slog.LevelDebug)
	logger.Debug("kept too")
	logger.With("a", 1).WithGroup("g").Debug("derived handlers follow it")
	if got, want := buf.String(), "WARN  kept\nDEBUG kept too\nDEBUG derived handlers follow it a=1\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// writes records each Write call, and is not safe for concurrent use: the race detector catches two at once.
type writes struct{ calls []string }

func (w *writes) Write(p []byte) (int, error) {
	w.calls = append(w.calls, string(p))
	return len(p), nil
}

func TestHandlerOneWritePerRecord(t *testing.T) {
	w := &writes{}
	root := slog.New(golog.NewHandler(w, &golog.Options{ReplaceAttr: noTime}))
	loggers := []*slog.Logger{root, root.With("a", 1), root.WithGroup("g").With("b", 2), root.WithGroup("g")}

	var wg sync.WaitGroup
	for i, l := range loggers {
		wg.Go(func() {
			for j := range 50 {
				l.Info("record", "logger", i, "n", j, "pad", strings.Repeat("x", 1+100*j))
			}
		})
	}
	wg.Wait()

	if len(w.calls) != 4*50 {
		t.Fatalf("got %d writes, want %d", len(w.calls), 4*50)
	}
	whole := regexp.MustCompile(`^INFO  record (a=1 |g\.b=2 )?(g\.)?logger=\d (g\.)?n=\d+ (g\.)?pad=x*\n$`)
	for _, c := range w.calls {
		if !whole.MatchString(c) {
			t.Fatalf("write %q is not one whole record", c)
		}
	}
}

func TestHandlerLargeRecord(t *testing.T) {
	big := strings.Repeat("y", 64<<10)
	for range 3 { // past the pooled buffer's cap, and again after it was dropped
		got := line(t, nil, slog.LevelInfo, "m", slog.String("big", big))
		if want := "2026-10-24T09:41:07.312+11:00 INFO  m big=" + big + "\n"; got != want {
			t.Fatalf("got %d bytes, want %d", len(got), len(want))
		}
	}
	if got, want := line(t, nil, slog.LevelInfo, "m"), "2026-10-24T09:41:07.312+11:00 INFO  m\n"; got != want {
		t.Errorf("after a large record got %q, want %q", got, want)
	}
}

var errDiskFull = errors.New("disk full")

type failing struct{}

func (failing) Write([]byte) (int, error) { return 0, errDiskFull }

func TestHandlerWriteError(t *testing.T) {
	err := golog.NewHandler(failing{}, nil).Handle(t.Context(), slog.NewRecord(at, slog.LevelInfo, "m", 0))
	if !errors.Is(err, errDiskFull) {
		t.Errorf("Handle() error = %v, want %v", err, errDiskFull)
	}
}

func TestHandlerColor(t *testing.T) {
	got := line(t, &golog.Options{Color: true, Level: slog.LevelDebug}, slog.LevelError, "payment failed",
		slog.Int("order", 1042), slog.Any("err", errors.New("card declined")))
	want := "\x1b[2m2026-10-24T09:41:07.312+11:00\x1b[0m \x1b[1;31mERROR\x1b[0m payment failed " +
		"\x1b[2morder=\x1b[0m1042 \x1b[2merr=\x1b[0m\x1b[31m\"card declined\"\x1b[0m\n"
	if got != want {
		t.Errorf("got  %q\nwant %q", got, want)
	}

	levels := []struct {
		level slog.Level
		want  string
	}{
		{slog.LevelDebug, "\x1b[1;35mDEBUG\x1b[0m"},
		{slog.LevelInfo, "\x1b[1;36mINFO\x1b[0m "},
		{slog.LevelWarn, "\x1b[1;33mWARN\x1b[0m "},
		{slog.LevelError, "\x1b[1;31mERROR\x1b[0m"},
		{golog.LevelCritical, "\x1b[1;97;41mCRITICAL\x1b[0m"},
	}
	for _, tt := range levels {
		got := line(t, &golog.Options{Color: true, Level: slog.LevelDebug}, tt.level, "m")
		if want := "\x1b[2m2026-10-24T09:41:07.312+11:00\x1b[0m " + tt.want + " m\n"; got != want {
			t.Errorf("%v: got %q, want %q", tt.level, got, want)
		}
	}
}

// sources keeps the source slog itself reports for each record, to check the handler's against.
type sources struct{ files *[]string }

func (sources) Enabled(context.Context, slog.Level) bool { return true }
func (s sources) Handle(_ context.Context, r slog.Record) error {
	src := r.Source()
	*s.files = append(*s.files, fmt.Sprintf("%s:%d", src.File, src.Line))
	return nil
}
func (s sources) WithAttrs([]slog.Attr) slog.Handler { return s }
func (s sources) WithGroup(string) slog.Handler      { return s }

func TestHandlerSource(t *testing.T) {
	// With ReplaceAttr, the source comes from slog.Record.Source; without it, from golog's own lookup.
	for _, replace := range []func([]string, slog.Attr) slog.Attr{noTime, nil} {
		var buf bytes.Buffer
		var want []string
		logger := slog.New(slog.NewMultiHandler(
			golog.NewHandler(&buf, &golog.Options{AddSource: true, ReplaceAttr: replace}),
			sources{&want},
		))

		for range 2 { // the second time, from the cache
			logger.Info("direct")
			logFromHelper(logger)
			func() { logger.Info("closure") }()
		}

		got := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")
		if len(got) != len(want) {
			t.Fatalf("got %d lines, want %d", len(got), len(want))
		}
		for i, msg := range []string{"direct", "from helper", "closure", "direct", "from helper", "closure"} {
			line := got[i]
			if replace == nil {
				line = line[strings.IndexByte(line, ' ')+1:] // after the time
			}
			if w := "INFO  " + want[i] + " " + msg; line != w {
				t.Errorf("got  %q\nwant %q", line, w)
			}
			if !strings.Contains(want[i], "handler_test.go:") {
				t.Errorf("slog reports %q for %q, want a line in handler_test.go", want[i], msg)
			}
		}
	}
}

// logFromHelper is small enough to be inlined into its caller: the source is still this line.
func logFromHelper(l *slog.Logger) { l.Info("from helper") }

// Sibling WithGroup calls from a handler three groups deep must not see each other's group in ReplaceAttr.
func TestHandlerSiblingGroups(t *testing.T) {
	var seen []string
	rep := func(groups []string, a slog.Attr) slog.Attr {
		if groups != nil {
			seen = append(seen, strings.Join(groups, ".")+"."+a.Key)
		}
		return a
	}
	l := slog.New(golog.NewHandler(io.Discard, &golog.Options{ReplaceAttr: rep})).WithGroup("a").WithGroup("b").WithGroup("c")
	d, e := l.WithGroup("d"), l.WithGroup("e")
	d.Info("m", "k", 1)
	e.Info("m", "k", 2)
	if want := "a.b.c.d.k a.b.c.e.k"; strings.Join(seen, " ") != want {
		t.Errorf("ReplaceAttr saw %q, want %q", seen, want)
	}
}

type resolvesTo string

func (r resolvesTo) LogValue() slog.Value { return slog.StringValue(string(r)) }

func TestHandlerBuiltinReplaced(t *testing.T) {
	tests := []struct {
		name    string
		replace func(a slog.Attr) slog.Attr
		want    string
	}{
		{"time as a group", func(a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Group("time", "x", 1, "y", 2)
			}
			return a
		}, "time.x=1 time.y=2 INFO  m\n"},
		{"time as an empty group", func(a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Group("time")
			}
			return a
		}, "INFO  m\n"},
		{"message as a LogValuer", func(a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			if a.Key == slog.MessageKey {
				return slog.Any(slog.MessageKey, resolvesTo("resolved"))
			}
			return a
		}, "INFO  resolved\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			rep := func(groups []string, a slog.Attr) slog.Attr {
				if groups != nil {
					return a
				}
				return tt.replace(a)
			}
			slog.New(golog.NewHandler(&buf, &golog.Options{ReplaceAttr: rep})).Info("m")
			if got := buf.String(); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestHandlerSourceReplaced(t *testing.T) {
	tests := []struct {
		name    string
		replace func(src *slog.Source) slog.Attr
		want    string // with %d for the line
	}{
		{"shortened", func(src *slog.Source) slog.Attr {
			src.File = filepath.Base(src.File)
			return slog.Any(slog.SourceKey, src)
		}, "INFO  handler_test.go:%d here\n"},
		{"to another kind", func(src *slog.Source) slog.Attr {
			return slog.Int(slog.SourceKey, src.Line)
		}, "INFO  %d here\n"},
		{"dropped", func(*slog.Source) slog.Attr { return slog.Attr{} }, "INFO  here\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			replace := func(groups []string, a slog.Attr) slog.Attr {
				if src, ok := a.Value.Any().(*slog.Source); ok && a.Key == slog.SourceKey && groups == nil {
					return tt.replace(src)
				}
				return noTime(groups, a)
			}
			logger := slog.New(golog.NewHandler(&buf, &golog.Options{AddSource: true, ReplaceAttr: replace}))
			_, _, line, _ := runtime.Caller(0)
			func() { logger.Info("here") }() // the line after runtime.Caller

			want := tt.want
			if strings.Contains(want, "%d") {
				want = fmt.Sprintf(want, line+1)
			}
			if got := buf.String(); got != want {
				t.Errorf("got %q, want %q", got, want)
			}
		})
	}
}

func TestHandlerSourceUnknown(t *testing.T) {
	keep := func(_ []string, a slog.Attr) slog.Attr { return a }
	for _, replace := range []func([]string, slog.Attr) slog.Attr{nil, keep} {
		for _, pc := range []uintptr{0, 1} { // no PC, and one no function holds
			var buf bytes.Buffer
			r := slog.NewRecord(at, slog.LevelInfo, "m", pc)
			h := golog.NewHandler(&buf, &golog.Options{AddSource: true, ReplaceAttr: replace})
			if err := h.Handle(t.Context(), r); err != nil {
				t.Fatal(err)
			}
			if got, want := buf.String(), "2026-10-24T09:41:07.312+11:00 INFO  m\n"; got != want {
				t.Errorf("PC %d, ReplaceAttr %v: got %q, want %q", pc, replace != nil, got, want)
			}
		}
	}
}

func TestHandlerReplaceAttr(t *testing.T) {
	tests := []struct {
		name    string
		replace func(groups []string, a slog.Attr) slog.Attr
		want    string
	}{
		{"keeps everything", func(_ []string, a slog.Attr) slog.Attr { return a },
			"2026-10-24T09:41:07.312+11:00 WARN  m a=1 g.b=2\n"},
		{"drops built-ins", func(groups []string, a slog.Attr) slog.Attr {
			if groups == nil && (a.Key == slog.TimeKey || a.Key == slog.LevelKey || a.Key == slog.MessageKey) {
				return slog.Attr{}
			}
			return a
		}, "a=1 g.b=2\n"},
		{"rewrites built-ins", func(groups []string, a slog.Attr) slog.Attr {
			switch {
			case groups != nil:
				return a
			case a.Key == slog.TimeKey:
				return slog.Time(a.Key, a.Value.Time().UTC())
			case a.Key == slog.LevelKey:
				return slog.String(a.Key, "warning")
			case a.Key == slog.MessageKey:
				return slog.Int(a.Key, 42)
			}
			return a
		}, "2026-10-23T22:41:07.312Z warning 42 a=1 g.b=2\n"},
		{"rewrites the time to text", func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey && groups == nil {
				return slog.String(a.Key, a.Value.Time().Format(time.Kitchen))
			}
			return a
		}, "9:41AM WARN  m a=1 g.b=2\n"},
		{"maps the level", func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.LevelKey && groups == nil {
				return slog.Any(a.Key, golog.LevelCritical)
			}
			return a
		}, "2026-10-24T09:41:07.312+11:00 CRITICAL m a=1 g.b=2\n"},
		{"rewrites the message", func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.MessageKey && groups == nil {
				return slog.String(a.Key, "line\nbreak")
			}
			return a
		}, "2026-10-24T09:41:07.312+11:00 WARN  \"line\\nbreak\" a=1 g.b=2\n"},
		{"drops an attr in a group", func(groups []string, a slog.Attr) slog.Attr {
			if len(groups) == 1 && groups[0] == "g" && a.Key == "b" {
				return slog.Attr{}
			}
			return a
		}, "2026-10-24T09:41:07.312+11:00 WARN  m a=1\n"},
		{"returns a group", func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == "a" {
				return slog.Group("x", slog.Int("y", 1))
			}
			return a
		}, "2026-10-24T09:41:07.312+11:00 WARN  m x.y=1 g.b=2\n"},
		{"returns a LogValuer", func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == "a" {
				return slog.Any("a", token("secret"))
			}
			return a
		}, "2026-10-24T09:41:07.312+11:00 WARN  m a=REDACTED g.b=2\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := line(t, &golog.Options{ReplaceAttr: tt.replace}, slog.LevelWarn, "m",
				slog.Int("a", 1), slog.Group("g", slog.Int("b", 2)))
			if got != tt.want {
				t.Errorf("got  %q\nwant %q", got, tt.want)
			}
		})
	}
}

// ReplaceAttr sees the groups an attribute is in, from WithGroup and from slog.Group alike,
// and is never called for a group itself.
func TestHandlerReplaceAttrGroups(t *testing.T) {
	var seen []string
	record := func(groups []string, a slog.Attr) slog.Attr {
		if a.Value.Kind() == slog.KindGroup {
			t.Errorf("ReplaceAttr called for group %q", a.Key)
		}
		seen = append(seen, strings.Join(slices.Concat(groups, []string{a.Key}), "."))
		return a
	}
	var buf bytes.Buffer
	logger := slog.New(golog.NewHandler(&buf, &golog.Options{ReplaceAttr: record}))
	logger.With("a", 1).WithGroup("g").With("b", 2).WithGroup("h").Info("m", slog.Group("i", "c", 3), "d", 4)

	want := []string{"a", "g.b", "time", "level", "msg", "g.h.i.c", "g.h.d"}
	if !slices.Equal(seen, want) {
		t.Errorf("ReplaceAttr saw %q, want %q", seen, want)
	}
}

var ansi = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func TestSlogtest(t *testing.T) {
	for _, color := range []bool{false, true} {
		t.Run(fmt.Sprintf("color=%v", color), func(t *testing.T) {
			var buf bytes.Buffer
			slogtest.Run(t, func(*testing.T) slog.Handler {
				buf.Reset()
				return golog.NewHandler(&buf, &golog.Options{Color: color})
			}, func(t *testing.T) map[string]any {
				return parse(t, ansi.ReplaceAllString(buf.String(), ""))
			})
		})
	}
}

// parse reads one line back into the map slogtest wants: the time if there is one, the level, the message
// (every word before the first key=value) and the attributes, with dotted keys as nested maps.
func parse(t *testing.T, line string) map[string]any {
	t.Helper()
	line, ok := strings.CutSuffix(line, "\n")
	if !ok || strings.Contains(line, "\n") {
		t.Fatalf("want one line, got %q", line)
	}
	words := fields(t, line)
	m := map[string]any{}
	if len(words) > 0 {
		if tm, err := time.Parse(time.RFC3339, words[0]); err == nil {
			m[slog.TimeKey] = tm
			words = words[1:]
		}
	}
	if len(words) == 0 {
		t.Fatalf("no level in %q", line)
	}
	m[slog.LevelKey], words = words[0], words[1:]
	var msg []string
	for len(words) > 0 && !strings.Contains(words[0], "=") {
		msg, words = append(msg, words[0]), words[1:]
	}
	m[slog.MessageKey] = strings.Join(msg, " ")
	for _, w := range words {
		k, v, _ := strings.Cut(w, "=")
		dst := m
		keys := strings.Split(k, ".")
		for _, g := range keys[:len(keys)-1] {
			sub, ok := dst[g].(map[string]any)
			if !ok {
				sub = map[string]any{}
				dst[g] = sub
			}
			dst = sub
		}
		dst[keys[len(keys)-1]] = v
	}
	return m
}

// fields splits a line on spaces, keeping quoted keys and values whole and unquoting them.
func fields(t *testing.T, line string) []string {
	var out []string
	for line != "" {
		line = strings.TrimLeft(line, " ")
		var word strings.Builder
		for line != "" && line[0] != ' ' {
			if line[0] != '"' {
				word.WriteByte(line[0])
				line = line[1:]
				continue
			}
			q, err := strconv.QuotedPrefix(line)
			if err != nil {
				t.Fatalf("bad quoting in %q: %v", line, err)
			}
			s, _ := strconv.Unquote(q)
			word.WriteString(s)
			line = line[len(q):]
		}
		out = append(out, word.String())
	}
	return out
}

// FuzzHandlerQuoting checks that values and messages come out as they went in, or quoted exactly as strconv.Quote
// quotes them, and that a quoted value or message reads back with strconv.Unquote.
func FuzzHandlerQuoting(f *testing.F) {
	for _, s := range []string{"", "plain", "two words", `a"b`, `back\slash`, "a=b", "tab\there", "new\nline",
		"\x00\x1b\x7f", "żółw", "a b", "a b", "\xff\xfe", "�", "emoji 🙂", "\U0010ffff"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		var buf bytes.Buffer
		r := slog.NewRecord(time.Time{}, slog.LevelInfo, s, 0)
		r.AddAttrs(slog.String("k", s))
		if err := golog.NewHandler(&buf, nil).Handle(t.Context(), r); err != nil {
			t.Fatal(err)
		}
		line, ok := strings.CutPrefix(buf.String(), "INFO  ")
		if !ok {
			t.Fatalf("line %q doesn't start with the level", buf.String())
		}
		value := strconv.Quote(s)
		if s != "" && !strings.ContainsFunc(s, func(r rune) bool {
			return r == utf8.RuneError || r == ' ' || r == '=' || r == '"' || unicode.IsSpace(r) || !unicode.IsPrint(r)
		}) {
			value = s
		}
		msg := strconv.Quote(s)
		if !strings.ContainsFunc(s, func(r rune) bool {
			return r == utf8.RuneError || !unicode.IsPrint(r) && !unicode.Is(unicode.Zs, r) // spaces of any kind are fine
		}) {
			msg = s
		}
		want := msg + " k=" + value + "\n"
		if s == "" {
			want = `k=""` + "\n"
		}
		if line != want {
			t.Fatalf("got  %q\nwant %q", line, want)
		}
	})
}

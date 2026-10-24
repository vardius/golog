package golog_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/vardius/golog/v2"
)

// attrs returns n attributes of the kinds a request log has: strings, ints, a bool, a duration and an error.
func attrs(n int) []slog.Attr {
	all := []slog.Attr{
		slog.String("method", "GET"),
		slog.String("path", "/api/v1/users"),
		slog.Int("status", 200),
		slog.Duration("duration", 1234567*time.Nanosecond),
		slog.Int64("bytes", 52341),
		slog.String("remote", "192.0.2.1:52341"),
		slog.Bool("cached", false),
		slog.Float64("ratio", 0.75),
		slog.Any("err", errors.New("connection reset by peer")),
		slog.String("user agent", "curl/8.7.1"),
	}
	return all[:n]
}

// TestAllocations guards the allocations per record. Lower a number when an optimisation lands;
// never raise one without a reason in the commit.
func TestAllocations(t *testing.T) {
	if raceEnabled {
		t.Skip("the race detector changes allocation counts")
	}

	ctx := context.Background()
	plain := slog.New(golog.NewHandler(io.Discard, nil))
	color := slog.New(golog.NewHandler(io.Discard, &golog.Options{Color: true}))
	source := slog.New(golog.NewHandler(io.Discard, &golog.Options{AddSource: true}))
	keep := func(_ []string, a slog.Attr) slog.Attr { return a }
	replace := slog.New(golog.NewHandler(io.Discard, &golog.Options{ReplaceAttr: keep, Level: slog.LevelDebug}))
	replaceSource := slog.New(golog.NewHandler(io.Discard, &golog.Options{ReplaceAttr: keep, AddSource: true}))
	grouped := plain.With("service", "api", "version", 2).WithGroup("http")
	five, ten := attrs(5), attrs(10)
	tenRecord := slog.NewRecord(time.Now(), slog.LevelInfo, "request", 0)
	tenRecord.AddAttrs(ten...)
	h := golog.NewHandler(io.Discard, nil)

	tests := []struct {
		name string
		call func()
		want float64
	}{
		{"message only", func() { plain.LogAttrs(ctx, slog.LevelInfo, "request") }, 0},
		{"5 attrs", func() { plain.LogAttrs(ctx, slog.LevelInfo, "request", five...) }, 0},
		{"10 attrs, straight to Handle", func() { _ = h.Handle(ctx, tenRecord) }, 0},
		// slog.Record keeps five attributes inline and allocates for the rest, whatever the handler.
		{"10 attrs, through a Logger", func() { plain.LogAttrs(ctx, slog.LevelInfo, "request", ten...) }, 1},
		{"WithAttrs, WithGroup and a group", func() {
			grouped.LogAttrs(ctx, slog.LevelInfo, "request", slog.Attr{Key: "req", Value: slog.GroupValue(five...)})
		}, 0},
		{"colour", func() { color.LogAttrs(ctx, slog.LevelInfo, "request", five...) }, 0},
		{"source", func() { source.LogAttrs(ctx, slog.LevelInfo, "request", five...) }, 0},
		{"source, in a func inlined into its caller", func() { logInlined(source) }, 0},
		{"below the level", func() { plain.LogAttrs(ctx, slog.LevelDebug, "request", five...) }, 0},
		{"ReplaceAttr", func() { replace.LogAttrs(ctx, slog.LevelInfo, "request", five...) }, 0},
		// ReplaceAttr is handed the level as an any: boxing a negative int allocates.
		{"ReplaceAttr, Debug", func() { replace.LogAttrs(ctx, slog.LevelDebug, "request", five...) }, 1},
		// ReplaceAttr is handed a *slog.Source, from slog.Record.Source: the Source and runtime.CallersFrames' two.
		{"ReplaceAttr and source", func() { replaceSource.LogAttrs(ctx, slog.LevelInfo, "request", five...) }, 3},
		// A key that needs quoting is quoted together with its group's name; a short join stays on the stack.
		{"key that needs quoting, in a group", func() {
			grouped.LogAttrs(ctx, slog.LevelInfo, "request", ten[9])
		}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := testing.AllocsPerRun(1000, tt.call); got != tt.want {
				t.Errorf("allocates %v times per record, want %v", got, tt.want)
			}
		})
	}
}

// logInlined is small enough to be inlined into its caller, which gives its call site a program counter
// that runtime.FuncForPC allocates for.
func logInlined(l *slog.Logger) { l.LogAttrs(context.Background(), slog.LevelInfo, "request") }

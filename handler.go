package golog

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"slices"
	"sync"
)

// LevelCritical is v1's Critical level: above slog.LevelError, for failures that need someone now.
// Handler prints it as CRITICAL; slog's own handlers print it as ERROR+4.
const LevelCritical = slog.LevelError + 4

// Options configure a Handler. The zero value logs Info and above, without source or colour.
type Options struct {
	// Level is the minimum level to log; records below it are dropped before their attributes are formatted.
	// Nil means slog.LevelInfo. Use a *slog.LevelVar to change it while the program runs.
	Level slog.Leveler

	// AddSource adds the file and line of the logging call, between the level and the message.
	AddSource bool

	// ReplaceAttr rewrites or drops attributes before they are written, with the same contract as
	// slog.HandlerOptions.ReplaceAttr: it is called for the built-in time, level, source and message
	// attributes (with slog.TimeKey, slog.LevelKey, slog.SourceKey and slog.MessageKey), and for every
	// other attribute that is not a group, with the groups it is in. Returning the zero Attr drops it.
	// The built-ins are written by position, so their keys are not printed.
	ReplaceAttr func(groups []string, a slog.Attr) slog.Attr

	// Color colours the level, dims the time, source and keys, and shows errors in red, with ANSI escape codes.
	// Turn it on for a terminal and off for files and pipes.
	Color bool
}

// Handler is a slog.Handler that writes each record as one line of text for people to read:
//
//	2026-10-24T09:41:07.312+11:00 INFO  server started addr=:8080 tls=false
//
// Its methods are safe to call concurrently. Handlers derived with WithAttrs and WithGroup share
// the writer and its lock, so their lines never interleave.
type Handler struct {
	opts   Options
	w      io.Writer
	mu     *sync.Mutex
	attrs  []byte   // attributes from WithAttrs, already formatted
	prefix string   // the open groups, each followed by a dot, prefixed to every record attribute's key
	groups []string // the open groups, for ReplaceAttr
}

// NewHandler returns a Handler that writes to w. A nil opts is the same as &Options{}.
func NewHandler(w io.Writer, opts *Options) *Handler {
	h := &Handler{w: w, mu: new(sync.Mutex)}
	if opts != nil {
		h.opts = *opts
	}
	if h.opts.Level == nil {
		h.opts.Level = slog.LevelInfo
	}
	return h
}

// Enabled reports whether level is at or above the handler's minimum level.
func (h *Handler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.opts.Level.Level()
}

// Handle formats r as one line and writes it to the handler's writer with a single Write call.
// It doesn't read ctx.
func (h *Handler) Handle(_ context.Context, r slog.Record) error {
	s := newState(h)
	defer s.free()

	s.appendBuiltins(r)
	if len(h.attrs) > 0 {
		s.sep()
		s.buf = append(s.buf, h.attrs...)
	}
	if r.NumAttrs() > 0 {
		s.prefix = append(s.prefix, h.prefix...)
		s.groups = append(s.groups, h.groups...)
		r.Attrs(func(a slog.Attr) bool {
			s.appendAttr(a)
			return true
		})
	}
	s.buf = append(s.buf, '\n')

	if err := h.write(s.buf); err != nil {
		return fmt.Errorf("golog: write record: %w", err)
	}
	return nil
}

func (h *Handler) write(line []byte) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := h.w.Write(line)
	return err
}

// WithAttrs returns a Handler that writes attrs on every line, after the message.
// They are formatted once, here, not on every record.
func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}
	s := newState(h)
	defer s.free()

	s.buf = append(s.buf, h.attrs...)
	s.prefix = append(s.prefix, h.prefix...)
	s.groups = append(s.groups, h.groups...)
	for _, a := range attrs {
		s.appendAttr(a)
	}

	h2 := *h
	h2.attrs = bytes.Clone(s.buf)
	return &h2
}

// WithGroup returns a Handler that qualifies the keys of later attributes with name and a dot,
// as in name.key=value. An empty name returns h.
func (h *Handler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	h2 := *h
	h2.prefix = h.prefix + name + "."
	h2.groups = append(slices.Clip(h.groups), name)
	return &h2
}

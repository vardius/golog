package golog

import (
	"bytes"
	"encoding"
	"fmt"
	"log/slog"
	"runtime"
	"strconv"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"
)

// ANSI escape codes for Options.Color.
const (
	ansiReset    = "\x1b[0m"
	ansiFaint    = "\x1b[2m"
	ansiRed      = "\x1b[31m"
	ansiDebug    = "\x1b[1;35m"    // bold magenta
	ansiInfo     = "\x1b[1;36m"    // bold cyan
	ansiWarn     = "\x1b[1;33m"    // bold yellow
	ansiError    = "\x1b[1;31m"    // bold red
	ansiCritical = "\x1b[1;97;41m" // bold white on red
)

// levelWidth is the width of the longest of slog's level names, so messages after them line up.
const levelWidth = len("ERROR")

// maxPooled caps the buffers kept for reuse, so one huge record doesn't pin its memory forever.
const maxPooled = 16 << 10

// state is the scratch space for formatting one record: the line, and the groups open at the attribute being written.
type state struct {
	h      *Handler
	buf    []byte
	prefix []byte   // the open groups, each followed by a dot
	groups []string // the open groups, for ReplaceAttr
	pad    int      // spaces still owed after the level, written before the next field

	// The last record time this state formatted, and the second and zone it is in. A record in the same second
	// and zone copies it and writes only its own milliseconds.
	clock    []byte
	clockDot int // the index of the '.' before the milliseconds
	clockSec int64
	clockLoc *time.Location
}

var statePool = sync.Pool{New: func() any { return &state{buf: make([]byte, 0, 1024)} }}

func newState(h *Handler) *state {
	s := statePool.Get().(*state)
	s.h = h
	return s
}

func (s *state) free() {
	if cap(s.buf) > maxPooled || cap(s.prefix) > maxPooled {
		return
	}
	s.h = nil
	s.buf, s.prefix, s.groups, s.pad = s.buf[:0], s.prefix[:0], s.groups[:0], 0
	statePool.Put(s)
}

// sep separates the next field from the last one.
func (s *state) sep() {
	if len(s.buf) == 0 {
		return
	}
	for ; s.pad > 0; s.pad-- {
		s.buf = append(s.buf, ' ')
	}
	s.buf = append(s.buf, ' ')
}

func (s *state) color(code string) {
	if s.h.opts.Color {
		s.buf = append(s.buf, code...)
	}
}

// appendBuiltins writes the time, level, source and message, by position and without their keys.
func (s *state) appendBuiltins(r slog.Record) {
	rep := s.h.opts.ReplaceAttr
	if !r.Time.IsZero() {
		if rep == nil {
			s.sep()
			s.color(ansiFaint)
			s.appendRecordTime(r.Time)
			s.color(ansiReset)
		} else if v, ok := s.replaced(slog.Time(slog.TimeKey, r.Time.Round(0))); ok {
			s.sep()
			s.color(ansiFaint)
			s.appendValue(v)
			s.color(ansiReset)
		}
	}

	if rep == nil {
		s.appendLevel(r.Level, r.Level)
	} else if v, ok := s.replaced(slog.Any(slog.LevelKey, r.Level)); ok {
		if l, isLevel := v.Any().(slog.Level); isLevel {
			s.appendLevel(r.Level, l)
		} else {
			s.sep()
			s.color(levelColor(r.Level))
			s.appendValue(v)
			s.color(ansiReset)
		}
	}

	if s.h.opts.AddSource && r.PC != 0 {
		s.appendSource(r)
	}

	if rep == nil {
		s.appendMessage(r.Message)
	} else if v, ok := s.replaced(slog.String(slog.MessageKey, r.Message)); ok {
		if v.Kind() == slog.KindString {
			s.appendMessage(v.String())
		} else {
			s.sep()
			s.appendValue(v)
		}
	}
}

// replaced calls ReplaceAttr on a built-in attribute, which is in no group, and reports whether the caller is to
// write the value it returns. A group it returns is written here, as key.member=value, as slog's handlers do.
func (s *state) replaced(a slog.Attr) (slog.Value, bool) {
	a = s.h.opts.ReplaceAttr(nil, a)
	a.Value = a.Value.Resolve()
	if a.Value.Kind() == slog.KindGroup {
		s.appendAttr(a)
		return slog.Value{}, false
	}
	return a.Value, !a.Equal(slog.Attr{})
}

// appendLevel writes label's name in the colour of the record's level, and owes the spaces that line up the message.
func (s *state) appendLevel(level, label slog.Level) {
	s.sep()
	s.color(levelColor(level))
	n := len(s.buf)
	s.buf = appendLevel(s.buf, label)
	s.pad = levelWidth - (len(s.buf) - n)
	s.color(ansiReset)
}

func levelColor(l slog.Level) string {
	switch {
	case l >= LevelCritical:
		return ansiCritical
	case l >= slog.LevelError:
		return ansiError
	case l >= slog.LevelWarn:
		return ansiWarn
	case l >= slog.LevelInfo:
		return ansiInfo
	default:
		return ansiDebug
	}
}

// appendLevel writes l the way slog.Level.String does, as DEBUG, INFO, WARN, ERROR or INFO+2,
// with CRITICAL from LevelCritical up.
func appendLevel(b []byte, l slog.Level) []byte {
	name, base := "DEBUG", slog.LevelDebug
	switch {
	case l >= LevelCritical:
		name, base = "CRITICAL", LevelCritical
	case l >= slog.LevelError:
		name, base = "ERROR", slog.LevelError
	case l >= slog.LevelWarn:
		name, base = "WARN", slog.LevelWarn
	case l >= slog.LevelInfo:
		name, base = "INFO", slog.LevelInfo
	}
	b = append(b, name...)
	if l == base {
		return b
	}
	if l > base {
		b = append(b, '+')
	}
	return strconv.AppendInt(b, int64(l-base), 10)
}

func (s *state) appendSource(r slog.Record) {
	if rep := s.h.opts.ReplaceAttr; rep != nil {
		src := r.Source()
		if src.File == "" { // a PC no function holds
			return
		}
		v, ok := s.replaced(slog.Any(slog.SourceKey, src))
		if !ok {
			return
		}
		s.sep()
		s.color(ansiFaint)
		if src, isSource := v.Any().(*slog.Source); isSource {
			s.appendFileLine(src.File, src.Line)
		} else {
			s.appendValue(v)
		}
		s.color(ansiReset)
		return
	}

	src := sourceOf(r.PC)
	if src == nil {
		return
	}
	s.sep()
	s.color(ansiFaint)
	s.appendFileLine(src.file, src.line)
	s.color(ansiReset)
}

type fileLine struct {
	file string
	line int
}

// fileLines holds the file and line of every call site logged from so far, by program counter. A program counter
// always means the same place, so the cache never goes stale, and it is bounded by the size of the program.
// A cached lookup takes a third of the time FuncForPC and FileLine take, and allocates nothing where FuncForPC
// allocates: for a call inlined into its caller.
var fileLines sync.Map // uintptr → *fileLine

// sourceOf returns the file and line of the call that pc, a return address, returns to, or nil if no function
// holds it.
func sourceOf(pc uintptr) *fileLine {
	if fl, ok := fileLines.Load(pc); ok {
		return fl.(*fileLine)
	}
	// The call is the instruction before the return address. FuncForPC and FileLine report the innermost
	// inlined function, as runtime.CallersFrames does.
	fn := runtime.FuncForPC(pc - 1)
	if fn == nil {
		return nil // not cached, so arbitrary integers can't grow the cache
	}
	file, line := fn.FileLine(pc - 1)
	fl, _ := fileLines.LoadOrStore(pc, &fileLine{file, line})
	return fl.(*fileLine)
}

func (s *state) appendFileLine(file string, line int) {
	s.appendString(file)
	s.buf = append(s.buf, ':')
	s.buf = strconv.AppendInt(s.buf, int64(line), 10)
}

// appendMessage writes msg as it is, or quoted if it holds a newline, another control character, a character
// that doesn't print, other than a space, or invalid UTF-8, so a record is always one line.
func (s *state) appendMessage(msg string) {
	if msg == "" {
		return
	}
	s.sep()
	if messageNeedsQuoting(msg) {
		s.buf = appendQuoted(s.buf, msg)
		return
	}
	s.buf = append(s.buf, msg...)
}

// appendAttr writes a as key=value, and a group's attributes as group.key=value. It follows slog's rules:
// values are resolved, empty attributes and groups are dropped, and a group with no key is inlined.
func (s *state) appendAttr(a slog.Attr) {
	a.Value = a.Value.Resolve()
	rep := s.h.opts.ReplaceAttr
	if rep != nil && a.Value.Kind() != slog.KindGroup {
		a = rep(s.groups, a)
		a.Value = a.Value.Resolve()
	}
	if a.Equal(slog.Attr{}) {
		return
	}

	if a.Value.Kind() != slog.KindGroup {
		s.appendKey(a.Key)
		s.appendValue(a.Value)
		return
	}

	attrs := a.Value.Group()
	if a.Key == "" {
		for _, ga := range attrs {
			s.appendAttr(ga)
		}
		return
	}
	n := len(s.prefix)
	s.prefix = append(append(s.prefix, a.Key...), '.')
	s.groups = append(s.groups, a.Key)
	for _, ga := range attrs {
		s.appendAttr(ga)
	}
	s.prefix = s.prefix[:n]
	s.groups = s.groups[:len(s.groups)-1]
}

func (s *state) appendKey(key string) {
	s.sep()
	s.color(ansiFaint)
	switch {
	case len(s.prefix) == 0:
		s.appendString(key)
	case needsQuoting(key) || bytesNeedQuoting(s.prefix):
		s.buf = appendQuoted(s.buf, string(s.prefix)+key)
	default:
		s.buf = append(s.buf, s.prefix...)
		s.buf = append(s.buf, key...)
	}
	s.buf = append(s.buf, '=')
	s.color(ansiReset)
}

func (s *state) appendValue(v slog.Value) {
	switch v.Kind() {
	case slog.KindString:
		s.appendString(v.String())
	case slog.KindInt64:
		s.buf = strconv.AppendInt(s.buf, v.Int64(), 10)
	case slog.KindUint64:
		s.buf = strconv.AppendUint(s.buf, v.Uint64(), 10)
	case slog.KindFloat64:
		s.buf = strconv.AppendFloat(s.buf, v.Float64(), 'g', -1, 64)
	case slog.KindBool:
		s.buf = strconv.AppendBool(s.buf, v.Bool())
	case slog.KindDuration:
		// Duration.String is inlined, so its short string stays on the stack.
		s.buf = append(s.buf, v.Duration().String()...)
	case slog.KindTime:
		s.buf = appendTime(s.buf, v.Time())
	default:
		s.appendAny(v.Any())
	}
}

// appendAny writes a value of any other type: as text if it can append or marshal itself,
// as its Error() if it is an error, quoted if it is a []byte, and with fmt's %+v otherwise.
func (s *state) appendAny(x any) {
	if _, isErr := x.(error); isErr {
		s.color(ansiRed)
		defer s.color(ansiReset)
	}

	start := len(s.buf)
	defer func() {
		// A method that panics, as Error does on a nil pointer, gets what fmt makes of it: <nil>, or the panic.
		if recover() != nil {
			s.buf = fmt.Appendf(s.buf[:start], "%+v", x)
			s.quoteFrom(start)
		}
	}()

	switch x := x.(type) {
	case encoding.TextAppender:
		b, e := x.AppendText(s.buf)
		if e != nil {
			s.buf = s.buf[:start]
			s.appendString("!ERROR:" + e.Error())
			return
		}
		s.buf = b
		s.quoteFrom(start)
	case encoding.TextMarshaler:
		b, e := x.MarshalText()
		if e != nil {
			s.appendString("!ERROR:" + e.Error())
			return
		}
		s.buf = append(s.buf, b...)
		s.quoteFrom(start)
	case error:
		s.appendString(x.Error())
	case []byte:
		s.buf = appendQuoted(s.buf, string(x))
	default:
		s.buf = fmt.Appendf(s.buf, "%+v", x)
		s.quoteFrom(start)
	}
}

// quoteFrom quotes the text written since start, if it needs it.
func (s *state) quoteFrom(start int) {
	if !bytesNeedQuoting(s.buf[start:]) {
		return
	}
	text := string(s.buf[start:])
	s.buf = appendQuoted(s.buf[:start], text)
}

func (s *state) appendString(str string) {
	if needsQuoting(str) {
		s.buf = appendQuoted(s.buf, str)
		return
	}
	s.buf = append(s.buf, str...)
}

// appendQuoted appends s quoted exactly as strconv.AppendQuote does, but copies the runs of characters that need
// no escaping in one go, where strconv goes rune by rune: for an error message, that was most of its cost.
func appendQuoted(b []byte, s string) []byte {
	b = append(b, '"')
	start := 0
	for i := 0; i < len(s); {
		c := s[i]
		if c >= ' ' && c < utf8.RuneSelf-1 && c != '"' && c != '\\' {
			i++
			continue
		}
		n := 1
		if c >= utf8.RuneSelf {
			var r rune
			r, n = utf8.DecodeRuneInString(s[i:])
			if r != utf8.RuneError && strconv.IsPrint(r) {
				i += n
				continue
			}
		}
		// strconv escapes this character, as \n or \xff: keep what it writes between the quotes.
		b = append(b, s[start:i]...)
		q := len(b)
		b = strconv.AppendQuote(b, s[i:i+n])
		b = append(b[:q], b[q+1:len(b)-1]...)
		i += n
		start = i
	}
	b = append(b, s[start:]...)
	return append(b, '"')
}

// appendRecordTime writes t as appendTime does, reusing the time this state formatted last when t is in the same
// second and zone: a zone's offset only changes on a whole second.
func (s *state) appendRecordTime(t time.Time) {
	sec, loc := t.Unix(), t.Location()
	if len(s.clock) == 0 || sec != s.clockSec || loc != s.clockLoc {
		s.clock = appendTime(s.clock[:0], t)
		s.clockDot = bytes.IndexByte(s.clock, '.')
		s.clockSec, s.clockLoc = sec, loc
	}
	n := len(s.buf)
	s.buf = append(s.buf, s.clock...)
	ms := t.Nanosecond() / int(time.Millisecond)
	millis := s.buf[n+s.clockDot+1 : n+s.clockDot+4]
	millis[0], millis[1], millis[2] = byte('0'+ms/100), byte('0'+ms/10%10), byte('0'+ms%10)
}

// appendTime writes t in RFC 3339 with milliseconds, as slog's handlers do: 2026-10-24T09:41:07.312+11:00.
// AppendFormat is fastest with RFC3339Nano, so it formats that and drops the extra digit: adding a tenth of a
// millisecond to the truncated time makes the fraction exactly four digits, as trailing zeros would be trimmed.
func appendTime(b []byte, t time.Time) []byte {
	n := len(b)
	b = t.Truncate(time.Millisecond).Add(time.Millisecond/10).AppendFormat(b, time.RFC3339Nano)
	dot := n + bytes.IndexByte(b[n:], '.')
	return append(b[:dot+4], b[dot+5:]...)
}

// needsQuoting reports whether s is empty or holds a space, '=', '"', a control character,
// a non-printing character or invalid UTF-8, so a key or value reads back unambiguously.
func needsQuoting(s string) bool {
	if s == "" {
		return true
	}
	for i := 0; i < len(s); {
		if c := s[i]; c < utf8.RuneSelf {
			if !safeASCII[c] {
				return true
			}
			i++
			continue
		}
		r, n := utf8.DecodeRuneInString(s[i:])
		if !safeRune(r) {
			return true
		}
		i += n
	}
	return false
}

func bytesNeedQuoting(b []byte) bool {
	if len(b) == 0 {
		return true
	}
	for i := 0; i < len(b); {
		if c := b[i]; c < utf8.RuneSelf {
			if !safeASCII[c] {
				return true
			}
			i++
			continue
		}
		r, n := utf8.DecodeRune(b[i:])
		if !safeRune(r) {
			return true
		}
		i += n
	}
	return false
}

// safeASCII holds the ASCII characters a key or value can hold unquoted: the printable ones but space, '=' and '"'.
var safeASCII = func() (safe [utf8.RuneSelf]bool) {
	for c := '!'; c <= '~'; c++ {
		safe[c] = c != '=' && c != '"'
	}
	return safe
}()

func safeRune(r rune) bool {
	return r != utf8.RuneError && !unicode.IsSpace(r) && unicode.IsPrint(r)
}

// messageNeedsQuoting reports whether msg holds anything that could break the line or the terminal:
// a control character, a non-printing character or invalid UTF-8. Spaces of any kind, '=' and '"' are fine;
// format characters such as bidi controls are not.
func messageNeedsQuoting(msg string) bool {
	for i := 0; i < len(msg); {
		if c := msg[i]; c < utf8.RuneSelf {
			if c < ' ' || c == utf8.RuneSelf-1 {
				return true
			}
			i++
			continue
		}
		r, n := utf8.DecodeRuneInString(msg[i:])
		if r == utf8.RuneError || !unicode.IsPrint(r) && !unicode.Is(unicode.Zs, r) {
			return true
		}
		i += n
	}
	return false
}

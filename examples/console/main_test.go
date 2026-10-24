package main

import (
	"bytes"
	"regexp"
	"strings"
	"testing"
)

func TestServe(t *testing.T) {
	var buf bytes.Buffer
	serve(newLogger(&buf, false, false))

	want := []string{
		`INFO  server started addr=:8080`,
		`INFO  request http.method=GET http.path=/ http.status=200 http.took=3ms`,
		`WARN  slow request http.method=GET http.path=/search http.status=200 http.took=1.204s`,
		`ERROR payment failed order=1042 err="card declined"`,
	}
	lines := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")
	if len(lines) != len(want) {
		t.Fatalf("got %d lines, want %d:\n%s", len(lines), len(want), buf.String())
	}
	for i, line := range lines {
		// Every line starts with the time.
		time, rest, _ := strings.Cut(line, " ")
		if !regexp.MustCompile(`^\d{4}-\d\d-\d\dT\d\d:\d\d:\d\d\.\d{3}`).MatchString(time) || rest != want[i] {
			t.Errorf("line %d = %q, want the time and %q", i, line, want[i])
		}
	}
}

func TestServeDebugInColour(t *testing.T) {
	var buf bytes.Buffer
	serve(newLogger(&buf, true, true))

	out := buf.String()
	if !strings.Contains(out, "\x1b[1;35mDEBUG\x1b[0m config loaded") {
		t.Errorf("no Debug record in magenta in:\n%q", out)
	}
	if strings.Count(out, "\n") != 5 {
		t.Errorf("got %d lines, want 5", strings.Count(out, "\n"))
	}
}

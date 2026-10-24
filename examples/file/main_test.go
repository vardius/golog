package main

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

func TestRun(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.log")
	var console bytes.Buffer
	for range 2 { // the second run appends
		if err := run(path, &console); err != nil {
			t.Fatal(err)
		}
	}

	file, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	plain := regexp.MustCompile(`^(\S+ DEBUG cache warmed entries=1200\n\S+ INFO  server started addr=:8080\n){2}$`)
	if !plain.Match(file) {
		t.Errorf("file holds:\n%s\nwant both records from both runs, without colour", file)
	}
	coloured := regexp.MustCompile(`^(\x1b\[2m\S+\x1b\[0m \x1b\[1;36mINFO\x1b\[0m  server started \x1b\[2maddr=\x1b\[0m:8080\n){2}$`)
	if !coloured.Match(console.Bytes()) {
		t.Errorf("console shows:\n%q\nwant the Info record from both runs, in colour", console.String())
	}
}

func TestRunCannotOpen(t *testing.T) {
	err := run(filepath.Join(t.TempDir(), "missing", "app.log"), &bytes.Buffer{})
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("run() error = %v, want %v", err, fs.ErrNotExist)
	}
}

package main

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/vardius/golog/v2"
)

func TestRequestID(t *testing.T) {
	var buf bytes.Buffer
	h := golog.NewHandler(&buf, &golog.Options{Level: slog.LevelDebug})
	srv := httptest.NewServer(newServer(slog.New(contextHandler{h})))
	defer srv.Close()

	for range 2 {
		resp, err := http.Get(srv.URL + "/hello")
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		if string(body) != "hello\n" {
			t.Fatalf("body = %q, want %q", body, "hello\n")
		}
	}

	want := regexp.MustCompile(`^\S+ DEBUG saying hello request_id=1
\S+ INFO  served method=GET path=/hello took=\S+ request_id=1
\S+ DEBUG saying hello request_id=2
\S+ INFO  served method=GET path=/hello took=\S+ request_id=2
$`)
	if !want.Match(buf.Bytes()) {
		t.Errorf("logged:\n%s\nwant each request's records with its own ID", buf.String())
	}
}

func TestNoRequestID(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(contextHandler{golog.NewHandler(&buf, nil)})
	logger.With("a", 1).WithGroup("g").InfoContext(t.Context(), "outside a request", "b", 2)

	if want := regexp.MustCompile(`^\S+ INFO  outside a request a=1 g\.b=2\n$`); !want.Match(buf.Bytes()) {
		t.Errorf("logged %q, want no request_id", buf.String())
	}
}

// Command requestid serves HTTP and logs every request with an ID that its middleware puts in the request's
// context. The logger finds it there: a handler wrapped around golog's reads it on every record.
//
//	go run ./examples/requestid
//	curl localhost:8080/hello
package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/vardius/golog/v2"
)

type requestIDKey struct{}

// contextHandler adds the request ID in the context, if there is one, to each record it hands on.
type contextHandler struct{ slog.Handler }

func (h contextHandler) Handle(ctx context.Context, r slog.Record) error {
	if id, ok := ctx.Value(requestIDKey{}).(string); ok {
		r.AddAttrs(slog.String("request_id", id))
	}
	return h.Handler.Handle(ctx, r)
}

func (h contextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return contextHandler{h.Handler.WithAttrs(attrs)}
}

func (h contextHandler) WithGroup(name string) slog.Handler {
	return contextHandler{h.Handler.WithGroup(name)}
}

// withRequestID gives every request an ID in its context and logs it when it's served.
func withRequestID(logger *slog.Logger, next http.Handler) http.Handler {
	var last atomic.Uint64
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ctx := context.WithValue(r.Context(), requestIDKey{}, strconv.FormatUint(last.Add(1), 10))
		next.ServeHTTP(w, r.WithContext(ctx))
		logger.InfoContext(ctx, "served", "method", r.Method, "path", r.URL.Path, "took", time.Since(start))
	})
}

func newServer(logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /hello", func(w http.ResponseWriter, r *http.Request) {
		logger.DebugContext(r.Context(), "saying hello")
		w.Write([]byte("hello\n"))
	})
	return withRequestID(logger, mux)
}

func main() {
	h := golog.NewHandler(os.Stderr, &golog.Options{Level: slog.LevelDebug})
	logger := slog.New(contextHandler{h})
	log.Fatal(http.ListenAndServe("localhost:8080", newServer(logger)))
}

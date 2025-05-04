package mdlwr

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/otel/trace"
)

func Logger(logger slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			start := time.Now()
			defer func() {
				logger.LogAttrs(r.Context(), slog.LevelInfo, "request completed",
					slog.String("request-id", GetReqID(r.Context())),
					slog.Int("status", ww.Status()),
					slog.Int("bytes", ww.BytesWritten()),
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.String("query", r.URL.RawQuery),
					slog.String("ip", r.RemoteAddr),
					slog.String("trace.id", trace.SpanFromContext(r.Context()).SpanContext().TraceID().String()),
					slog.String("user-agent", r.UserAgent()),
					slog.Duration("latency", time.Since(start)),
				)
			}()

			next.ServeHTTP(ww, r)
		}
		return http.HandlerFunc(fn)
	}
}

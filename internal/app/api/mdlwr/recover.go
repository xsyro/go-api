package mdlwr

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-stack/stack"
	"go.opentelemetry.io/otel/trace"
)

func Recover(logger slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(rw http.ResponseWriter, r *http.Request) {
			defer func() {
				if p := recover(); p != nil {
					err, ok := p.(error)
					if !ok {
						err = fmt.Errorf("%v", p)
					}

					var stackTrace stack.CallStack
					// Get the current stacktrace but trim the runtime
					traces := stack.Trace().TrimRuntime()

					// Format the stack trace removing the clutter from it
					for i := 0; i < len(traces); i++ {
						t := traces[i]
						tFunc := t.Frame().Function

						// Opentelemetry is recovering from the panics on span.End defets and throwing them again
						// we don't want this noise to appear on our logs
						if tFunc == "runtime.gopanic" || tFunc == "go.opentelemetry.io/otel/sdk/trace.(*span).End" {
							continue
						}

						// This call is made before the code reaching our handlers, we don't want to log things that are coming before
						// our own code, just from our handlers and donwards.
						if tFunc == "net/http.HandlerFunc.ServeHTTP" {
							break
						}
						stackTrace = append(stackTrace, t)
					}

					// sticking with zerolog's .PanicLevel - 5
					logger.LogAttrs(r.Context(), 5, "panic",
						slog.Any("error", err),
						slog.String("trace.id", trace.SpanFromContext(r.Context()).SpanContext().TraceID().String()),
						slog.String("request-id", GetReqID(r.Context())),
						slog.String("stack", fmt.Sprintf("%+v", stackTrace)),
					)

					http.Error(rw, http.StatusText(http.StatusInternalServerError),
						http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(rw, r)
		}
		return http.HandlerFunc(fn)
	}
}

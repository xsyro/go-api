package presenter

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/xsyro/goapi/internal/app/api/mdlwr"
	"github.com/xsyro/goapi/internal/utils"
	"go.opentelemetry.io/otel/trace"
)

// Presenter will format the response appropriately to
// our client and also log the occurrence
type Presenter interface {
	JSON(w http.ResponseWriter, r *http.Request, v interface{})
	Error(w http.ResponseWriter, r *http.Request, err error)
}

type presenter struct {
	logger slog.Logger
}

func NewPresenters(logger slog.Logger) Presenter {
	return &presenter{logger: logger}
}

func (p *presenter) JSON(w http.ResponseWriter, r *http.Request, v interface{}) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	err := utils.WriteJson(w, v)
	if err != nil {
		p.Error(w, r, err)
	}
}

func (p *presenter) Error(w http.ResponseWriter, r *http.Request, err error) {
	span := trace.SpanFromContext(r.Context())
	span.RecordError(err)

	var apiErr utils.APIError
	if errors.As(err, &apiErr) {
		// We can retrieve the StatusError here and write out a specific
		// HTTP status code and an API-safe error message
		se := apiErr.APIError()

		p.logger.LogAttrs(r.Context(), slog.LevelError, "error",
			slog.Any("error", err),
			slog.String("caller", se.Caller),
			slog.String("request-id", mdlwr.GetReqID(r.Context())),
			slog.String("trace.id", span.SpanContext().TraceID().String()),
		)

		utils.JsonError(w, se.Code, se.Msg)
		return
	} else {
		// Any error types we don't specifically look out for default
		// to serving a HTTP Internal Server Error

		p.logger.LogAttrs(r.Context(), slog.LevelError, "unhandled error",
			slog.Any("error", err),
			slog.String("request-id", mdlwr.GetReqID(r.Context())),
			slog.String("trace.id", span.SpanContext().TraceID().String()),
		)

		utils.JsonError(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}
}

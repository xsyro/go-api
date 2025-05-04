package mdlwr

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type ctxKey int

const ridKey ctxKey = ctxKey(0)

func GetReqID(ctx context.Context) string {
	return ctx.Value(ridKey).(string)
}

func RequestId(next http.Handler) http.Handler {
	fn := func(rw http.ResponseWriter, r *http.Request) {
		rid := r.Header.Get("X-Request-ID")
		if rid == "" {
			rid = uuid.NewString()
		}
		ctx := context.WithValue(r.Context(), ridKey, rid)
		rw.Header().Add("X-Request-ID", rid)

		next.ServeHTTP(rw, r.WithContext(ctx))
	}

	return http.HandlerFunc(fn)
}

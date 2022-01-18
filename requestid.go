package requestid

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/twitchtv/twirp"
)

type ctxKey int

const ridKey ctxKey = ctxKey(0)

const HeaderName = "X-Request-ID"

// NewContext creates a context with request id set as context value
// as well as twirp request header.
func NewContext(ctx context.Context, rid string) context.Context {
	headers, ok := twirp.HTTPRequestHeaders(ctx)
	if !ok {
		headers = http.Header{}
	}
	headers.Set(HeaderName, rid)
	ctx, _ = twirp.WithHTTPRequestHeaders(ctx, headers)
	return context.WithValue(ctx, ridKey, rid)
}

// FromContext returns the request id from context.
func FromContext(ctx context.Context) (string, bool) {
	rid, ok := ctx.Value(ridKey).(string)
	return rid, ok
}

// RequestIDHandler sets unique request id.
// If header `X-Request-ID` is already present in the request, that is considered the
// request id. Otherwise, generates a new unique ID.
func RequestIDHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rid := r.Header.Get(HeaderName)
		if rid == "" {
			rid = uuid.New().String()
			r.Header.Set(HeaderName, rid)
		}
		w.Header().Set(HeaderName, rid)
		ctx := NewContext(r.Context(), rid)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

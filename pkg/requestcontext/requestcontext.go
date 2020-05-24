package requestcontext

import (
	"context"
	"net/http"

	"github.com/boltdb/bolt"
)

type ctxKey struct{}

// RequestContext holds DB connection and stuff
type RequestContext struct {
	DB     *bolt.DB
	Errors chan error
}

// New initializes a new RequestContext.
func New(db *bolt.DB, errorChan chan error) *RequestContext {
	return &RequestContext{
		DB:     db,
		Errors: errorChan,
	}
}

// WithContext create a new context from the given one and stores RequestContext object in it.
func WithContext(ctx context.Context, rc *RequestContext) context.Context {
	return context.WithValue(ctx, ctxKey{}, rc)
}

// GetRequestContext returns the RequestContext stored in the given context.
func GetRequestContext(ctx context.Context) *RequestContext {
	if val, ok := ctx.Value(ctxKey{}).(*RequestContext); ok {
		return val
	}
	return nil
}

// DB returns DB stored in the context
func DB(ctx context.Context) *bolt.DB {
	if r := GetRequestContext(ctx); r != nil {
		return r.DB
	}
	return nil
}

// Errors returns error chan stored in the context
func Errors(ctx context.Context) chan error {
	if r := GetRequestContext(ctx); r != nil {
		return r.Errors
	}
	return nil
}

// InjectRequestContextMiddleware injects a given request context into HTTP request.
func InjectRequestContextMiddleware(handler http.Handler, rc *RequestContext) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := WithContext(r.Context(), rc)
		newReq := r.WithContext(ctx)
		handler.ServeHTTP(w, newReq)
	})
}

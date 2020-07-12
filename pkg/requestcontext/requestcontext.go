package requestcontext

import (
	"context"
	"net/http"

	"github.com/boltdb/bolt"
	"github.com/mkorenkov/covid-19/pkg/config"
	"github.com/mkorenkov/covid-19/pkg/documents"
)

type ctxKey struct{}

// RequestContext holds DB connection and stuff
type RequestContext struct {
	Config   config.Config
	DB       *bolt.DB
	Errors   chan error
	UploadS3 chan documents.CollectionEntry
}

// New initializes a new RequestContext.
func New(cfg config.Config, db *bolt.DB, errorChan chan error, s3backup chan documents.CollectionEntry) *RequestContext {
	return &RequestContext{
		Config:   cfg,
		DB:       db,
		Errors:   errorChan,
		UploadS3: s3backup,
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

// ToS3 returns documents.CollectionEntry to upload payloads to S3
func ToS3(ctx context.Context) chan documents.CollectionEntry {
	if r := GetRequestContext(ctx); r != nil {
		return r.UploadS3
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

package server

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
)

// PanicRecoveryMiddleware recovers from panic().
func PanicRecoveryMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				buf := make([]byte, 1<<16)
				runtime.Stack(buf, true)
				fmt.Fprintf(w, "INTERNAL SERVER ERROR \n%s\n%s", err, buf)
				log.Printf("[ERROR] %s\n%s", err, buf)
			}
		}()
		handler.ServeHTTP(w, r)
	})
}

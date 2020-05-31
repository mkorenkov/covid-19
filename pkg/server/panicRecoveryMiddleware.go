package server

import (
	"fmt"
	"log"
	"net/http"
	"runtime"

	"github.com/mkorenkov/covid-19/pkg/reporter"
	"github.com/pkg/errors"
)

// PanicRecoveryMiddleware recovers from panic().
func PanicRecoveryMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if smth := recover(); smth != nil {
				var err error
				if asErr, ok := err.(error); ok {
					err = asErr
				} else {
					err = errors.Errorf("Non-error panic %v", smth)
				}
				w.WriteHeader(http.StatusInternalServerError)
				buf := make([]byte, 1<<16)
				runtime.Stack(buf, true)
				fmt.Fprintf(w, "INTERNAL SERVER ERROR \n%s\n%s", err, buf)
				log.Printf("[ERROR] %s\n%s", err, buf)
				reporter.Report(err)
			}
		}()
		handler.ServeHTTP(w, r)
	})
}

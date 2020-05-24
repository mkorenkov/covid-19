package server

import (
	"log"
	"net/http"
	"strings"
	"time"
)

type logResponseWrapper struct {
	http.ResponseWriter

	status        int
	responseBytes int64
}

func (r *logResponseWrapper) Write(p []byte) (int, error) {
	written, err := r.ResponseWriter.Write(p)
	r.responseBytes += int64(written)
	return written, err
}

func (r *logResponseWrapper) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

//LogMiddleware simple log middleware
func LogMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := r.RemoteAddr
		if colon := strings.LastIndex(clientIP, ":"); colon != -1 {
			clientIP = clientIP[:colon]
		}

		wrapper := &logResponseWrapper{ResponseWriter: w, status: http.StatusOK}
		startTime := time.Now()
		handler.ServeHTTP(wrapper, r)
		finishTime := time.Now()
		elapsedTime := finishTime.Sub(startTime)

		log.Printf("[INFO] %s %s %s %d (%d bytes in %.4fms) by %s agent %s", r.Method, r.URL, r.Proto, wrapper.status, wrapper.responseBytes, elapsedTime.Seconds()*1000, clientIP, r.UserAgent())
	})
}

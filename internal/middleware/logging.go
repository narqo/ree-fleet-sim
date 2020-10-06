package middleware

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

const headerRequestID = "X-Request-Id"

func LoggingHandler(out io.Writer, next http.Handler) http.Handler {
	h := func(w http.ResponseWriter, r *http.Request) {
		ts := time.Now().UTC()

		resp := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(resp, r)

		fmt.Fprintf(
			out,
			"ts=%s method=%s uri=%s code=%v rtime=%s\n",
			ts.Format(time.RFC3339),
			r.Method,
			r.RequestURI,
			resp.statusCode,
			time.Since(ts),
		)
	}
	return http.HandlerFunc(h)
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (r *responseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.statusCode = statusCode
}

func (r *responseWriter) Flush() {
	if f, ok := r.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

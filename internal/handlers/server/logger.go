package server

import (
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

type (
	loggingData struct {
		status int
		size   int
		data   string
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *loggingData
	}
)

var _ http.ResponseWriter = &loggingResponseWriter{}

//nolint:wrapcheck
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	r.responseData.data += string(b)

	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func WithLogging(h http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &loggingData{
			status: 0,
			size:   0,
		}
		lw := &loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}

		h.ServeHTTP(lw, r)

		duration := time.Since(start)

		log.Info().
			Str("uri", r.RequestURI).
			Str("method", r.Method).
			Int("status", responseData.status).
			Dur("duration", duration).
			Int("size", responseData.size).
			Str("data", responseData.data).
			Msg("new request")
	}

	return http.HandlerFunc(logFn)
}

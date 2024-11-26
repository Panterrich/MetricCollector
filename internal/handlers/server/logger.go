package server

import (
	"net/http"
)

type (
	responseData struct {
		status int
		size   int
		data   string
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

//nolint:wrapcheck
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	r.responseData.data = string(b)

	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

package server

import (
	"compress/gzip"
	"io"
	"net/http"
	"slices"
	"strings"
)

var contentTypeCompressed = map[string]struct{}{
	"application/json": {},
	"text/html":        {},
}

type gzipWriter struct {
	http.ResponseWriter
	Writer     io.Writer
	StatusCode int
}

var _ io.Writer = &gzipWriter{}

func (w *gzipWriter) WriteHeader(statusCode int) {
	w.StatusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

//nolint:wrapcheck
func (w *gzipWriter) Write(b []byte) (int, error) {
	contentType := w.Header().Get("Content-Type")

	if _, ok := contentTypeCompressed[contentType]; !ok {
		if w.StatusCode == http.StatusOK {
			w.Header().Del("Content-Encoding")
			return w.ResponseWriter.Write(b)
		}
	}

	return w.Writer.Write(b)
}

func WithGzipCompression(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		encodings := r.Header.Values("Accept-Encoding")

		isGzipEncoding := func(encoding string) bool {
			return strings.Contains(encoding, "gzip")
		}

		if !slices.ContainsFunc(encodings, isGzipEncoding) {
			next.ServeHTTP(w, r)
			return
		}

		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		defer gz.Close()

		w.Header().Set("Content-Encoding", "gzip")

		next.ServeHTTP(&gzipWriter{ResponseWriter: w, Writer: gz}, r)
	})
}

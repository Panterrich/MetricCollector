package server

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/Panterrich/MetricCollector/pkg/hash"
)

type (
	hashingResponseWriter struct {
		http.ResponseWriter
		responseData *bytes.Buffer
	}
)

var _ http.ResponseWriter = &hashingResponseWriter{}

//nolint:wrapcheck
func (r *hashingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.responseData.Write(b)
	if err != nil {
		return size, fmt.Errorf("hashing wrapper write: %w", err)
	}

	return r.ResponseWriter.Write(b)
}

func WithHashing(key []byte) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		hashFn := func(w http.ResponseWriter, r *http.Request) {
			var (
				bodyBytes []byte
				err       error
				check     bool
			)

			if r.Body != nil {
				bodyBytes, err = io.ReadAll(r.Body)
				if err != nil {
					http.Error(w, fmt.Sprintf("read body error: %v", err), http.StatusInternalServerError)
					return
				}
			}

			rdr := io.NopCloser(bytes.NewBuffer(bodyBytes))
			r.Body = rdr

			h := r.Header.Get("HashSHA256")
			if h != "" {
				check, err = hash.CheckMessage(bodyBytes, key, []byte(h))
				if err != nil {
					http.Error(w, fmt.Sprintf("hash check message: %v", err), http.StatusInternalServerError)
					return
				}

				if !check {
					http.Error(w, fmt.Sprintf("hash message invalid: %v", err), http.StatusBadRequest)
					return
				}
			}

			hw := &hashingResponseWriter{
				ResponseWriter: w,
				responseData:   bytes.NewBuffer(nil),
			}

			next.ServeHTTP(hw, r)

			rhash, err := hash.Message(hw.responseData.Bytes(), key)
			if err != nil {
				http.Error(w, fmt.Sprintf("hashing response: %v", err), http.StatusInternalServerError)
				return
			}

			w.Header().Set("HashSHA256", string(rhash))
		}

		return http.HandlerFunc(hashFn)
	}
}

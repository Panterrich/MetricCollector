package server

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"

	"github.com/Panterrich/MetricCollector/pkg/hash"
)

type (
	hashingResponseWriter struct {
		http.ResponseWriter
		headers      http.Header
		statusCode   int
		wroteHeader  bool
		responseData *bytes.Buffer
	}
)

var _ http.ResponseWriter = &hashingResponseWriter{}

func (r *hashingResponseWriter) Header() http.Header {
	return r.headers
}

func (r *hashingResponseWriter) WriteHeader(statusCode int) {
	if r.wroteHeader {
		return
	}

	r.statusCode = statusCode
	r.wroteHeader = true
}

//nolint:wrapcheck
func (r *hashingResponseWriter) Write(b []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}

	size, err := r.responseData.Write(b)
	if err != nil {
		return size, fmt.Errorf("hashing wrapper write: %w", err)
	}

	return size, nil
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

			hBytes, err := hex.DecodeString(h)
			if err != nil {
				http.Error(w, fmt.Sprintf("decode hash: %v", err), http.StatusBadRequest)
				return
			}

			check, err = hash.CheckMessage(bodyBytes, key, hBytes)
			if err != nil {
				http.Error(w, fmt.Sprintf("hash check message: %v", err), http.StatusInternalServerError)
				return
			}

			if !check {
				http.Error(w, fmt.Sprintf("hash message invalid: %v", h), http.StatusBadRequest)
				return
			}

			hw := &hashingResponseWriter{
				ResponseWriter: w,
				headers:        make(http.Header),
				responseData:   bytes.NewBuffer(nil),
			}

			next.ServeHTTP(hw, r)

			rhash, err := hash.Message(hw.responseData.Bytes(), key)
			if err != nil {
				http.Error(w, fmt.Sprintf("hashing response: %v", err), http.StatusInternalServerError)
				return
			}

			for key, values := range hw.Header() {
				for _, value := range values {
					w.Header().Add(key, value)
				}
			}

			w.Header().Set("HashSHA256", hex.EncodeToString(rhash))

			if hw.statusCode == 0 {
				hw.statusCode = http.StatusOK
			}

			w.WriteHeader(hw.statusCode)

			_, _ = w.Write(hw.responseData.Bytes())
		}

		return http.HandlerFunc(hashFn)
	}
}

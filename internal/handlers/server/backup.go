package server

import (
	"net/http"

	"github.com/Panterrich/MetricCollector/internal/collector"
	"github.com/Panterrich/MetricCollector/pkg/serialization"
	"github.com/rs/zerolog/log"
)

func WithBackup(collector collector.Collector, path string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)

			if err := serialization.Save(collector, path); err != nil {
				log.Error().Msgf("can't save database: %v", err)
			}
		})
	}
}

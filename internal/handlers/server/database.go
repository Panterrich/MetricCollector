package server

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/Panterrich/MetricCollector/internal/collector"
	"github.com/Panterrich/MetricCollector/internal/storages"
)

type DatabaseHandler func(db *sql.DB, w http.ResponseWriter, r *http.Request)

func WithDatabase(c collector.Collector, next DatabaseHandler) http.HandlerFunc {
	if db, ok := c.(*storages.Database); ok {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next(db.DB, w, r)
		})
	}

	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})
}

// @Summary Pings the database
// @Description Checks the availability of the database by pinging it
// @Tags Database
// @Accept text/plain
// @Produce text/plain
// @Success 200 "Pong"
// @Failure 500 {string} error "Failed to ping database"
// @Router /ping [get]
func PingDatabase(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		http.Error(w, fmt.Sprintf("ping database is failed: %v", err), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
}

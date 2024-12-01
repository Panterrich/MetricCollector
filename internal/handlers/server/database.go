package server

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"
)

type DatabaseHandler func(db *sql.DB, w http.ResponseWriter, r *http.Request)

func WithDatabase(db *sql.DB, next DatabaseHandler) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		next(db, w, r)
	}

	return http.HandlerFunc(fn)
}

func PingDatabase(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		http.Error(w, fmt.Sprintf("ping database is failed: %v", err), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
}

package main

import (
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/ignius299792458/techkraft-ch-svr/utils"
)

func main() {
	mux := http.NewServeMux()

	// pass to resolver
	resolver(mux)

	// Wrap with CORS middleware
	handler := utils.CORS(mux)

	// server setup
	const addr string = ":8080"
	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Start server
	slog.Info("server starting", "addr", addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("server error", "err", err)
		os.Exit(1)
	}
}

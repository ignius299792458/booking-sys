package main

import (
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"
)

func main() {
	mux := http.NewServeMux()

	// pass to resolver
	resolver(mux)

	// server setup
	const addr string = ":8080"
	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
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

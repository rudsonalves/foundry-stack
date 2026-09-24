package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
)

type httpServer interface {
	ListenAndServe() error
	Shutdown(context.Context) error
	Close() error
}

// serve starts the HTTP server and handles any errors that occur during its execution.
// If the server is closed gracefully, it returns nil. For any other errors, it wraps and returns the error.
func serve(
	ctx context.Context,
	server httpServer,
	shutdownTimeout time.Duration,
) error {
	errorsCh := make(chan error, 1)

	go func() {
		errorsCh <- server.ListenAndServe()
	}()

	select {
	case err := <-errorsCh:
		return normalizeServeError(err)
	case <-ctx.Done():
	}

	log.Println("shutdown initiated")

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		shutdownTimeout,
	)
	defer cancel()

	shutdownErr := server.Shutdown(shutdownCtx)
	if shutdownErr != nil {
		log.Printf(
			"graceful shutdown failed; forcing server close: %v",
			shutdownErr,
		)

		closeErr := server.Close()
		serveErr := <-errorsCh

		return errors.Join(
			fmt.Errorf(
				"shutdown HTTP server: %w",
				shutdownErr,
			),
			normalizeCloseError(closeErr),
			normalizeServeError(serveErr),
		)
	}

	serveErr := normalizeServeError(<-errorsCh)
	if serveErr != nil {
		return serveErr
	}

	log.Println("shutdown completed")
	return nil
}

func normalizeServeError(err error) error {
	if err == nil || errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return fmt.Errorf("serve HTTP: %w", err)
}
func normalizeCloseError(err error) error {
	if err == nil || errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return fmt.Errorf("close HTTP server: %w", err)
}

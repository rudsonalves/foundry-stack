package main

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/rudsonalves/foundry-stack/api/internal/bootstrap"
)

func TestMainRejectsInvalidFlags(t *testing.T) {
	if os.Getenv("FOUNDRY_STACK_TEST_INVOKE_MAIN") == "1" {
		os.Args = []string{"foundry-stack", "-definitely-invalid"}
		main()
		return
	}

	command := exec.Command(os.Args[0], "-test.run=^TestMainRejectsInvalidFlags$")
	command.Env = append(os.Environ(), "FOUNDRY_STACK_TEST_INVOKE_MAIN=1")
	output, err := command.CombinedOutput()
	if err == nil {
		t.Fatal("main() with invalid flags unexpectedly succeeded")
	}
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
		t.Fatalf("main() error = %v, want exit code 1", err)
	}
	if !strings.Contains(string(output), "invalid flags") {
		t.Fatalf("main() output = %q, want invalid flags", output)
	}
}

func TestShouldLogEmailVerificationCode(t *testing.T) {
	tests := []struct {
		environment string
		want        bool
	}{
		{environment: bootstrap.EnvDevelopment, want: true},
		{environment: bootstrap.EnvStaging, want: true},
		{environment: bootstrap.EnvProduction, want: false},
		{environment: bootstrap.EnvTest, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.environment, func(t *testing.T) {
			got := shouldLogEmailVerificationCode(tt.environment)
			if got != tt.want {
				t.Fatalf("shouldLogEmailVerificationCode(%q) = %t, want %t", tt.environment, got, tt.want)
			}
		})
	}
}

func TestRunRejectsInvalidFlags(t *testing.T) {
	t.Parallel()

	err := run(context.Background(), []string{"-definitely-invalid"})
	if err == nil {
		t.Fatal("run() with invalid flags unexpectedly succeeded")
	}
	if !strings.Contains(err.Error(), "invalid flags") {
		t.Fatalf("run() error = %q, want invalid flags", err)
	}
}

func TestRunConfiguredClosesDatabaseAfterServerStops(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("sqlmock.New() unexpected error: %v", err)
	}
	t.Cleanup(func() {
		sqlOpen = sql.Open
		buildServer = func(cfg bootstrap.Config, handler http.Handler) httpServer {
			return newServer(cfg, handler)
		}
		db.Close()
	})

	mock.ExpectPing()
	mock.ExpectPing()
	mock.ExpectClose()
	sqlOpen = func(string, string) (*sql.DB, error) {
		return db, nil
	}

	server := newControlledHTTPServer()
	buildServer = func(bootstrap.Config, http.Handler) httpServer {
		return server
	}
	cfg := bootstrap.Config{
		Environment:     bootstrap.EnvTest,
		Port:            "8080",
		AllowedOrigin:   "http://localhost:5173",
		ReadTimeout:     time.Second,
		ShutdownTimeout: time.Second,
		Database: bootstrap.DBConfig{
			URL:             "postgres://test",
			MaxOpenConns:    1,
			MaxIdleConns:    1,
			ConnMaxIdleTime: time.Minute,
			ConnMaxLifetime: time.Minute,
		},
		JWT: bootstrap.JWTConfig{
			Secret:     bytes.Repeat([]byte("a"), 32),
			Issuer:     "test",
			Audience:   "test",
			ClientIDs:  []string{"test"},
			AccessTTL:  time.Minute,
			RefreshTTL: time.Minute,
		},
		EmailVerification: bootstrap.EmailVerificationConfig{
			CodeSecret:              bytes.Repeat([]byte("b"), 32),
			CodeTTL:                 10 * time.Minute,
			TokenSecret:             bytes.Repeat([]byte("c"), 32),
			TokenTTL:                24 * time.Hour,
			MaxConfirmAttempts:      5,
			ResendCooldown:          time.Minute,
			ResendEmailLimitPerHour: 5,
			ResendIPLimitPerHour:    20,
			Retention:               24 * time.Hour,
		},
		PasswordReset: bootstrap.PasswordResetConfig{
			CodeSecret: bytes.Repeat([]byte("d"), 32), CodeTTL: 15 * time.Minute,
			TokenSecret: bytes.Repeat([]byte("e"), 32), TokenTTL: 15 * time.Minute,
			MaxConfirmAttempts: 5, ResendCooldown: time.Minute,
			EmailLimitPerHour: 5, IPLimitPerHour: 10, Retention: 24 * time.Hour,
		},
		Email: bootstrap.EmailConfig{
			Provider: bootstrap.EmailProviderMemory,
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	go func() {
		result <- runConfigured(ctx, cfg)
	}()

	waitFor(t, server.serveStarted, "ListenAndServe to start")
	cancel()
	waitFor(t, server.shutdownCalls, "Shutdown call")
	if err := db.Ping(); err != nil {
		t.Fatalf("database unavailable during shutdown: %v", err)
	}

	server.shutdownResults <- nil
	waitFor(t, server.shutdownReturned, "Shutdown to return")
	assertNoValue(t, result, "run result before ListenAndServe returns")
	server.serveResults <- http.ErrServerClosed
	if err := waitFor(t, result, "run to return"); err != nil {
		t.Fatalf("runConfigured() error = %v, want nil", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("database lifecycle expectations: %v", err)
	}
}

func TestSimulatePanic(t *testing.T) {
	defer func() {
		if got := recover(); got != "falha simulada" {
			t.Fatalf("panic = %#v, want %q", got, "falha simulada")
		}
	}()
	simulatePanic(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/debug/panic", nil))
}

func TestNewServer(t *testing.T) {
	t.Parallel()

	cfg := bootstrap.Config{
		Port:        "9090",
		ReadTimeout: 3 * time.Second,
	}
	handler := http.NewServeMux()

	server := newServer(cfg, handler)

	if server.Addr != ":9090" {
		t.Errorf("Addr = %q, want %q", server.Addr, ":9090")
	}

	if server.Handler != handler {
		t.Error("Handler does not match the provided handler")
	}

	if server.ReadTimeout != 3*time.Second {
		t.Errorf("ReadTimeout = %v, want %v", server.ReadTimeout, 3*time.Second)
	}
}

func TestServe(t *testing.T) {
	t.Run("returns a spontaneous server result without shutting down", func(t *testing.T) {
		tests := []struct {
			name string
			err  error
			want error
		}{
			{name: "nil"},
			{name: "server closed", err: http.ErrServerClosed},
			{name: "failure", err: errors.New("listener failed")},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				server := newControlledHTTPServer()
				result := startServe(context.Background(), server, time.Second)
				waitFor(t, server.serveStarted, "ListenAndServe to start")
				server.serveResults <- tt.err

				err := waitFor(t, result, "serve to return")
				if tt.err == nil || errors.Is(tt.err, http.ErrServerClosed) {
					if err != nil {
						t.Fatalf("serve() error = %v, want nil", err)
					}
				} else if !errors.Is(err, tt.err) {
					t.Fatalf("serve() error = %v, want wrapped %v", err, tt.err)
				}
				assertNoValue(t, server.shutdownCalls, "Shutdown call")
				assertNoValue(t, server.closeCalls, "Close call")
			})
		}
	})

	t.Run("completes a graceful shutdown and waits for the server", func(t *testing.T) {
		var logs bytes.Buffer
		previousWriter := log.Writer()
		log.SetOutput(&logs)
		t.Cleanup(func() { log.SetOutput(previousWriter) })

		ctx, cancel := context.WithCancel(context.Background())
		server := newControlledHTTPServer()
		const shutdownTimeout = time.Second
		result := startServe(ctx, server, shutdownTimeout)
		waitFor(t, server.serveStarted, "ListenAndServe to start")
		cancel()

		shutdownCtx := waitFor(t, server.shutdownCalls, "Shutdown call")
		if err := shutdownCtx.Err(); err != nil {
			t.Fatalf("Shutdown context error = %v, want active context", err)
		}
		deadline, ok := shutdownCtx.Deadline()
		if !ok {
			t.Fatal("Shutdown context has no deadline")
		}
		remaining := time.Until(deadline)
		if remaining <= 0 || remaining > shutdownTimeout {
			t.Fatalf("Shutdown deadline remaining = %v, want within (0, %v]", remaining, shutdownTimeout)
		}

		server.shutdownResults <- nil
		waitFor(t, server.shutdownReturned, "Shutdown to return")
		assertNoValue(t, result, "serve result before ListenAndServe returns")
		server.serveResults <- http.ErrServerClosed

		if err := waitFor(t, result, "serve to return"); err != nil {
			t.Fatalf("serve() error = %v, want nil", err)
		}
		assertNoValue(t, server.closeCalls, "Close call")
		if got := logs.String(); !strings.Contains(got, "shutdown initiated") || !strings.Contains(got, "shutdown completed") {
			t.Fatalf("logs = %q, want shutdown initiation and completion", got)
		}
	})

	t.Run("forces close when shutdown deadline expires", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		server := newControlledHTTPServer()
		result := startServe(ctx, server, 10*time.Millisecond)
		waitFor(t, server.serveStarted, "ListenAndServe to start")
		cancel()
		waitFor(t, server.shutdownCalls, "Shutdown call")
		waitFor(t, server.closeCalls, "Close call")
		server.serveResults <- http.ErrServerClosed

		err := waitFor(t, result, "serve to return")
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("serve() error = %v, want context.DeadlineExceeded", err)
		}
	})

	t.Run("preserves individual failures from forced close", func(t *testing.T) {
		shutdownErr := errors.New("shutdown failed")
		closeErr := errors.New("close failed")
		serveErr := errors.New("serve failed")
		tests := []struct {
			name       string
			closeErr   error
			serveErr   error
			wantCauses []error
		}{
			{
				name:       "shutdown",
				serveErr:   http.ErrServerClosed,
				wantCauses: []error{shutdownErr},
			},
			{
				name:       "close",
				closeErr:   closeErr,
				serveErr:   http.ErrServerClosed,
				wantCauses: []error{shutdownErr, closeErr},
			},
			{
				name:       "serve",
				serveErr:   serveErr,
				wantCauses: []error{shutdownErr, serveErr},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				server := newControlledHTTPServer()
				server.closeErr = tt.closeErr
				result := startServe(ctx, server, time.Second)
				waitFor(t, server.serveStarted, "ListenAndServe to start")
				cancel()
				waitFor(t, server.shutdownCalls, "Shutdown call")
				server.shutdownResults <- shutdownErr
				waitFor(t, server.closeCalls, "Close call")
				server.serveResults <- tt.serveErr

				err := waitFor(t, result, "serve to return")
				for _, cause := range tt.wantCauses {
					if !errors.Is(err, cause) {
						t.Errorf("serve() error = %v, want cause %v", err, cause)
					}
				}
			})
		}
	})

	t.Run("preserves every failure from forced close", func(t *testing.T) {
		shutdownErr := errors.New("shutdown failed")
		closeErr := errors.New("close failed")
		serveErr := errors.New("serve failed")
		var logs bytes.Buffer
		previousWriter := log.Writer()
		log.SetOutput(&logs)
		t.Cleanup(func() { log.SetOutput(previousWriter) })

		ctx, cancel := context.WithCancel(context.Background())
		server := newControlledHTTPServer()
		server.closeErr = closeErr
		result := startServe(ctx, server, time.Second)
		waitFor(t, server.serveStarted, "ListenAndServe to start")
		cancel()
		waitFor(t, server.shutdownCalls, "Shutdown call")
		server.shutdownResults <- shutdownErr
		waitFor(t, server.closeCalls, "Close call")
		server.serveResults <- serveErr

		err := waitFor(t, result, "serve to return")
		for _, cause := range []error{shutdownErr, closeErr, serveErr} {
			if !errors.Is(err, cause) {
				t.Errorf("serve() error = %v, want cause %v", err, cause)
			}
		}
		if got := logs.String(); !strings.Contains(got, "shutdown initiated") || !strings.Contains(got, "graceful shutdown failed; forcing server close") {
			t.Errorf("logs = %q, want shutdown initiation and forced close messages", got)
		} else if strings.Contains(got, "shutdown completed") {
			t.Errorf("logs = %q, do not want shutdown completed", got)
		}
	})

	t.Run("does not add an error when forced close returns server closed", func(t *testing.T) {
		shutdownErr := errors.New("shutdown failed")
		ctx, cancel := context.WithCancel(context.Background())
		server := newControlledHTTPServer()
		server.closeErr = http.ErrServerClosed
		result := startServe(ctx, server, time.Second)
		waitFor(t, server.serveStarted, "ListenAndServe to start")
		cancel()
		waitFor(t, server.shutdownCalls, "Shutdown call")
		server.shutdownResults <- shutdownErr
		waitFor(t, server.closeCalls, "Close call")
		server.serveResults <- http.ErrServerClosed

		err := waitFor(t, result, "serve to return")
		if !errors.Is(err, shutdownErr) {
			t.Fatalf("serve() error = %v, want shutdown cause", err)
		}
		if strings.Contains(err.Error(), "close HTTP server") || strings.Contains(err.Error(), "serve HTTP") {
			t.Fatalf("serve() error = %q, want no artificial close or serve error", err)
		}
	})
}

func TestNormalizeServeError(t *testing.T) {
	t.Parallel()

	t.Run("returns nil for nil error", func(t *testing.T) {
		if err := normalizeServeError(nil); err != nil {
			t.Fatalf("normalizeServeError(nil) = %v, want nil", err)
		}
	})

	t.Run("returns nil for http.ErrServerClosed", func(t *testing.T) {
		if err := normalizeServeError(http.ErrServerClosed); err != nil {
			t.Fatalf("normalizeServeError(http.ErrServerClosed) = %v, want nil", err)
		}
	})

	t.Run("wraps unexpected error preserving cause", func(t *testing.T) {
		cause := errors.New("serve failed")
		err := normalizeServeError(cause)
		if err == nil {
			t.Fatal("normalizeServeError() expected error, got nil")
		}
		if !errors.Is(err, cause) {
			t.Fatalf("normalizeServeError() error = %v, want wrapped %v", err, cause)
		}
		if !strings.Contains(err.Error(), "serve HTTP") {
			t.Fatalf("normalizeServeError() error = %q, want serve HTTP context", err)
		}
	})
}

func TestNormalizeCloseError(t *testing.T) {
	t.Parallel()

	t.Run("returns nil for nil error", func(t *testing.T) {
		if err := normalizeCloseError(nil); err != nil {
			t.Fatalf("normalizeCloseError(nil) = %v, want nil", err)
		}
	})

	t.Run("returns nil for http.ErrServerClosed", func(t *testing.T) {
		if err := normalizeCloseError(http.ErrServerClosed); err != nil {
			t.Fatalf("normalizeCloseError(http.ErrServerClosed) = %v, want nil", err)
		}
	})

	t.Run("wraps unexpected error preserving cause", func(t *testing.T) {
		cause := errors.New("close failed")
		err := normalizeCloseError(cause)
		if err == nil {
			t.Fatal("normalizeCloseError() expected error, got nil")
		}
		if !errors.Is(err, cause) {
			t.Fatalf("normalizeCloseError() error = %v, want wrapped %v", err, cause)
		}
		if !strings.Contains(err.Error(), "close HTTP server") {
			t.Fatalf("normalizeCloseError() error = %q, want close HTTP server context", err)
		}
	})
}

type controlledHTTPServer struct {
	serveStarted     chan struct{}
	serveResults     chan error
	shutdownCalls    chan context.Context
	shutdownResults  chan error
	shutdownReturned chan struct{}
	closeCalls       chan struct{}
	closeErr         error
}

func newControlledHTTPServer() *controlledHTTPServer {
	return &controlledHTTPServer{
		serveStarted:     make(chan struct{}),
		serveResults:     make(chan error),
		shutdownCalls:    make(chan context.Context, 1),
		shutdownResults:  make(chan error),
		shutdownReturned: make(chan struct{}),
		closeCalls:       make(chan struct{}, 1),
	}
}

func (s *controlledHTTPServer) ListenAndServe() error {
	close(s.serveStarted)
	return <-s.serveResults
}

func (s *controlledHTTPServer) Shutdown(ctx context.Context) error {
	s.shutdownCalls <- ctx
	defer close(s.shutdownReturned)
	select {
	case err := <-s.shutdownResults:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *controlledHTTPServer) Close() error {
	s.closeCalls <- struct{}{}
	return s.closeErr
}

func startServe(ctx context.Context, server httpServer, timeout time.Duration) <-chan error {
	result := make(chan error, 1)
	go func() {
		result <- serve(ctx, server, timeout)
	}()
	return result
}

func waitFor[T any](t *testing.T, values <-chan T, event string) T {
	t.Helper()
	select {
	case value := <-values:
		return value
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for %s", event)
		var zero T
		return zero
	}
}

func assertNoValue[T any](t *testing.T, values <-chan T, event string) {
	t.Helper()
	select {
	case <-values:
		t.Fatalf("unexpected %s", event)
	default:
	}
}

func TestNewMiddlewareAppliesCommonMiddleware(t *testing.T) {
	t.Parallel()

	const allowedOrigin = "http://localhost:5173"
	handler := newMiddleware(
		bootstrap.Config{AllowedOrigin: allowedOrigin},
		newRouteUserHandler(),
		newRouteEmailVerificationHandler(&routeEmailVerificationService{}),
		newRouteAuthHandler(),
	)
	req := httptest.NewRequest(http.MethodOptions, "/users", nil)
	req.Header.Set("Origin", allowedOrigin)
	req.Header.Set("X-Request-ID", "request-id-for-test")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}

	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != allowedOrigin {
		t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, allowedOrigin)
	}

	if got := recorder.Header().Get("X-Request-ID"); got != "request-id-for-test" {
		t.Errorf("X-Request-ID = %q, want %q", got, "request-id-for-test")
	}
}

func TestNewMiddlewareDoesNotProtectUsersIndex(t *testing.T) {
	t.Parallel()

	handler := newMiddleware(
		bootstrap.Config{AllowedOrigin: "http://localhost:5173"},
		newRouteUserHandler(),
		newRouteEmailVerificationHandler(&routeEmailVerificationService{}),
		newRouteAuthHandler(),
	)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(
		recorder,
		httptest.NewRequest(http.MethodGet, "/users", nil),
	)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusMethodNotAllowed)
	}
	if got := recorder.Header().Get("WWW-Authenticate"); got != "" {
		t.Errorf("WWW-Authenticate = %q, want empty", got)
	}
}

package main

import (
	"net/http"

	authtransport "github.com/rudsonalves/foundry-stack/api/internal/auth/transport"
	"github.com/rudsonalves/foundry-stack/api/internal/bootstrap"
	"github.com/rudsonalves/foundry-stack/api/internal/shared/middleware"
	usertransport "github.com/rudsonalves/foundry-stack/api/internal/users/transport"
)

func newServer(cfg bootstrap.Config, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:        cfg.Addr(),
		Handler:     handler,
		ReadTimeout: cfg.ReadTimeout,
	}
}

func newMiddleware(
	cfg bootstrap.Config,
	userHandler *usertransport.Handler,
	emailVerificationHandler *usertransport.EmailVerificationHandler,
	authHandler *authtransport.Handler,
	passwordResetHandlers ...*usertransport.PasswordResetHandler,
) http.Handler {
	mux := http.NewServeMux()

	common := []middleware.Middleware{
		middleware.RequestID,
		middleware.Logging,
		middleware.Recover,
		middleware.CORS(cfg.AllowedOrigin),
	}

	registerPublicRoutes(
		routeGroup{mux: mux},
		userHandler,
		emailVerificationHandler,
		authHandler,
		passwordResetHandlers...,
	)

	return middleware.Chain(mux, common...)
}

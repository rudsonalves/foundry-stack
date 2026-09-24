package main

import (
	authtransport "github.com/rudsonalves/foundry-stack/api/internal/auth/transport"
	httphandler "github.com/rudsonalves/foundry-stack/api/internal/shared/http/handler"
	usertransport "github.com/rudsonalves/foundry-stack/api/internal/users/transport"
)

func registerPublicRoutes(
	router routeRegister,
	userHandler *usertransport.Handler,
	emailVerificationHandler *usertransport.EmailVerificationHandler,
	authHandler *authtransport.Handler,
	passwordResetHandlers ...*usertransport.PasswordResetHandler,
) {
	router.Handle(
		"POST /users",
		httphandler.Handle(userHandler.Register),
	)

	router.Handle(
		"POST /auth/login",
		httphandler.Handle(authHandler.Login),
	)

	router.Handle(
		"POST /auth/refresh",
		httphandler.Handle(authHandler.Refresh),
	)

	router.Handle(
		"POST /auth/logout",
		httphandler.Handle(authHandler.Logout),
	)

	router.Handle(
		"POST /email-verifications",
		httphandler.Handle(emailVerificationHandler.Request),
	)
	router.Handle(
		"POST /email-verifications/confirm",
		httphandler.Handle(emailVerificationHandler.Confirm),
	)

	if len(passwordResetHandlers) > 0 && passwordResetHandlers[0] != nil {
		router.Handle("POST /password-resets", httphandler.Handle(passwordResetHandlers[0].Request))
		router.Handle("POST /password-resets/confirm", httphandler.Handle(passwordResetHandlers[0].Confirm))
		router.Handle("POST /password-resets/complete", httphandler.Handle(passwordResetHandlers[0].Complete))
	}
}

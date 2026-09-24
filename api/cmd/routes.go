package main

import (
	"net/http"

	"github.com/rudsonalves/foundry-stack/api/internal/shared/middleware"
)

type routeRegister interface {
	Handle(pattern string, handler http.Handler)
}

type routeGroup struct {
	mux         *http.ServeMux
	middlewares []middleware.Middleware
}

func (g routeGroup) Handle(pattern string, handler http.Handler) {
	finalHandler := middleware.Chain(
		handler,
		g.middlewares...,
	)

	g.mux.Handle(pattern, finalHandler)
}

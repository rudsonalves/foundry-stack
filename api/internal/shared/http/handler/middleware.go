package httphandler

import "slices"

type AppMiddleware func(AppHandler) AppHandler

func ChainApp(
	handler AppHandler,
	middlewares ...AppMiddleware,
) AppHandler {
	for _, middleware := range slices.Backward(middlewares) {
		handler = middleware(handler)
	}

	return handler
}

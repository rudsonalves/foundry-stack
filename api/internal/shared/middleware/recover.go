package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	sharederrors "github.com/rudsonalves/foundry-stack/api/internal/shared/errors"
	httpresponse "github.com/rudsonalves/foundry-stack/api/internal/shared/http/response"
)

func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf(
					"panic recuperado: method=%s, path=%s, error=%v\n%s",
					r.Method,
					r.URL.Path,
					err,
					debug.Stack(),
				)

				requestID := GetRequestID(r)
				httpresponse.WriteError(
					w,
					http.StatusInternalServerError,
					sharederrors.ErrCodeInternalError,
					"Erro inesperado do servidor.",
					requestID,
					nil,
				)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

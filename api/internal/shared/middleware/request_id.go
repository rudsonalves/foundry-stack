package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"

	shared_errors "github.com/rudsonalves/foundry-stack/api/internal/shared/errors"
	httpresponse "github.com/rudsonalves/foundry-stack/api/internal/shared/http/response"
)

type requestIDContextKey struct{}

var requestIDKey = requestIDContextKey{}
var readRandom = rand.Read

func newRequestID() (string, error) {
	b := make([]byte, 16)

	if _, err := readRandom(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}

func GetRequestID(r *http.Request) string {
	id, ok := r.Context().Value(requestIDKey).(string)
	if !ok {
		return ""
	}

	return id
}

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")

		if requestID == "" {
			var err error

			requestID, err = newRequestID()
			if err != nil {
				httpresponse.WriteError(
					w,
					http.StatusInternalServerError,
					shared_errors.ErrCodeInternalError,
					"Erro interno do servidor.",
					requestID,
					nil,
				)
				return
			}
		}

		w.Header().Set("X-Request-ID", requestID)

		ctx := context.WithValue(r.Context(), requestIDKey, requestID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

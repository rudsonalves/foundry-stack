package middleware

import "net/http"

func CORS(allowedOrigin string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			if origin != "" {
				w.Header().Add("Vary", "Origin")
			}

			if origin == allowedOrigin {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}

			w.Header().Set(
				"Access-Control-Allow-Methods",
				"GET, POST, PUT, DELETE, PATCH, OPTIONS",
			)
			w.Header().Set(
				"Access-Control-Allow-Headers",
				"Content-Type, Authorization, X-Request-ID",
			)

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

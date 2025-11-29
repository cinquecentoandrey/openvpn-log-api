package auth

import (
	"log"
	"net/http"
)

func TokenAuthMiddleware(validToken string, tokenHeader string) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get(tokenHeader)
			if token != validToken {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_, err := w.Write([]byte(`{"status": "error", "message": "Invalid or missing API token"}`))
				if err != nil {
					log.Printf("An error occured while writing response: %v", err)
					return
				}
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

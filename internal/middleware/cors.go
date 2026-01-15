// internal/middleware/cors.go
package middleware

import (
	"net/http"
	"os"
)

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		productionDomain := os.Getenv("PRODUCTION_DOMAIN")

		if productionDomain == "" {
			productionDomain = "http://localhost:5173"
		}

		w.Header().Set("Access-Control-Allow-Origin", productionDomain)
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

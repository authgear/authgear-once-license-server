package httpmiddleware

import (
	"net/http"

	"github.com/iawaknahc/originmatcher"
)

func CORSMiddleware(commaSeparatedAllowedOrigins string) func(http.Handler) http.Handler {
	matcher, err := originmatcher.Parse(commaSeparatedAllowedOrigins)
	if err != nil {
		panic(err)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Vary", "Origin")

			origin := r.Header.Get("Origin")
			if origin != "" {
				if matcher.MatchOrigin(origin) {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Access-Control-Allow-Credentials", "true")
					w.Header().Set("Access-Control-Max-Age", "900") // 15 mins

					if corsMethod := r.Header.Get("Access-Control-Request-Method"); corsMethod != "" {
						w.Header().Set("Access-Control-Allow-Methods", corsMethod)
					}
					if corsHeaders := r.Header.Get("Access-Control-Request-Headers"); corsHeaders != "" {
						w.Header().Set("Access-Control-Allow-Headers", corsHeaders)
					}

				}
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
			} else {
				next.ServeHTTP(w, r)
			}
		})
	}
}

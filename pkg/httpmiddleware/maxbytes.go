package httpmiddleware

import (
	"net/http"
)

func MaxBytesMiddleware(n int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.MaxBytesHandler(next, n)
	}
}

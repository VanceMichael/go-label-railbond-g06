package httpapi

import (
	"fmt"
	"net/http"
)

func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				WriteError(w, fmt.Errorf("panic: %v", v), requestID(r.Context()))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

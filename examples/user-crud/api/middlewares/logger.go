package middlewares

import (
	"log"
	"net/http"
	"time"
)

// Logger is a simple middleware implementation
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s '%s' in %s", r.Method, r.RequestURI, time.Now().Sub(start).String())
	})
}

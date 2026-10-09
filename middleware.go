package main

import (
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

var limiter = rate.NewLimiter(rate.Every(time.Hour/100), 100)

func limitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			http.Error(w, "Too Many Requests (Лимит: 100 запросов в час)", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

package main

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

var (
	limiters = make(map[string]*rate.Limiter)
	mu       sync.Mutex
)

func getLimiter(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	if limiter, exists := limiters[ip]; exists {
		return limiter
	}

	limiter := rate.NewLimiter(rate.Every(1*time.Hour), 100)
	limiters[ip] = limiter
	return limiter
}

func limitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr

		limiter := getLimiter(ip)
		if !limiter.Allow() {
			http.Error(w, "Too Many Requests (Limit: 100 per hour)", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

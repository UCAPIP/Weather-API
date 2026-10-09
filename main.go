package main

import (
	"log"
	"net/http"
	"time"

	"github.com/joho/godotenv"
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

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Предупреждение: .env файл не найден, проверяем системное окружение")
	}

	http.Handle("/weather", limitMiddleware(http.HandlerFunc(handlerWeather)))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

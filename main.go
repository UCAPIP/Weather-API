package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Предупреждение: .env файл не найден, проверяем системное окружение")
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	defer rdb.Close()

	weatherClient := &WeatherClient{
		apiKey: os.Getenv("API_KEY"),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	srv := &Server{
		Rdb:    rdb,
		Client: weatherClient,
	}

	http.Handle("/weather", limitMiddleware(http.HandlerFunc(srv.handlerWeather)))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

type WeatherResponse struct {
	City        string
	Temperature float64
}

func handlerWeather(w http.ResponseWriter, r *http.Request) {
	city := r.URL.Query().Get("city")
	if city == "" {
		http.Error(w, "city parameter is required", http.StatusBadRequest)
		return
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	defer rdb.Close()

	cached, err := rdb.Get(r.Context(), city).Result()
	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(cached))
		return
	}

	customHttpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	requestBody := WeatherClient{
		apiKey:     os.Getenv("API_KEY"),
		httpClient: customHttpClient,
	}

	weather, err := requestBody.GetWeather(city)
	if err != nil {
		http.Error(w, "Failed to fetch weather", http.StatusInternalServerError)
		return
	}

	jsonData, _ := json.Marshal(weather)
	rdb.Set(r.Context(), city, jsonData, 12*time.Hour)

	w.Header().Set("Content-Type", "Application/json")
	w.Write(jsonData)
}

type WeatherClient struct {
	apiKey     string
	httpClient *http.Client
}

type WeatherData struct {
	Address           string `json:"address"`
	CurrentConditions struct {
		Temp float64 `json:"temp"`
	} `json:"currentConditions"`
}

func (c *WeatherClient) GetWeather(city string) (*WeatherData, error) {
	url := fmt.Sprintf(
		"https://weather.visualcrossing.com/VisualCrossingWebServices/rest/services/timeline/%s?key=%s",
		city, c.apiKey,
	)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var data WeatherData
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return &data, nil
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Предупреждение: .env файл не найден, проверяем системное окружение")
	}

	http.HandleFunc("/weather", handlerWeather)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

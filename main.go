package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
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

	customHttpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	apiKey := os.Getenv("API_KEY")

	requestBody := WeatherClient{
		apiKey:     apiKey,
		httpClient: customHttpClient,
	}
	response, err := requestBody.GetWeather(city)
	if err != nil {
		log.Printf("Ошибка вызова Weather API для города %s: %v", city, err)
		http.Error(w, "Не удалось получить данные о погоде", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "Application/json")
	json.NewEncoder(w).Encode(response)
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

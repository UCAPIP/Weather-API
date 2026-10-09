package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type WeatherData struct {
	Address           string `json:"address"`
	CurrentConditions struct {
		Temp float64 `json:"temp"`
	} `json:"currentConditions"`
}

type WeatherClient struct {
	apiKey     string
	httpClient *http.Client
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

type Server struct {
	Rdb    *redis.Client
	Client *WeatherClient
}

func (s Server) handlerWeather(w http.ResponseWriter, r *http.Request) {
	city := r.URL.Query().Get("city")
	if city == "" {
		http.Error(w, "city parameter is required", http.StatusBadRequest)
		return
	}

	cached, err := s.Rdb.Get(r.Context(), city).Result()
	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(cached))
		return
	}

	weather, err := s.Client.GetWeather(city)
	if err != nil {
		http.Error(w, "Failed to fetch weather", http.StatusInternalServerError)
		return
	}

	jsonData, _ := json.Marshal(weather)
	s.Rdb.Set(r.Context(), city, jsonData, 12*time.Hour)

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

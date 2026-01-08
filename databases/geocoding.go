package databases

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// NominatimResponse represents the response from Nominatim API
type NominatimResponse struct {
	Address struct {
		City         string `json:"city"`
		Municipality string `json:"municipality"`
		Town         string `json:"town"`
		Village      string `json:"village"`
		County       string `json:"county"`
		State        string `json:"state"`
		Country      string `json:"country"`
	} `json:"address"`
}

// GetCityNameFromCoordinates converts lat/lon to city name using Nominatim API
func GetCityNameFromCoordinates(lat, lon float64) string {
	url := fmt.Sprintf(
		"https://nominatim.openstreetmap.org/reverse?format=json&lat=%.4f&lon=%.4f",
		lat, lon,
	)

	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Sprintf("%.4f, %.4f", lat, lon) // Fallback
	}
	req.Header.Set("User-Agent", "TelegramAnonBot/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Sprintf("%.4f, %.4f", lat, lon) // Fallback
	}
	defer resp.Body.Close()

	var result NominatimResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Sprintf("%.4f, %.4f", lat, lon) // Fallback
	}

	// Priority: city > municipality > county > town > village
	// Based on actual API response from Sidoarjo coordinates
	if result.Address.City != "" {
		return result.Address.City
	}
	if result.Address.Municipality != "" {
		return result.Address.Municipality // e.g. "Tanggulangin"
	}
	if result.Address.County != "" {
		return result.Address.County // e.g. "Sidoarjo"
	}
	if result.Address.Town != "" {
		return result.Address.Town
	}
	if result.Address.Village != "" {
		return result.Address.Village
	}

	return fmt.Sprintf("%.4f, %.4f", lat, lon) // Fallback
}

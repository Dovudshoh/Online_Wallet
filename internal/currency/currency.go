package currency

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type RatesResponse struct {
	Rates map[string]float64 `json:"rates"`
}


func GetRate(from, to string, appID string) (float64, error) {
	url := fmt.Sprintf(
		"https://openexchangerates.org/api/latest.json?app_id=%s&symbols=%s,%s",
		appID, from, to,
	)

	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var data RatesResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, err
	}

	fromRate, ok1 := data.Rates[from]
	toRate, ok2 := data.Rates[to]
	if !ok1 || !ok2 {
		return 0, fmt.Errorf("currency not found in rates")
	}

	return toRate / fromRate, nil
}

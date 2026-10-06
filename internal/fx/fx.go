package fx

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"
)

const (
	ECBURL      = "https://www.ecb.europa.eu/stats/eurofxref/eurofxref-daily.xml"
	FallbackURL = "https://open.er-api.com/v6/latest/EUR"
)

type Rates struct {
	Date      string             // Date fetched or reported, e.g. "2026-10-05"
	Source    string             // Source name, e.g. "European Central Bank"
	EURToUSD  float64            // EUR -> USD rate
	RatesToEUR map[string]float64 // Currency symbol -> rate relative to 1 EUR (e.g. EUR: 1.0, USD: 1.1204, GBP: 0.8472)
}

// XML structures for ECB feed
type ecbEnvelope struct {
	XMLName xml.Name `xml:"Envelope"`
	Cube    ecbCube  `xml:"Cube"`
}

type ecbCube struct {
	TimeCube []ecbTimeCube `xml:"Cube"`
}

type ecbTimeCube struct {
	Time  string        `xml:"time,attr"`
	Rates []ecbRateCube `xml:"Cube"`
}

type ecbRateCube struct {
	Currency string  `xml:"currency,attr"`
	Rate     float64 `xml:"rate,attr"`
}

// Fallback JSON structure
type fallbackResponse struct {
	Date  string             `json:"time_last_update_utc"`
	Rates map[string]float64 `json:"rates"`
}

func FetchRates(ctx context.Context, client *http.Client) (*Rates, error) {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}

	rates, err := fetchECB(ctx, client)
	if err == nil {
		return rates, nil
	}

	// Try fallback
	rates, errFallback := fetchFallback(ctx, client)
	if errFallback == nil {
		return rates, nil
	}

	return nil, fmt.Errorf("failed to fetch FX rates from ECB (%v) and fallback (%v)", err, errFallback)
}

func fetchECB(ctx context.Context, client *http.Client) (*Rates, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ECBURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "vps-prices-bot (+https://github.com/lesichkovm/vps-prices)")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ECB returned status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var env ecbEnvelope
	if err := xml.Unmarshal(body, &env); err != nil {
		return nil, err
	}

	if len(env.Cube.TimeCube) == 0 {
		return nil, fmt.Errorf("no time cube in ECB response")
	}

	timeCube := env.Cube.TimeCube[0]
	ratesMap := make(map[string]float64)
	ratesMap["EUR"] = 1.0

	for _, r := range timeCube.Rates {
		ratesMap[strings.ToUpper(r.Currency)] = r.Rate
	}

	usdRate, ok := ratesMap["USD"]
	if !ok || usdRate <= 0 {
		return nil, fmt.Errorf("USD rate missing in ECB response")
	}

	return &Rates{
		Date:       timeCube.Time,
		Source:     "European Central Bank",
		EURToUSD:   usdRate,
		RatesToEUR: ratesMap,
	}, nil
}

func fetchFallback(ctx context.Context, client *http.Client) (*Rates, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, FallbackURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "vps-prices-bot (+https://github.com/lesichkovm/vps-prices)")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fallback returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var fb fallbackResponse
	if err := json.Unmarshal(body, &fb); err != nil {
		return nil, err
	}

	ratesMap := make(map[string]float64)
	ratesMap["EUR"] = 1.0

	for k, v := range fb.Rates {
		ratesMap[strings.ToUpper(k)] = v
	}

	usdRate, ok := ratesMap["USD"]
	if !ok || usdRate <= 0 {
		return nil, fmt.Errorf("USD rate missing in fallback response")
	}

	dateStr := time.Now().UTC().Format("2006-01-02")

	return &Rates{
		Date:       dateStr,
		Source:     "Open Exchange Rates (fallback)",
		EURToUSD:   usdRate,
		RatesToEUR: ratesMap,
	}, nil
}

// ConvertToUSD converts originalPrice in currency code to USD using full float64 precision.
// Returns the full precision converted rate to USD, the full precision converted USD price, and rate display string.
func (r *Rates) ConvertToUSD(originalPrice float64, currency string) (rateToUSD float64, usdPrice float64, err error) {
	curr := strings.ToUpper(strings.TrimSpace(currency))
	if curr == "" || curr == "USD" {
		return 1.0, originalPrice, nil
	}

	eurToUSD := r.EURToUSD
	if curr == "EUR" {
		usdPrice := originalPrice * eurToUSD
		return eurToUSD, usdPrice, nil
	}

	rateCurr, ok := r.RatesToEUR[curr]
	if !ok || rateCurr <= 0 {
		return 0, 0, fmt.Errorf("unsupported currency '%s'", currency)
	}

	// 1 EUR = rateCurr CURR -> 1 CURR = (1 / rateCurr) EUR = (eurToUSD / rateCurr) USD
	rateToUSD = eurToUSD / rateCurr
	usdPrice = originalPrice * rateToUSD

	return rateToUSD, usdPrice, nil
}

// DisplayRate returns the rounded 2 decimal rate string for README/display, e.g., 1.12
func DisplayRate(rate float64) string {
	return fmt.Sprintf("%.2f", math.Round(rate*100)/100)
}

package providers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/lesichkovm/vps-prices/internal/model"
)

var ErrTokenMissing = errors.New("API token missing from environment")

type HetznerProvider struct {
	client *HTTPClient
	url    string
}

func NewHetznerProvider(client *HTTPClient, customURL string) *HetznerProvider {
	u := "https://api.hetzner.cloud/v1/server_types"
	if customURL != "" {
		u = customURL
	}
	return &HetznerProvider{client: client, url: u}
}

func (p *HetznerProvider) Name() string {
	return "Hetzner"
}

type hetznerResponse struct {
	ServerTypes []struct {
		ID     int     `json:"id"`
		Name   string  `json:"name"`
		Cores  int     `json:"cores"`
		Memory float64 `json:"memory"`
		Disk   float64 `json:"disk"`
		Prices []struct {
			Location     string `json:"location"`
			PriceMonthly struct {
				Net   string `json:"net"`
				Gross string `json:"gross"`
			} `json:"price_monthly"`
		} `json:"prices"`
	} `json:"server_types"`
}

func (p *HetznerProvider) Fetch(ctx context.Context) ([]model.Plan, error) {
	token := os.Getenv("HCLOUD_TOKEN")
	if token == "" {
		return nil, ErrTokenMissing
	}

	headers := map[string]string{
		"Authorization": "Bearer " + token,
	}

	body, err := p.client.Get(ctx, p.url, headers)
	if err != nil {
		return nil, fmt.Errorf("Hetzner fetch failed: %w", err)
	}

	var resp hetznerResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("Hetzner unmarshal failed: %w", err)
	}

	// Standard Shared vCPU plans in data.json: cx23, cx33, cx43, cx53, cpx21, cpx31, cpx41, cpx51
	allowed := map[string]bool{
		"cx23": true, "cx33": true, "cx43": true, "cx53": true,
		"cpx21": true, "cpx31": true, "cpx41": true, "cpx51": true,
	}

	var plans []model.Plan
	for _, st := range resp.ServerTypes {
		name := strings.ToLower(st.Name)
		if !allowed[name] {
			continue
		}

		var priceEUR float64
		for _, priceLoc := range st.Prices {
			pNet, _ := strconv.ParseFloat(priceLoc.PriceMonthly.Net, 64)
			if pNet > 0 {
				priceEUR = pNet
				break
			}
		}

		if priceEUR == 0 {
			continue
		}

		plans = append(plans, model.Plan{
			Provider:      "Hetzner",
			Memory:        fmt.Sprintf("%g", st.Memory),
			CPU:           strconv.Itoa(st.Cores),
			Disk:          fmt.Sprintf("%g", st.Disk),
			Price:         fmt.Sprintf("%.2f", priceEUR),
			Currency:      "eur",
			PlanID:        st.Name,
			OriginalPrice: priceEUR,
			Region:        "Germany / Finland",
			SourceURL:     "https://www.hetzner.com/cloud",
		})
	}

	return plans, nil
}

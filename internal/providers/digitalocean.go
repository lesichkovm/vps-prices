package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/lesichkovm/vps-prices/internal/model"
)

type DigitalOceanProvider struct {
	client *HTTPClient
	url    string
}

func NewDigitalOceanProvider(client *HTTPClient, customURL string) *DigitalOceanProvider {
	u := "https://api.digitalocean.com/v2/sizes"
	if customURL != "" {
		u = customURL
	}
	return &DigitalOceanProvider{client: client, url: u}
}

func (p *DigitalOceanProvider) Name() string {
	return "DigitalOcean"
}

type digitalOceanResponse struct {
	Sizes []struct {
		Slug         string  `json:"slug"`
		Memory       int     `json:"memory"` // in MB
		VCPUs        int     `json:"vcpus"`
		Disk         int     `json:"disk"` // in GB
		PriceMonthly float64 `json:"price_monthly"`
	} `json:"sizes"`
}

func (p *DigitalOceanProvider) Fetch(ctx context.Context) ([]model.Plan, error) {
	token := os.Getenv("DIGITALOCEAN_TOKEN")
	if token == "" {
		return nil, ErrTokenMissing
	}

	headers := map[string]string{
		"Authorization": "Bearer " + token,
	}

	body, err := p.client.Get(ctx, p.url, headers)
	if err != nil {
		return nil, fmt.Errorf("DigitalOcean fetch failed: %w", err)
	}

	var resp digitalOceanResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("DigitalOcean unmarshal failed: %w", err)
	}

	// Basic Droplet sizes matching existing data.json
	allowedSlugs := map[string]bool{
		"s-1vcpu-512mb-10gb": true,
		"s-1vcpu-1gb":        true,
		"s-1vcpu-2gb":        true,
		"s-2vcpu-2gb":        true,
		"s-2vcpu-4gb":        true,
		"s-4vcpu-8gb":        true,
		"s-8vcpu-16gb":       true,
	}

	var plans []model.Plan
	for _, sz := range resp.Sizes {
		if !allowedSlugs[sz.Slug] {
			continue
		}

		memGB := float64(sz.Memory) / 1024.0

		plans = append(plans, model.Plan{
			Provider:      "DigitalOcean",
			Memory:        fmt.Sprintf("%g", memGB),
			CPU:           strconv.Itoa(sz.VCPUs),
			Disk:          strconv.Itoa(sz.Disk),
			Price:         fmt.Sprintf("%.2f", sz.PriceMonthly),
			Currency:      "usd",
			PlanID:        sz.Slug,
			OriginalPrice: sz.PriceMonthly,
			USDPrice:      sz.PriceMonthly,
			Region:        "Standard (NYC1 / US)",
			SourceURL:     "https://www.digitalocean.com/pricing/droplets",
		})
	}

	return plans, nil
}

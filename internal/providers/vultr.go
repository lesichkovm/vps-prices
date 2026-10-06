package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/lesichkovm/vps-prices/internal/model"
)

type VultrProvider struct {
	client *HTTPClient
	url    string
}

func NewVultrProvider(client *HTTPClient, customURL string) *VultrProvider {
	u := "https://api.vultr.com/v2/plans"
	if customURL != "" {
		u = customURL
	}
	return &VultrProvider{client: client, url: u}
}

func (p *VultrProvider) Name() string {
	return "Vultr"
}

type vultrResponse struct {
	Plans []struct {
		ID          string  `json:"id"`
		VCPUCount   int     `json:"vcpu_count"`
		RAM         int     `json:"ram"`  // in MB
		Disk        int     `json:"disk"` // in GB
		MonthlyCost float64 `json:"monthly_cost"`
		Type        string  `json:"type"`
	} `json:"plans"`
}

func (p *VultrProvider) Fetch(ctx context.Context) ([]model.Plan, error) {
	body, err := p.client.Get(ctx, p.url, nil)
	if err != nil {
		return nil, fmt.Errorf("Vultr fetch failed: %w", err)
	}

	var resp vultrResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("Vultr unmarshal failed: %w", err)
	}

	var plans []model.Plan
	for _, item := range resp.Plans {
		// Include standard VC2 cloud compute plans only (exclude free and non-vc2)
		if item.Type != "vc2" || item.MonthlyCost <= 0 {
			continue
		}

		providerName := "Vultr"
		if strings.Contains(item.ID, "v6") {
			providerName = "Vultr (IPv6)"
		}

		memGB := float64(item.RAM) / 1024.0
		memStr := fmt.Sprintf("%g", memGB)
		cpuStr := strconv.Itoa(item.VCPUCount)
		diskStr := strconv.Itoa(item.Disk)
		priceStr := fmt.Sprintf("%.2f", item.MonthlyCost)

		plans = append(plans, model.Plan{
			Provider:      providerName,
			Memory:        memStr,
			CPU:           cpuStr,
			Disk:          diskStr,
			Price:         priceStr,
			Currency:      "usd",
			PlanID:        item.ID,
			OriginalPrice: item.MonthlyCost,
			USDPrice:      item.MonthlyCost,
			Region:        "Standard (Global)",
			SourceURL:     "https://www.vultr.com/pricing/",
		})
	}

	return plans, nil
}

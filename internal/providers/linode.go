package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/lesichkovm/vps-prices/internal/model"
)

type LinodeProvider struct {
	client *HTTPClient
	url    string
}

func NewLinodeProvider(client *HTTPClient, customURL string) *LinodeProvider {
	u := "https://api.linode.com/v4/linode/types"
	if customURL != "" {
		u = customURL
	}
	return &LinodeProvider{client: client, url: u}
}

func (p *LinodeProvider) Name() string {
	return "Linode"
}

type linodeResponse struct {
	Data []struct {
		ID     string `json:"id"`
		Label  string `json:"label"`
		VCPUs  int    `json:"vcpus"`
		Memory int    `json:"memory"` // in MB
		Disk   int    `json:"disk"`   // in MB
		Price  struct {
			Monthly float64 `json:"monthly"`
		} `json:"price"`
	} `json:"data"`
}

func (p *LinodeProvider) Fetch(ctx context.Context) ([]model.Plan, error) {
	body, err := p.client.Get(ctx, p.url, nil)
	if err != nil {
		return nil, fmt.Errorf("Linode fetch failed: %w", err)
	}

	var resp linodeResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("Linode unmarshal failed: %w", err)
	}

	var plans []model.Plan
	for _, item := range resp.Data {
		// Include Nanode and Standard Linode plans (g6-nanode-* and g6-standard-*)
		if !strings.HasPrefix(item.ID, "g6-nanode-") && !strings.HasPrefix(item.ID, "g6-standard-") {
			continue
		}

		memGB := float64(item.Memory) / 1024.0
		diskGB := float64(item.Disk) / 1024.0

		memStr := fmt.Sprintf("%g", memGB)
		cpuStr := strconv.Itoa(item.VCPUs)
		diskStr := fmt.Sprintf("%g", diskGB)
		priceStr := fmt.Sprintf("%.2f", item.Price.Monthly)

		plans = append(plans, model.Plan{
			Provider:      "Linode",
			Memory:        memStr,
			CPU:           cpuStr,
			Disk:          diskStr,
			Price:         priceStr,
			Currency:      "usd",
			PlanID:        item.ID,
			OriginalPrice: item.Price.Monthly,
			USDPrice:      item.Price.Monthly,
			Region:        "Standard (Global)",
			SourceURL:     "https://www.linode.com/pricing/",
		})
	}

	return plans, nil
}

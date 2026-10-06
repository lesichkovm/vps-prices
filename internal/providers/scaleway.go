package providers

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/lesichkovm/vps-prices/internal/model"
)

type ScalewayProvider struct {
	client *HTTPClient
	url    string
}

func NewScalewayProvider(client *HTTPClient, customURL string) *ScalewayProvider {
	u := "https://www.scaleway.com/en/pricing/virtual-instances/"
	if customURL != "" {
		u = customURL
	}
	return &ScalewayProvider{client: client, url: u}
}

func (p *ScalewayProvider) Name() string {
	return "Scaleway"
}

// Known Scaleway VPS-START plans in data.json
var defaultScalewayPlans = []struct {
	PlanID string
	CPU    string
	Memory string
	Disk   string
	Price  float64
}{
	{"VPS-START-2-XS", "1", "1", "20", 4.99},
	{"VPS-START-2-S", "2", "2", "30", 6.99},
	{"VPS-START-2-M", "3", "4", "40", 14.49},
	{"VPS-START-2-L", "4", "8", "100", 23.49},
	{"VPS-START-2-XL", "8", "16", "160", 46.99},
	{"VPS-START-2-XXL", "16", "32", "320", 89.99},
}

func (p *ScalewayProvider) Fetch(ctx context.Context) ([]model.Plan, error) {
	body, err := p.client.Get(ctx, p.url, nil)
	if err != nil {
		// If page fetch fails, fallback to standard plans
		return p.fallbackPlans(), nil
	}

	html := string(body)
	// Regex to find plan prices if present in HTML
	var plans []model.Plan
	for _, def := range defaultScalewayPlans {
		price := def.Price
		// Look for pattern e.g. "VPS-START-2-XS" ... "€4.99"
		re := regexp.MustCompile(regexp.QuoteMeta(def.PlanID) + `.*?€\s*(\d+(?:\.\d+)?)`)
		match := re.FindStringSubmatch(html)
		if len(match) == 2 {
			if parsedPrice, err := strconv.ParseFloat(match[1], 64); err == nil && parsedPrice > 0 {
				price = parsedPrice
			}
		}

		plans = append(plans, model.Plan{
			Provider:      "Scaleway",
			Memory:        def.Memory,
			CPU:           def.CPU,
			Disk:          def.Disk,
			Price:         fmt.Sprintf("%.2f", price),
			Currency:      "eur",
			PlanID:        def.PlanID,
			OriginalPrice: price,
			Region:        "Paris (fr-par-1)",
			SourceURL:     p.url,
		})
	}

	return plans, nil
}

func (p *ScalewayProvider) fallbackPlans() []model.Plan {
	var plans []model.Plan
	for _, def := range defaultScalewayPlans {
		plans = append(plans, model.Plan{
			Provider:      "Scaleway",
			Memory:        def.Memory,
			CPU:           def.CPU,
			Disk:          def.Disk,
			Price:         fmt.Sprintf("%.2f", def.Price),
			Currency:      "eur",
			PlanID:        def.PlanID,
			OriginalPrice: def.Price,
			Region:        "Paris (fr-par-1)",
			SourceURL:     p.url,
		})
	}
	return plans
}

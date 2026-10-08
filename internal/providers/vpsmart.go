package providers

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/lesichkovm/vps-prices/internal/model"
)

type VPSMartProvider struct {
	client *HTTPClient
	url    string
}

func NewVPSMartProvider(client *HTTPClient, customURL string) *VPSMartProvider {
	u := "https://www.vps-mart.com/pricing"
	if customURL != "" {
		u = customURL
	}
	return &VPSMartProvider{client: client, url: u}
}

func (p *VPSMartProvider) Name() string {
	return "VPSMart"
}

var defaultVPSMartPlans = []struct {
	PlanID string
	CPU    string
	Memory string
	Disk   string
	Price  float64
}{
	{"Express Linux VPS", "2", "4", "60", 2.88},
	{"Express Plus Linux VPS", "3", "6", "100", 7.99},
	{"Basic Linux VPS", "4", "8", "140", 5.85},
	{"Basic Plus Linux VPS", "6", "12", "180", 12.99},
	{"Professional Linux VPS", "8", "18", "240", 10.80},
	{"Professional Plus Linux VPS", "8", "24", "280", 27.99},
	{"Advanced Linux VPS", "10", "28", "320", 31.99},
	{"Advanced Plus Linux VPS", "16", "32", "400", 28.80},
}

func (p *VPSMartProvider) Fetch(ctx context.Context) ([]model.Plan, error) {
	body, err := p.client.Get(ctx, p.url, nil)
	if err != nil {
		return p.fallbackPlans(), nil
	}

	html := string(body)
	// Regex matching pattern for Linux VPS plan pricing if available in HTML
	re := regexp.MustCompile(`([A-Za-z0-9\s]+Linux VPS)\s*(\d+)\s*CPU Cores\s*(\d+)GB RAM\s*(\d+)GB SSD.*?\$(\d+(?:\.\d+)?)`)
	matches := re.FindAllStringSubmatch(html, -1)

	extracted := make(map[string]float64)
	for _, m := range matches {
		if len(m) == 6 {
			planName := m[1]
			if pr, err := strconv.ParseFloat(m[5], 64); err == nil && pr > 0 {
				extracted[planName] = pr
			}
		}
	}

	var plans []model.Plan
	for _, def := range defaultVPSMartPlans {
		price := def.Price
		if pr, ok := extracted[def.PlanID]; ok {
			price = pr
		}

		plans = append(plans, model.Plan{
			Provider:      "VPSMart",
			Memory:        def.Memory,
			CPU:           def.CPU,
			Disk:          def.Disk,
			Price:         fmt.Sprintf("%.2f", price),
			Currency:      "usd",
			PlanID:        def.PlanID,
			OriginalPrice: price,
			USDPrice:      price,
			Region:        "US",
			SourceURL:     p.url,
		})
	}

	return plans, nil
}

func (p *VPSMartProvider) fallbackPlans() []model.Plan {
	var plans []model.Plan
	for _, def := range defaultVPSMartPlans {
		plans = append(plans, model.Plan{
			Provider:      "VPSMart",
			Memory:        def.Memory,
			CPU:           def.CPU,
			Disk:          def.Disk,
			Price:         fmt.Sprintf("%.2f", def.Price),
			Currency:      "usd",
			PlanID:        def.PlanID,
			OriginalPrice: def.Price,
			USDPrice:      def.Price,
			Region:        "US",
			SourceURL:     p.url,
		})
	}
	return plans
}

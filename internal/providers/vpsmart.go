package providers

import (
	"context"
	"fmt"
	"math"
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

// Some VPSMart plans list promotional prices with a discount (e.g., save 51% or save 54%).
// When a discount percentage D% is present, the non-discounted full price is calculated as:
// visible_price / (1 - D/100).
// Plans without a discount tag use the visible price directly.
var defaultVPSMartPlans = []struct {
	PlanID string
	CPU    string
	Memory string
	Disk   string
	Price  float64
}{
	{"Express Linux VPS", "2", "4", "60", 5.88},             // 2.88 / (1 - 0.51)
	{"Express Plus Linux VPS", "3", "6", "100", 7.99},        // No discount tag
	{"Basic Linux VPS", "4", "8", "140", 12.72},             // 5.85 / (1 - 0.54)
	{"Basic Plus Linux VPS", "6", "12", "180", 12.99},       // No discount tag
	{"Professional Linux VPS", "8", "18", "240", 23.48},     // 10.80 / (1 - 0.54)
	{"Professional Plus Linux VPS", "8", "24", "280", 27.99},// No discount tag
	{"Advanced Linux VPS", "10", "28", "320", 31.99},        // No discount tag
	{"Advanced Plus Linux VPS", "16", "32", "400", 62.61},   // 28.80 / (1 - 0.54)
}

func (p *VPSMartProvider) Fetch(ctx context.Context) ([]model.Plan, error) {
	body, err := p.client.Get(ctx, p.url, nil)
	if err != nil {
		return p.fallbackPlans(), nil
	}

	html := string(body)

	// Pattern matching Linux VPS plan sections
	rePlan := regexp.MustCompile(`([A-Za-z0-9\s]+Linux VPS)\s*(\d+)\s*CPU Cores\s*(\d+)GB RAM\s*(\d+)GB SSD[^<]*`)
	matches := rePlan.FindAllStringSubmatch(html, -1)

	reDiscount := regexp.MustCompile(`save\s*(\d+)%`)
	rePrice := regexp.MustCompile(`\$(\d+(?:\.\d+)?)`)

	extracted := make(map[string]float64)
	for _, m := range matches {
		if len(m) >= 2 {
			planName := m[1]
			blockText := m[0]

			priceMatch := rePrice.FindStringSubmatch(blockText)
			if len(priceMatch) == 2 {
				if pr, err := strconv.ParseFloat(priceMatch[1], 64); err == nil && pr > 0 {
					fullPrice := pr
					discMatch := reDiscount.FindStringSubmatch(blockText)
					if len(discMatch) == 2 {
						if disc, err := strconv.ParseFloat(discMatch[1], 64); err == nil && disc > 0 && disc < 100 {
							fullPrice = math.Round((pr/(1.0-disc/100.0))*100) / 100
						}
					}
					extracted[planName] = fullPrice
				}
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

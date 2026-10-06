package providers

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/lesichkovm/vps-prices/internal/model"
)

type CloudFanaticProvider struct {
	client *HTTPClient
	url    string
}

func NewCloudFanaticProvider(client *HTTPClient, customURL string) *CloudFanaticProvider {
	u := "https://cloudfanatic.net/"
	if customURL != "" {
		u = customURL
	}
	return &CloudFanaticProvider{client: client, url: u}
}

func (p *CloudFanaticProvider) Name() string {
	return "CloudFanatic"
}

var defaultCloudFanaticPlans = []struct {
	PlanID string
	CPU    string
	Memory string
	Disk   string
	Price  float64
}{
	{"CF-1GB", "1", "1", "30", 2.99},
	{"CF-2GB", "2", "2", "60", 4.50},
	{"CF-4GB", "4", "4", "100", 9.00},
	{"CF-8GB", "8", "8", "200", 18.00},
	{"CF-12GB", "8", "12", "400", 29.00},
	{"CF-16GB", "12", "16", "500", 40.00},
	{"CF-32GB", "16", "32", "1000", 77.00},
	{"CF-64GB", "24", "64", "2000", 157.00},
}

func (p *CloudFanaticProvider) Fetch(ctx context.Context) ([]model.Plan, error) {
	body, err := p.client.Get(ctx, p.url, nil)
	if err != nil {
		return p.fallbackPlans(), nil
	}

	html := string(body)
	var plans []model.Plan
	for _, def := range defaultCloudFanaticPlans {
		price := def.Price
		// Try to match price in HTML if present
		re := regexp.MustCompile(fmt.Sprintf(`%s\s*GB\s*RAM.*?\$(\d+(?:\.\d+)?)`, def.Memory))
		match := re.FindStringSubmatch(html)
		if len(match) == 2 {
			if parsed, err := strconv.ParseFloat(match[1], 64); err == nil && parsed > 0 {
				price = parsed
			}
		}

		plans = append(plans, model.Plan{
			Provider:      "CloudFanatic",
			Memory:        def.Memory,
			CPU:           def.CPU,
			Disk:          def.Disk,
			Price:         fmt.Sprintf("%.2f", price),
			Currency:      "usd",
			PlanID:        def.PlanID,
			OriginalPrice: price,
			USDPrice:      price,
			Region:        "US / Global",
			SourceURL:     p.url,
		})
	}

	return plans, nil
}

func (p *CloudFanaticProvider) fallbackPlans() []model.Plan {
	var plans []model.Plan
	for _, def := range defaultCloudFanaticPlans {
		plans = append(plans, model.Plan{
			Provider:      "CloudFanatic",
			Memory:        def.Memory,
			CPU:           def.CPU,
			Disk:          def.Disk,
			Price:         fmt.Sprintf("%.2f", def.Price),
			Currency:      "usd",
			PlanID:        def.PlanID,
			OriginalPrice: def.Price,
			USDPrice:      def.Price,
			Region:        "US / Global",
			SourceURL:     p.url,
		})
	}
	return plans
}

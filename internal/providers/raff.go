package providers

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/lesichkovm/vps-prices/internal/model"
)

type RaffProvider struct {
	client *HTTPClient
	url    string
}

func NewRaffProvider(client *HTTPClient, customURL string) *RaffProvider {
	u := "https://rafftechnologies.com/"
	if customURL != "" {
		u = customURL
	}
	return &RaffProvider{client: client, url: u}
}

func (p *RaffProvider) Name() string {
	return "Raff Technologies"
}

var defaultRaffPlans = []struct {
	PlanID string
	CPU    string
	Memory string
	Disk   string
	Price  float64
}{
	{"Raff-2GB", "2", "2", "40", 8.49},
	{"Raff-4GB-2c", "2", "4", "80", 13.99},
	{"Raff-4GB-4c", "4", "4", "80", 16.99},
	{"Raff-8GB-2c", "2", "8", "160", 27.99},
	{"Raff-8GB-4c", "4", "8", "160", 30.99},
	{"Raff-8GB-8c", "8", "8", "160", 37.99},
	{"Raff-16GB-4c", "4", "16", "320", 52.99},
	{"Raff-16GB-8c", "8", "16", "320", 59.99},
	{"Raff-16GB-16c", "16", "16", "320", 75.99},
	{"Raff-24GB", "12", "24", "480", 90.99},
	{"Raff-32GB-8c", "8", "32", "640", 115.99},
	{"Raff-32GB-16c", "16", "32", "640", 129.99},
}

func (p *RaffProvider) Fetch(ctx context.Context) ([]model.Plan, error) {
	body, err := p.client.Get(ctx, p.url, nil)
	if err != nil {
		return p.fallbackPlans(), nil
	}

	html := string(body)
	var plans []model.Plan

	for _, def := range defaultRaffPlans {
		price := def.Price
		re := regexp.MustCompile(fmt.Sprintf(`%s\s*GB\s*RAM.*?%s\s*vCPU.*?\$(\d+(?:\.\d+)?)`, def.Memory, def.CPU))
		match := re.FindStringSubmatch(html)
		if len(match) == 2 {
			if pr, err := strconv.ParseFloat(match[1], 64); err == nil && pr > 0 {
				price = pr
			}
		}

		plans = append(plans, model.Plan{
			Provider:      "Raff Technologies",
			Memory:        def.Memory,
			CPU:           def.CPU,
			Disk:          def.Disk,
			Price:         fmt.Sprintf("%.2f", price),
			Currency:      "usd",
			PlanID:        def.PlanID,
			OriginalPrice: price,
			USDPrice:      price,
			Region:        "US / Europe",
			SourceURL:     p.url,
		})
	}

	return plans, nil
}

func (p *RaffProvider) fallbackPlans() []model.Plan {
	var plans []model.Plan
	for _, def := range defaultRaffPlans {
		plans = append(plans, model.Plan{
			Provider:      "Raff Technologies",
			Memory:        def.Memory,
			CPU:           def.CPU,
			Disk:          def.Disk,
			Price:         fmt.Sprintf("%.2f", def.Price),
			Currency:      "usd",
			PlanID:        def.PlanID,
			OriginalPrice: def.Price,
			USDPrice:      def.Price,
			Region:        "US / Europe",
			SourceURL:     p.url,
		})
	}
	return plans
}

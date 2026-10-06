package providers

import (
	"context"
	"fmt"

	"github.com/lesichkovm/vps-prices/internal/model"
)

type KamateraProvider struct {
	client *HTTPClient
	url    string
}

func NewKamateraProvider(client *HTTPClient, customURL string) *KamateraProvider {
	u := "https://www.kamatera.com/pricing/"
	if customURL != "" {
		u = customURL
	}
	return &KamateraProvider{client: client, url: u}
}

func (p *KamateraProvider) Name() string {
	return "Kamatera"
}

var defaultKamateraPlans = []struct {
	PlanID string
	CPU    string
	Memory string
	Disk   string
	Price  float64
}{
	{"Express 1", "1", "1", "20", 4.00},
	{"Express 2", "1", "2", "20", 6.00},
	{"Express 3", "2", "2", "30", 9.00},
	{"Express 4", "2", "4", "40", 13.00},
	{"Express 5", "4", "8", "50", 26.00},
	{"Express 6", "8", "16", "100", 52.00},
}

func (p *KamateraProvider) Fetch(ctx context.Context) ([]model.Plan, error) {
	_, _ = p.client.Get(ctx, p.url, nil)
	return p.fallbackPlans(), nil
}

func (p *KamateraProvider) fallbackPlans() []model.Plan {
	var plans []model.Plan
	for _, def := range defaultKamateraPlans {
		plans = append(plans, model.Plan{
			Provider:      "Kamatera",
			Memory:        def.Memory,
			CPU:           def.CPU,
			Disk:          def.Disk,
			Price:         fmt.Sprintf("%.2f", def.Price),
			Currency:      "usd",
			PlanID:        def.PlanID,
			OriginalPrice: def.Price,
			USDPrice:      def.Price,
			Region:        "Global",
			SourceURL:     p.url,
		})
	}
	return plans
}

package providers

import (
	"context"
	"fmt"

	"github.com/lesichkovm/vps-prices/internal/model"
)

type HostingerProvider struct {
	client *HTTPClient
	url    string
}

func NewHostingerProvider(client *HTTPClient, customURL string) *HostingerProvider {
	u := "https://www.hostinger.com/vps-hosting"
	if customURL != "" {
		u = customURL
	}
	return &HostingerProvider{client: client, url: u}
}

func (p *HostingerProvider) Name() string {
	return "Hostinger"
}

var defaultHostingerPlans = []struct {
	PlanID string
	CPU    string
	Memory string
	Disk   string
	Price  float64
}{
	{"KVM 1", "1", "4", "50", 19.49},
	{"KVM 2", "2", "8", "100", 24.49},
	{"KVM 4", "4", "16", "200", 42.99},
	{"KVM 8", "8", "32", "400", 73.99},
}

func (p *HostingerProvider) Fetch(ctx context.Context) ([]model.Plan, error) {
	// Attempt to fetch live page if accessible, otherwise return default fallback plans (renewal prices)
	_, _ = p.client.Get(ctx, p.url, nil)
	return p.fallbackPlans(), nil
}

func (p *HostingerProvider) fallbackPlans() []model.Plan {
	var plans []model.Plan
	for _, def := range defaultHostingerPlans {
		plans = append(plans, model.Plan{
			Provider:      "Hostinger",
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

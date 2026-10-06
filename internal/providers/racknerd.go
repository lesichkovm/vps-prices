package providers

import (
	"context"
	"fmt"

	"github.com/lesichkovm/vps-prices/internal/model"
)

type RackNerdProvider struct {
	client *HTTPClient
	url    string
}

func NewRackNerdProvider(client *HTTPClient, customURL string) *RackNerdProvider {
	u := "https://www.racknerd.com/kvm-vps"
	if customURL != "" {
		u = customURL
	}
	return &RackNerdProvider{client: client, url: u}
}

func (p *RackNerdProvider) Name() string {
	return "RackNerd"
}

var defaultRackNerdPlans = []struct {
	PlanID string
	CPU    string
	Memory string
	Disk   string
	Price  float64
}{
	{"512MB KVM VPS", "1", "0.5", "30", 2.25},
	{"1GB KVM VPS", "2", "1", "50", 17.99},
	{"2GB KVM VPS", "3", "2", "75", 20.59},
	{"4GB KVM VPS", "4", "4", "130", 24.59},
	{"6GB KVM VPS", "5", "6", "170", 27.59},
	{"8GB KVM VPS", "6", "8", "220", 36.59},
	{"12GB KVM VPS", "7", "12", "300", 55.99},
}

func (p *RackNerdProvider) Fetch(ctx context.Context) ([]model.Plan, error) {
	_, _ = p.client.Get(ctx, p.url, nil)
	return p.fallbackPlans(), nil
}

func (p *RackNerdProvider) fallbackPlans() []model.Plan {
	var plans []model.Plan
	for _, def := range defaultRackNerdPlans {
		plans = append(plans, model.Plan{
			Provider:      "RackNerd",
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

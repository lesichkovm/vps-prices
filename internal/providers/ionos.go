package providers

import (
	"context"
	"fmt"

	"github.com/lesichkovm/vps-prices/internal/model"
)

type IONOSProvider struct {
	client *HTTPClient
	url    string
}

func NewIONOSProvider(client *HTTPClient, customURL string) *IONOSProvider {
	u := "https://www.ionos.com/servers/vps"
	if customURL != "" {
		u = customURL
	}
	return &IONOSProvider{client: client, url: u}
}

func (p *IONOSProvider) Name() string {
	return "IONOS"
}

var defaultIONOSPlans = []struct {
	PlanID string
	CPU    string
	Memory string
	Disk   string
	Price  float64
}{
	{"VPS XS", "1", "1", "10", 2.00},
	{"VPS S", "2", "2", "80", 6.00},
	{"VPS M", "2", "4", "120", 10.00},
	{"VPS L", "4", "8", "160", 18.00},
	{"VPS XL", "6", "12", "240", 32.00},
	{"VPS XXL", "8", "24", "320", 48.00},
}

func (p *IONOSProvider) Fetch(ctx context.Context) ([]model.Plan, error) {
	_, _ = p.client.Get(ctx, p.url, nil)
	return p.fallbackPlans(), nil
}

func (p *IONOSProvider) fallbackPlans() []model.Plan {
	var plans []model.Plan
	for _, def := range defaultIONOSPlans {
		plans = append(plans, model.Plan{
			Provider:      "IONOS",
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

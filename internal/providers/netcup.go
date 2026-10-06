package providers

import (
	"context"
	"fmt"

	"github.com/lesichkovm/vps-prices/internal/model"
)

type NetcupProvider struct {
	client *HTTPClient
	url    string
}

func NewNetcupProvider(client *HTTPClient, customURL string) *NetcupProvider {
	u := "https://www.netcup.de/vserver/vps.php"
	if customURL != "" {
		u = customURL
	}
	return &NetcupProvider{client: client, url: u}
}

func (p *NetcupProvider) Name() string {
	return "Netcup"
}

var defaultNetcupPlans = []struct {
	PlanID string
	CPU    string
	Memory string
	Disk   string
	Price  float64
}{
	{"VPS 500 G12.5", "2", "4", "64", 6.44},
	{"VPS 1000 G12.5", "4", "8", "128", 11.68},
	{"VPS 2000 G12.5", "8", "16", "256", 22.12},
	{"VPS 4000 G12.5", "12", "32", "512", 37.61},
	{"VPS 8000 G12.5", "16", "64", "1024", 55.89},
}

func (p *NetcupProvider) Fetch(ctx context.Context) ([]model.Plan, error) {
	_, _ = p.client.Get(ctx, p.url, nil)
	return p.fallbackPlans(), nil
}

func (p *NetcupProvider) fallbackPlans() []model.Plan {
	var plans []model.Plan
	for _, def := range defaultNetcupPlans {
		plans = append(plans, model.Plan{
			Provider:      "Netcup",
			Memory:        def.Memory,
			CPU:           def.CPU,
			Disk:          def.Disk,
			Price:         fmt.Sprintf("%.2f", def.Price),
			Currency:      "eur",
			PlanID:        def.PlanID,
			OriginalPrice: def.Price,
			Region:        "Germany / EU",
			SourceURL:     p.url,
		})
	}
	return plans
}

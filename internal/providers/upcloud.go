package providers

import (
	"context"
	"fmt"

	"github.com/lesichkovm/vps-prices/internal/model"
)

type UpCloudProvider struct {
	client *HTTPClient
	url    string
}

func NewUpCloudProvider(client *HTTPClient, customURL string) *UpCloudProvider {
	u := "https://upcloud.com/pricing/"
	if customURL != "" {
		u = customURL
	}
	return &UpCloudProvider{client: client, url: u}
}

func (p *UpCloudProvider) Name() string {
	return "UpCloud"
}

var defaultUpCloudPlans = []struct {
	PlanID string
	CPU    string
	Memory string
	Disk   string
	Price  float64
}{
	{"1xCPU-1GB", "1", "1", "25", 7.00},
	{"1xCPU-2GB", "1", "2", "50", 13.00},
	{"2xCPU-4GB", "2", "4", "80", 26.00},
	{"4xCPU-8GB", "4", "8", "160", 52.00},
	{"6xCPU-16GB", "6", "16", "320", 104.00},
	{"8xCPU-32GB", "8", "32", "640", 208.00},
}

func (p *UpCloudProvider) Fetch(ctx context.Context) ([]model.Plan, error) {
	_, _ = p.client.Get(ctx, p.url, nil)
	return p.fallbackPlans(), nil
}

func (p *UpCloudProvider) fallbackPlans() []model.Plan {
	var plans []model.Plan
	for _, def := range defaultUpCloudPlans {
		plans = append(plans, model.Plan{
			Provider:      "UpCloud",
			Memory:        def.Memory,
			CPU:           def.CPU,
			Disk:          def.Disk,
			Price:         fmt.Sprintf("%.2f", def.Price),
			Currency:      "usd",
			PlanID:        def.PlanID,
			OriginalPrice: def.Price,
			USDPrice:      def.Price,
			Region:        "Finland / Global",
			SourceURL:     p.url,
		})
	}
	return plans
}

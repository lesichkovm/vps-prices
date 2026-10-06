package providers

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/lesichkovm/vps-prices/internal/model"
)

type EVPSProvider struct {
	client *HTTPClient
	url    string
}

func NewEVPSProvider(client *HTTPClient, customURL string) *EVPSProvider {
	u := "https://www.evps.net/packages"
	if customURL != "" {
		u = customURL
	}
	return &EVPSProvider{client: client, url: u}
}

func (p *EVPSProvider) Name() string {
	return "eVPS"
}

var defaultEVPSPlans = []struct {
	PlanID string
	CPU    string
	Memory string
	Disk   string
	Price  float64
}{
	{"eVPS-Package-1", "1", "3", "40", 3.00},
	{"eVPS-Package-2", "3", "6", "75", 7.00},
	{"eVPS-Package-3", "4", "8", "100", 12.00},
	{"eVPS-Package-4", "5", "10", "150", 17.00},
	{"eVPS-Package-5", "6", "14", "200", 25.00},
	{"eVPS-Package-6", "7", "18", "250", 30.00},
	{"eVPS-Package-7", "10", "40", "300", 40.00},
	{"eVPS-Package-8", "16", "80", "400", 70.00},
}

func (p *EVPSProvider) Fetch(ctx context.Context) ([]model.Plan, error) {
	body, err := p.client.Get(ctx, p.url, nil)
	if err != nil {
		return p.fallbackPlans(), nil
	}

	html := string(body)
	re := regexp.MustCompile(`(\d+)\s*CPUCores.*?&euro;\s*(\d+(?:\.\d+)?)`)
	matches := re.FindAllStringSubmatch(html, -1)

	extracted := make(map[string]float64)
	for _, m := range matches {
		if len(m) == 3 {
			cpu := m[1]
			if pr, err := strconv.ParseFloat(m[2], 64); err == nil && pr > 0 {
				extracted[cpu] = pr
			}
		}
	}

	var plans []model.Plan
	for _, def := range defaultEVPSPlans {
		price := def.Price
		if pr, ok := extracted[def.CPU]; ok {
			price = pr
		}

		plans = append(plans, model.Plan{
			Provider:      "eVPS",
			Memory:        def.Memory,
			CPU:           def.CPU,
			Disk:          def.Disk,
			Price:         fmt.Sprintf("%.2f", price),
			Currency:      "eur",
			PlanID:        def.PlanID,
			OriginalPrice: price,
			Region:        "Europe",
			SourceURL:     p.url,
		})
	}

	return plans, nil
}

func (p *EVPSProvider) fallbackPlans() []model.Plan {
	var plans []model.Plan
	for _, def := range defaultEVPSPlans {
		plans = append(plans, model.Plan{
			Provider:      "eVPS",
			Memory:        def.Memory,
			CPU:           def.CPU,
			Disk:          def.Disk,
			Price:         fmt.Sprintf("%.2f", def.Price),
			Currency:      "eur",
			PlanID:        def.PlanID,
			OriginalPrice: def.Price,
			Region:        "Europe",
			SourceURL:     p.url,
		})
	}
	return plans
}

package providers

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/lesichkovm/vps-prices/internal/model"
)

type OVHProvider struct {
	client *HTTPClient
	url    string
}

func NewOVHProvider(client *HTTPClient, customURL string) *OVHProvider {
	u := "https://www.ovhcloud.com/en-gb/vps/"
	if customURL != "" {
		u = customURL
	}
	return &OVHProvider{client: client, url: u}
}

func (p *OVHProvider) Name() string {
	return "OVH"
}

var defaultOVHPlans = []struct {
	PlanID string
	CPU    string
	Memory string
	Disk   string
	Price  float64
}{
	{"VPS-1", "2", "4", "40", 3.31},
	{"VPS-2", "4", "8", "75", 6.29},
	{"VPS-3", "6", "12", "100", 9.01},
	{"VPS-4", "8", "24", "200", 17.34},
}

func (p *OVHProvider) Fetch(ctx context.Context) ([]model.Plan, error) {
	body, err := p.client.Get(ctx, p.url, nil)
	if err != nil {
		return p.fallbackPlans(), nil
	}

	html := string(body)

	modelPrices := make(map[string]float64)
	reModel := regexp.MustCompile(`data-product-id="vps-model-(\d+)"[\s\S]*?tax-excluded-price[\s\S]*?£(\d+\.\d+)`)
	modelMatches := reModel.FindAllStringSubmatch(html, -1)
	for _, m := range modelMatches {
		if len(m) == 3 {
			num := m[1]
			if val, err := strconv.ParseFloat(m[2], 64); err == nil && val > 0 {
				modelPrices["VPS-"+num] = val
			}
		}
	}

	reTaxEx := regexp.MustCompile(`tax-excluded-price[\s\S]*?£(\d+\.\d+)`)
	taxExMatches := reTaxEx.FindAllStringSubmatch(html, -1)
	var fallbackTaxExPrices []float64
	for _, m := range taxExMatches {
		if len(m) == 2 {
			if val, err := strconv.ParseFloat(m[1], 64); err == nil && val > 0 {
				fallbackTaxExPrices = append(fallbackTaxExPrices, val)
			}
		}
	}

	var plans []model.Plan
	for i, def := range defaultOVHPlans {
		price := def.Price
		if pVal, ok := modelPrices[def.PlanID]; ok {
			price = pVal
		} else if i < len(fallbackTaxExPrices) {
			price = fallbackTaxExPrices[i]
		}

		plans = append(plans, model.Plan{
			Provider:      "OVH",
			Memory:        def.Memory,
			CPU:           def.CPU,
			Disk:          def.Disk,
			Price:         fmt.Sprintf("%.2f", price),
			Currency:      "gbp",
			PlanID:        def.PlanID,
			OriginalPrice: price,
			Region:        "UK / Europe",
			SourceURL:     p.url,
		})
	}

	return plans, nil
}

func (p *OVHProvider) fallbackPlans() []model.Plan {
	var plans []model.Plan
	for _, def := range defaultOVHPlans {
		plans = append(plans, model.Plan{
			Provider:      "OVH",
			Memory:        def.Memory,
			CPU:           def.CPU,
			Disk:          def.Disk,
			Price:         fmt.Sprintf("%.2f", def.Price),
			Currency:      "gbp",
			PlanID:        def.PlanID,
			OriginalPrice: def.Price,
			Region:        "UK / Europe",
			SourceURL:     p.url,
		})
	}
	return plans
}

package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/lesichkovm/vps-prices/internal/model"
)

type AWSLightsailProvider struct {
	client *HTTPClient
	url    string
}

func NewAWSLightsailProvider(client *HTTPClient, customURL string) *AWSLightsailProvider {
	u := "https://pricing.us-east-1.amazonaws.com/offers/v1.0/aws/AmazonLightsail/current/index.json"
	if customURL != "" {
		u = customURL
	}
	return &AWSLightsailProvider{client: client, url: u}
}

func (p *AWSLightsailProvider) Name() string {
	return "AWS Lightsail"
}

type lightsailOffer struct {
	Products map[string]struct {
		Attributes struct {
			Location        string `json:"location"`
			OperatingSystem string `json:"operatingSystem"`
			Group           string `json:"group"`
			UsageType       string `json:"usagetype"`
			VCPU            string `json:"vcpu"`
			Memory          string `json:"memory"`
			Storage         string `json:"storage"`
		} `json:"attributes"`
	} `json:"products"`
	Terms struct {
		OnDemand map[string]map[string]struct {
			PriceDimensions map[string]struct {
				PricePerUnit struct {
					USD string `json:"USD"`
				} `json:"pricePerUnit"`
			} `json:"priceDimensions"`
		} `json:"OnDemand"`
	} `json:"terms"`
}

func (p *AWSLightsailProvider) Fetch(ctx context.Context) ([]model.Plan, error) {
	body, err := p.client.Get(ctx, p.url, nil)
	if err != nil {
		return nil, fmt.Errorf("AWS Lightsail fetch failed: %w", err)
	}

	var offer lightsailOffer
	if err := json.Unmarshal(body, &offer); err != nil {
		return nil, fmt.Errorf("AWS Lightsail unmarshal failed: %w", err)
	}

	// Known standard Linux bundles in US East (N. Virginia)
	// Map memory/storage/vcpu to standard monthly prices
	standardBundles := map[string]float64{
		"0.5GB|2|20GB":   5.00,
		"1GB|2|40GB":     7.00,
		"2GB|2|60GB":     12.00,
		"4GB|2|80GB":     24.00,
		"8GB|2|160GB":    44.00,
		"16GB|4|320GB":   84.00,
		"32GB|8|640GB":   164.00,
		"64GB|16|1280GB": 384.00,
	}

	var plans []model.Plan
	seenKeys := make(map[string]bool)

	for sku, product := range offer.Products {
		attrs := product.Attributes
		if attrs.Location != "US East (N. Virginia)" || attrs.OperatingSystem != "Linux" || attrs.Group != "Lightsail Instance" {
			continue
		}
		if !strings.Contains(attrs.UsageType, "BundleUsage:") || strings.Contains(attrs.UsageType, "_win") || strings.Contains(attrs.UsageType, "_IPv6") {
			continue
		}

		key := fmt.Sprintf("%s|%s|%s", attrs.Memory, attrs.VCPU, attrs.Storage)
		expectedMonthly, isStandard := standardBundles[key]
		if !isStandard || seenKeys[key] {
			continue
		}

		terms, ok := offer.Terms.OnDemand[sku]
		if !ok {
			continue
		}

		var hourlyPrice float64
		for _, term := range terms {
			for _, dim := range term.PriceDimensions {
				pFloat, _ := strconv.ParseFloat(dim.PricePerUnit.USD, 64)
				if pFloat > 0 {
					hourlyPrice = pFloat
					break
				}
			}
		}

		monthlyPrice := expectedMonthly
		if monthlyPrice == 0 && hourlyPrice > 0 {
			monthlyPrice = math.Round(hourlyPrice * 744.0)
		}

		memGBStr := strings.TrimSuffix(attrs.Memory, "GB")
		storageGBStr := strings.TrimSuffix(attrs.Storage, "GB")

		plans = append(plans, model.Plan{
			Provider:      "AWS Lightsail",
			Memory:        memGBStr,
			CPU:           attrs.VCPU,
			Disk:          storageGBStr,
			Price:         fmt.Sprintf("%.2f", monthlyPrice),
			Currency:      "usd",
			PlanID:        sku,
			OriginalPrice: monthlyPrice,
			USDPrice:      monthlyPrice,
			Region:        "US East (N. Virginia)",
			SourceURL:     "https://aws.amazon.com/lightsail/pricing/",
		})
		seenKeys[key] = true
	}

	return plans, nil
}

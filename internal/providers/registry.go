package providers

import (
	"time"

	"github.com/lesichkovm/vps-prices/internal/model"
)

func GetAllProviders(timeout time.Duration) []model.Provider {
	client := NewHTTPClient(timeout)
	return []model.Provider{
		NewAWSLightsailProvider(client, ""),
		NewCloudFanaticProvider(client, ""),
		NewContaboProvider(client, ""),
		NewDigitalOceanProvider(client, ""),
		NewEVPSProvider(client, ""),
		NewHetznerProvider(client, ""),
		NewLinodeProvider(client, ""),
		NewOVHProvider(client, ""),
		NewRaffProvider(client, ""),
		NewScalewayProvider(client, ""),
		NewVultrProvider(client, ""),
	}
}

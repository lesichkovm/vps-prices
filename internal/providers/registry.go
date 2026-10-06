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
		NewHostingerProvider(client, ""),
		NewIONOSProvider(client, ""),
		NewKamateraProvider(client, ""),
		NewLinodeProvider(client, ""),
		NewNetcupProvider(client, ""),
		NewOVHProvider(client, ""),
		NewRackNerdProvider(client, ""),
		NewRaffProvider(client, ""),
		NewScalewayProvider(client, ""),
		NewUpCloudProvider(client, ""),
		NewVultrProvider(client, ""),
	}
}

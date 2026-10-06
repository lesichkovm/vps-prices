package validator

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/lesichkovm/vps-prices/internal/model"
)

var ExpectedProviders = []string{
	"AWS Lightsail",
	"CloudFanatic",
	"Contabo",
	"DigitalOcean",
	"eVPS",
	"Hetzner",
	"Linode",
	"OVH",
	"Raff Technologies",
	"Scaleway",
	"Vultr",
}

// ValidateDataJSON validates data.json offline according to the schema rules.
func ValidateDataJSON(filepath string) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	var records []model.Plan
	if err := json.Unmarshal(data, &records); err != nil {
		return fmt.Errorf("unmarshaling json: %w", err)
	}

	seenPairs := make(map[string]bool)
	providerFound := make(map[string]bool)

	for i, r := range records {
		if r.Provider == "" {
			return fmt.Errorf("record %d: missing provider", i)
		}

		baseProvider := r.Provider
		if strings.HasPrefix(r.Provider, "Vultr") {
			baseProvider = "Vultr"
		}
		providerFound[baseProvider] = true

		mem, err := strconv.ParseFloat(r.Memory, 64)
		if err != nil || mem <= 0 {
			return fmt.Errorf("record %d (%s): invalid memory '%s'", i, r.Provider, r.Memory)
		}

		cpu, err := strconv.ParseFloat(r.CPU, 64)
		if err != nil || cpu <= 0 {
			return fmt.Errorf("record %d (%s): invalid cpu '%s'", i, r.Provider, r.CPU)
		}

		disk, err := strconv.ParseFloat(r.Disk, 64)
		if err != nil || disk <= 0 {
			return fmt.Errorf("record %d (%s): invalid disk '%s'", i, r.Provider, r.Disk)
		}

		price, err := strconv.ParseFloat(r.Price, 64)
		if err != nil || price <= 0 {
			return fmt.Errorf("record %d (%s): invalid price '%s'", i, r.Provider, r.Price)
		}

		if r.UpdatedAt == "" {
			return fmt.Errorf("record %d (%s): missing updated_at", i, r.Provider)
		}

		pair := fmt.Sprintf("%s|%s|%s|%s", r.Provider, r.Memory, r.CPU, r.Disk)
		if seenPairs[pair] {
			return fmt.Errorf("duplicate record found: %s", pair)
		}
		seenPairs[pair] = true

		curr := strings.ToLower(r.Currency)
		if curr == "" {
			curr = "usd"
		}
		if curr != "usd" && curr != "eur" && curr != "gbp" {
			return fmt.Errorf("record %d (%s): unknown currency '%s'", i, r.Provider, r.Currency)
		}
	}

	for _, p := range ExpectedProviders {
		if !providerFound[p] {
			return fmt.Errorf("expected provider '%s' not found in data.json", p)
		}
	}

	return nil
}

func RoundUSDPrice(price float64) float64 {
	return math.Round(price*100) / 100
}

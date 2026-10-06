package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Record struct {
	Provider  string `json:"provider"`
	Memory    string `json:"memory"`
	CPU       string `json:"cpu"`
	Disk      string `json:"disk"`
	Price     string `json:"price"`
	UpdatedAt string `json:"updated_at"`
	Currency  string `json:"currency,omitempty"`
}

var expectedProviders = []string{
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

var exchangeRates = map[string]float64{
	"usd": 1.00,
	"eur": 1.12,
	"gbp": 1.32,
}

func main() {
	data, err := os.ReadFile("data.json")
	if err != nil {
		fmt.Printf("Error reading data.json: %v\n", err)
		os.Exit(1)
	}

	var records []Record
	if err := json.Unmarshal(data, &records); err != nil {
		fmt.Printf("Error unmarshaling data.json: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Validating %d records in data.json...\n", len(records))

	seenPairs := make(map[string]bool)
	providerFound := make(map[string]bool)

	for i, r := range records {
		if r.Provider == "" {
			fmt.Printf("Record %d: Missing provider\n", i)
			os.Exit(1)
		}

		// Normalize provider name for mapping to expectedProviders
		baseProvider := r.Provider
		if strings.HasPrefix(r.Provider, "Vultr") {
			baseProvider = "Vultr"
		}
		providerFound[baseProvider] = true

		mem, err := strconv.ParseFloat(r.Memory, 64)
		if err != nil || mem <= 0 {
			fmt.Printf("Record %d (%s): Invalid memory '%s'\n", i, r.Provider, r.Memory)
			os.Exit(1)
		}

		cpu, err := strconv.ParseFloat(r.CPU, 64)
		if err != nil || cpu <= 0 {
			fmt.Printf("Record %d (%s): Invalid cpu '%s'\n", i, r.Provider, r.CPU)
			os.Exit(1)
		}

		disk, err := strconv.ParseFloat(r.Disk, 64)
		if err != nil || disk <= 0 {
			fmt.Printf("Record %d (%s): Invalid disk '%s'\n", i, r.Provider, r.Disk)
			os.Exit(1)
		}

		price, err := strconv.ParseFloat(r.Price, 64)
		if err != nil || price <= 0 {
			fmt.Printf("Record %d (%s): Invalid price '%s'\n", i, r.Provider, r.Price)
			os.Exit(1)
		}

		if r.UpdatedAt == "" {
			fmt.Printf("Record %d (%s): Missing updated_at\n", i, r.Provider)
			os.Exit(1)
		}

		pair := fmt.Sprintf("%s|%s|%s|%s", r.Provider, r.Memory, r.CPU, r.Disk)
		if seenPairs[pair] {
			fmt.Printf("Duplicate record found: %s\n", pair)
			os.Exit(1)
		}
		seenPairs[pair] = true

		curr := strings.ToLower(r.Currency)
		if curr == "" {
			curr = "usd"
		}
		rate, ok := exchangeRates[curr]
		if !ok {
			fmt.Printf("Record %d (%s): Unknown currency '%s'\n", i, r.Provider, r.Currency)
			os.Exit(1)
		}

		usdPrice := math.Round(price*rate*100) / 100
		_ = usdPrice
	}

	for _, p := range expectedProviders {
		if !providerFound[p] {
			fmt.Printf("Expected provider '%s' not found in data.json!\n", p)
			os.Exit(1)
		}
	}

	// Verify research files exist
	files, err := os.ReadDir("research")
	if err != nil {
		fmt.Printf("Error reading research dir: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d research files.\n", len(files))
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".md") {
			content, err := os.ReadFile(filepath.Join("research", f.Name()))
			if err != nil || len(content) == 0 {
				fmt.Printf("Empty or unreadable research file: %s\n", f.Name())
				os.Exit(1)
			}
		}
	}

	fmt.Println("All validations passed successfully!")
}

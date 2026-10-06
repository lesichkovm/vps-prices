package runner

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/lesichkovm/vps-prices/internal/fx"
	"github.com/lesichkovm/vps-prices/internal/model"
	"github.com/lesichkovm/vps-prices/internal/providers"
)

type UpdateOptions struct {
	DryRun     bool
	Timeout    time.Duration
	ReportPath string
	Providers  []string
}

func getSourceType(pName string) (string, bool, string) {
	switch pName {
	case "AWS Lightsail":
		return "bulk file", false, ""
	case "CloudFanatic":
		return "HTML", false, ""
	case "Contabo":
		return "manual", false, ""
	case "DigitalOcean":
		return "token API", true, "DIGITALOCEAN_TOKEN"
	case "eVPS":
		return "HTML", false, ""
	case "Hetzner":
		return "token API", true, "HCLOUD_TOKEN"
	case "Linode":
		return "API", false, ""
	case "OVH":
		return "API", false, ""
	case "Raff Technologies":
		return "HTML", false, ""
	case "Scaleway":
		return "API", false, ""
	case "Vultr":
		return "API", false, ""
	default:
		return "API / HTML", false, ""
	}
}

func getProviderSlug(pName string) string {
	switch pName {
	case "AWS Lightsail":
		return "aws_lightsail"
	case "CloudFanatic":
		return "cloudfanatic"
	case "Contabo":
		return "contabo"
	case "DigitalOcean":
		return "digitalocean"
	case "eVPS":
		return "evps"
	case "Hetzner":
		return "hetzner"
	case "Linode":
		return "linode"
	case "OVH":
		return "ovh"
	case "Raff Technologies":
		return "raff_technologies"
	case "Scaleway":
		return "scaleway"
	case "Vultr":
		return "vultr"
	default:
		return strings.ToLower(strings.ReplaceAll(pName, " ", "_"))
	}
}

type providerResult struct {
	ProviderName string
	Status       ProviderRunStatus
	Plans        []model.Plan
	Failed       bool
	Warnings     []string
}

func RunUpdate(ctx context.Context, opts UpdateOptions) error {
	slog.Info("Starting VPS price update run", "dry_run", opts.DryRun, "timeout", opts.Timeout)

	// Step 1: Load existing data.json
	existingData, err := os.ReadFile("data.json")
	if err != nil {
		return fmt.Errorf("failed to read data.json: %w", err)
	}

	var existingPlans []model.Plan
	if err := json.Unmarshal(existingData, &existingPlans); err != nil {
		return fmt.Errorf("failed to parse existing data.json: %w", err)
	}

	existingMap := make(map[string][]model.Plan)
	for _, p := range existingPlans {
		baseProvider := p.Provider
		if strings.HasPrefix(p.Provider, "Vultr") {
			baseProvider = "Vultr"
		}
		existingMap[baseProvider] = append(existingMap[baseProvider], p)
	}

	// Step 2: Fetch Exchange Rates
	ratesClient := providers.NewHTTPClient(30 * time.Second)
	ratesCtx, ratesCancel := context.WithTimeout(ctx, 30*time.Second)
	defer ratesCancel()

	rates, err := fx.FetchRates(ratesCtx, ratesClient.Client())
	if err != nil {
		return fmt.Errorf("aborting run: FX rates unavailable: %w", err)
	}
	slog.Info("Fetched FX rates", "date", rates.Date, "source", rates.Source, "EUR_USD", rates.EURToUSD)

	// Step 3: Filter Providers
	allProviders := providers.GetAllProviders(opts.Timeout)
	var activeProviders []model.Provider

	if len(opts.Providers) > 0 {
		filterMap := make(map[string]bool)
		for _, name := range opts.Providers {
			filterMap[strings.ToLower(strings.TrimSpace(name))] = true
		}
		for _, p := range allProviders {
			if filterMap[strings.ToLower(p.Name())] {
				activeProviders = append(activeProviders, p)
			}
		}
	} else {
		activeProviders = allProviders
	}

	// Step 4: Run Fetchers Concurrently with Worker Pool
	workerCap := 4
	jobs := make(chan model.Provider, len(activeProviders))
	results := make(chan providerResult, len(activeProviders))

	for _, p := range activeProviders {
		jobs <- p
	}
	close(jobs)

	var wg sync.WaitGroup
	for i := 0; i < workerCap; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for p := range jobs {
				res := fetchAndValidateProvider(ctx, p, existingMap[p.Name()], rates)
				results <- res
			}
		}()
	}

	wg.Wait()
	close(results)

	// Step 5: Process Results
	runDate := time.Now().UTC().Format("2006-01-02")
	finalPlansMap := make(map[string][]model.Plan)

	// Copy unselected providers from existingMap
	for pName, pPlans := range existingMap {
		finalPlansMap[pName] = pPlans
	}

	var reportRun RunReport
	reportRun.Date = runDate
	reportRun.Rates = rates

	for res := range results {
		pName := res.ProviderName
		reportRun.Providers = append(reportRun.Providers, res.Status)
		reportRun.Warnings = append(reportRun.Warnings, res.Warnings...)

		if !res.Failed && len(res.Plans) > 0 {
			finalPlansMap[pName] = res.Plans
		} else {
			// Keep existing plans for skipped/failed/manual provider
			slog.Warn("Keeping existing plans for provider", "provider", pName, "status", res.Status.Status)
		}
	}

	// Flatten all final plans
	var finalPlans []model.Plan
	for _, pList := range finalPlansMap {
		finalPlans = append(finalPlans, pList...)
	}

	// Calculate changes summary
	reportRun.PlansAdded, reportRun.PlansUpdated, reportRun.PlansRemoved, reportRun.LargeChanges = comparePlans(existingPlans, finalPlans)

	// Step 6: Write files unless dry-run
	if !opts.DryRun {
		slog.Info("Writing data.json atomically")
		if err := WriteDataJSONAtomically("data.json", finalPlans); err != nil {
			return fmt.Errorf("failed writing data.json: %w", err)
		}

		slog.Info("Updating README.md FX section")
		if err := fx.UpdateREADMERates("README.md", rates); err != nil {
			slog.Error("Failed to update README FX rates", "error", err)
		}

		// Update research notes for fetched providers
		for _, pStatus := range reportRun.Providers {
			if pStatus.Status == "ok" {
				pPlans := finalPlansMap[pStatus.Name]
				if err := updateResearchNote(pStatus.Name, runDate, rates, pPlans); err != nil {
					slog.Error("Failed updating research note", "provider", pStatus.Name, "error", err)
				}
			}
		}

		if opts.ReportPath != "" {
			slog.Info("Writing run report", "path", opts.ReportPath)
			reportMd := GenerateReportMarkdown(reportRun)
			if err := atomicWriteFile(opts.ReportPath, []byte(reportMd), 0644); err != nil {
				slog.Error("Failed writing report markdown", "error", err)
			}
		}
	} else {
		slog.Info("Dry run finished. No files modified.")
	}

	return nil
}

func fetchAndValidateProvider(ctx context.Context, p model.Provider, oldPlans []model.Plan, rates *fx.Rates) providerResult {
	pName := p.Name()
	sourceType, needsSec, secName := getSourceType(pName)

	status := ProviderRunStatus{
		Name:         pName,
		SourceType:   sourceType,
		NeedsSecrets: needsSec,
		SecretName:   secName,
	}

	fetched, err := p.Fetch(ctx)

	if errors.Is(err, providers.ErrTokenMissing) {
		status.Status = "skipped"
		status.Message = fmt.Sprintf("Skipped: %s secret not provided", secName)
		return providerResult{ProviderName: pName, Status: status, Failed: true, Warnings: []string{status.Message}}
	}

	if errors.Is(err, providers.ErrManualProvider) {
		status.Status = "manual"
		status.Message = "Requires manual review (Cloudflare anti-bot / no public API)"
		return providerResult{ProviderName: pName, Status: status, Failed: true, Warnings: []string{status.Message}}
	}

	if err != nil {
		status.Status = "failed"
		status.Message = fmt.Sprintf("Fetch failed: %v", err)
		return providerResult{ProviderName: pName, Status: status, Failed: true, Warnings: []string{status.Message}}
	}

	// Safety Guard 1: Empty check
	if len(fetched) == 0 {
		status.Status = "failed"
		status.Message = "Fetched zero plans"
		return providerResult{ProviderName: pName, Status: status, Failed: true, Warnings: []string{status.Message}}
	}

	// Safety Guard 2: Drop > 30% check
	if len(oldPlans) > 0 {
		dropRatio := float64(len(oldPlans)-len(fetched)) / float64(len(oldPlans))
		if dropRatio > 0.30 {
			status.Status = "failed"
			status.Message = fmt.Sprintf("Safety check failed: plan count dropped by %.1f%% (%d -> %d)", dropRatio*100, len(oldPlans), len(fetched))
			return providerResult{ProviderName: pName, Status: status, Failed: true, Warnings: []string{status.Message}}
		}
	}

	// Safety Guard 3: Non-zero values & Duplicate checks
	seenPairs := make(map[string]bool)
	var processed []model.Plan
	runDate := time.Now().UTC().Format("2006-01-02")

	for _, plan := range fetched {
		memVal, errMem := strconv.ParseFloat(plan.Memory, 64)
		cpuVal, errCPU := strconv.ParseFloat(plan.CPU, 64)
		diskVal, errDisk := strconv.ParseFloat(plan.Disk, 64)
		origPrice, errPrice := strconv.ParseFloat(plan.Price, 64)

		if errMem != nil || memVal <= 0 || errCPU != nil || cpuVal <= 0 || errDisk != nil || diskVal <= 0 || errPrice != nil || origPrice <= 0 {
			status.Status = "failed"
			status.Message = fmt.Sprintf("Safety check failed: plan with zero/invalid specs found (mem: %s, cpu: %s, disk: %s, price: %s)", plan.Memory, plan.CPU, plan.Disk, plan.Price)
			return providerResult{ProviderName: pName, Status: status, Failed: true, Warnings: []string{status.Message}}
		}

		pairKey := fmt.Sprintf("%s|%s|%s|%s", plan.Provider, plan.Memory, plan.CPU, plan.Disk)
		if seenPairs[pairKey] {
			status.Status = "failed"
			status.Message = fmt.Sprintf("Safety check failed: duplicate plan pair detected (%s)", pairKey)
			return providerResult{ProviderName: pName, Status: status, Failed: true, Warnings: []string{status.Message}}
		}
		seenPairs[pairKey] = true

		// Convert price to USD
		rateToUSD, usdPrice, errFX := rates.ConvertToUSD(origPrice, plan.Currency)
		if errFX != nil {
			status.Status = "failed"
			status.Message = fmt.Sprintf("FX conversion failed for currency %s: %v", plan.Currency, errFX)
			return providerResult{ProviderName: pName, Status: status, Failed: true, Warnings: []string{status.Message}}
		}

		plan.OriginalPrice = origPrice
		plan.USDPrice = usdPrice
		_ = rateToUSD
		plan.UpdatedAt = runDate

		processed = append(processed, plan)
	}

	status.Status = "ok"
	status.PlansCount = len(processed)
	status.Message = "Fetched successfully"

	return providerResult{
		ProviderName: pName,
		Status:       status,
		Plans:        processed,
		Failed:       false,
	}
}

func comparePlans(oldPlans, newPlans []model.Plan) (added, updated, removed int, largeChanges []LargePriceChange) {
	oldMap := make(map[string]model.Plan)
	newMap := make(map[string]model.Plan)

	for _, p := range oldPlans {
		key := fmt.Sprintf("%s|%s|%s|%s", p.Provider, p.Memory, p.CPU, p.Disk)
		oldMap[key] = p
	}

	for _, p := range newPlans {
		key := fmt.Sprintf("%s|%s|%s|%s", p.Provider, p.Memory, p.CPU, p.Disk)
		newMap[key] = p
	}

	for key, nPlan := range newMap {
		oPlan, exists := oldMap[key]
		if !exists {
			added++
		} else {
			oPrice, _ := strconv.ParseFloat(oPlan.Price, 64)
			nPrice, _ := strconv.ParseFloat(nPlan.Price, 64)

			if oPrice != nPrice || oPlan.Currency != nPlan.Currency {
				updated++
			}

			// Large change check (> 50%)
			if oPrice > 0 {
				pctChange := ((nPrice - oPrice) / oPrice) * 100.0
				if pctChange > 50.0 || pctChange < -50.0 {
					largeChanges = append(largeChanges, LargePriceChange{
						Provider: nPlan.Provider,
						PlanSpec: fmt.Sprintf("%s GB RAM, %s vCPU, %s GB Disk", nPlan.Memory, nPlan.CPU, nPlan.Disk),
						OldPrice: oPrice,
						NewPrice: nPrice,
						Change:   pctChange,
					})
				}
			}
		}
	}

	for key := range oldMap {
		if _, exists := newMap[key]; !exists {
			removed++
		}
	}

	return added, updated, removed, largeChanges
}

func updateResearchNote(pName, dateStr string, rates *fx.Rates, plans []model.Plan) error {
	slug := getProviderSlug(pName)
	path := FindResearchFile(slug)

	sourceType, _, _ := getSourceType(pName)

	var url string
	if len(plans) > 0 && plans[0].SourceURL != "" {
		url = plans[0].SourceURL
	} else {
		url = "https://www.google.com/"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Research: %s\n", slug))
	sb.WriteString(fmt.Sprintf("**Date/Time (UTC):** %sT00:00:00.000000\n", dateStr))
	sb.WriteString(fmt.Sprintf("**URL:** %s\n\n", url))

	sb.WriteString("## Findings\n\n")
	sb.WriteString(fmt.Sprintf("Pricing verified from official source (net of VAT, on-demand, data source: %s).\n\n", sourceType))

	sb.WriteString("## Pricing Table\n\n")
	sb.WriteString("| Plan | vCPU | RAM (GB) | Disk (GB) | Price (Original) | Currency | Exchange Rate | USD Price |\n")
	sb.WriteString("| --- | --- | --- | --- | --- | --- | --- | --- |\n")

	for _, p := range plans {
		currUpper := strings.ToUpper(p.Currency)
		rateToUSD, usdPrice, _ := rates.ConvertToUSD(p.OriginalPrice, p.Currency)
		rateDisplay := fx.DisplayRate(rateToUSD)

		currSymbol := "$"
		if currUpper == "EUR" {
			currSymbol = "€"
		} else if currUpper == "GBP" {
			currSymbol = "£"
		}

		planName := fmt.Sprintf("%s %sGB", p.Provider, p.Memory)
		if p.PlanID != "" {
			planName = p.PlanID
		}

		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s%s | %s | %s | $%.2f |\n",
			planName, p.CPU, p.Memory, p.Disk, currSymbol, p.Price, currUpper, rateDisplay, usdPrice))
	}

	sb.WriteString("\n## Notes\n")
	region := "Standard Region"
	if len(plans) > 0 && plans[0].Region != "" {
		region = plans[0].Region
	}
	sb.WriteString(fmt.Sprintf("- Region: %s.\n", region))
	sb.WriteString("- VAT: Net of VAT.\n")
	sb.WriteString("- Billing basis: Monthly on-demand.\n")
	sb.WriteString(fmt.Sprintf("- Data Source Type: %s.\n", sourceType))

	return atomicWriteFile(path, []byte(sb.String()), 0644)
}

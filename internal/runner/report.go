package runner

import (
	"fmt"
	"strings"

	"github.com/lesichkovm/vps-prices/internal/fx"
)

type ProviderRunStatus struct {
	Name           string
	SourceType     string // API / token API / bulk file / HTML / manual
	Status         string // ok / skipped / manual / failed
	PlansCount     int
	Message        string
	NeedsSecrets   bool
	SecretName     string
}

type LargePriceChange struct {
	Provider string
	PlanSpec string
	OldPrice float64
	NewPrice float64
	Change   float64
}

type RunReport struct {
	Date              string
	Rates             *fx.Rates
	Providers         []ProviderRunStatus
	PlansAdded        int
	PlansUpdated      int
	PlansRemoved      int
	LargeChanges      []LargePriceChange
	Warnings          []string
}

func GenerateReportMarkdown(report RunReport) string {
	var sb strings.Builder

	sb.WriteString("# VPS Price Update Report\n\n")
	sb.WriteString(fmt.Sprintf("**Date (UTC):** %s  \n", report.Date))
	sb.WriteString(fmt.Sprintf("**FX Source:** %s  \n", report.Rates.Source))
	eurToUSD, _, _ := report.Rates.ConvertToUSD(1.0, "EUR")
	gbpToUSD, _, _ := report.Rates.ConvertToUSD(1.0, "GBP")
	sb.WriteString(fmt.Sprintf("**Exchange Rates:** 1 EUR = %s USD | 1 GBP = %s USD\n\n", fx.DisplayRate(eurToUSD), fx.DisplayRate(gbpToUSD)))

	sb.WriteString("## Provider Status\n\n")
	sb.WriteString("| Provider | Data Source Type | Status | Plans Count | Notes |\n")
	sb.WriteString("| --- | --- | --- | --- | --- |\n")

	for _, p := range report.Providers {
		statusBadge := p.Status
		switch p.Status {
		case "ok":
			statusBadge = "✅ ok"
		case "skipped":
			statusBadge = "⚠️ skipped"
		case "manual":
			statusBadge = "✋ manual"
		case "failed":
			statusBadge = "❌ failed"
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %d | %s |\n", p.Name, p.SourceType, statusBadge, p.PlansCount, p.Message))
	}

	sb.WriteString("\n## Changes Summary\n\n")
	sb.WriteString(fmt.Sprintf("- **Plans Added:** %d\n", report.PlansAdded))
	sb.WriteString(fmt.Sprintf("- **Plans Updated:** %d\n", report.PlansUpdated))
	sb.WriteString(fmt.Sprintf("- **Plans Removed:** %d\n", report.PlansRemoved))

	if len(report.LargeChanges) > 0 {
		sb.WriteString("\n## Large Price Changes (> 50%)\n\n")
		sb.WriteString("| Provider | Plan Spec | Old Price (USD) | New Price (USD) | Change |\n")
		sb.WriteString("| --- | --- | --- | --- | --- |\n")
		for _, lc := range report.LargeChanges {
			sb.WriteString(fmt.Sprintf("| %s | %s | $%.2f | $%.2f | %+.1f%% |\n", lc.Provider, lc.PlanSpec, lc.OldPrice, lc.NewPrice, lc.Change))
		}
	} else {
		sb.WriteString("\n## Large Price Changes (> 50%)\n\nNo price changes over 50% detected.\n")
	}

	if len(report.Warnings) > 0 {
		sb.WriteString("\n## Warnings & Manual Review Items\n\n")
		for _, w := range report.Warnings {
			sb.WriteString(fmt.Sprintf("- %s\n", w))
		}
	}

	return sb.String()
}

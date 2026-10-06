package generator

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/lesichkovm/vps-prices/internal/fx"
	"github.com/lesichkovm/vps-prices/internal/model"
)

type Generator struct {
	DataPath    string
	IndexPath   string
	SitemapPath string
}

func NewGenerator(dataPath, indexPath, sitemapPath string) *Generator {
	if dataPath == "" {
		dataPath = "data.json"
	}
	if indexPath == "" {
		indexPath = "index.html"
	}
	if sitemapPath == "" {
		sitemapPath = "sitemap.xml"
	}
	return &Generator{
		DataPath:    dataPath,
		IndexPath:   indexPath,
		SitemapPath: sitemapPath,
	}
}

func (g *Generator) GenerateAll(rates *fx.Rates) error {
	if err := g.GenerateHTML(rates); err != nil {
		return fmt.Errorf("failed generating HTML: %w", err)
	}
	if err := g.GenerateSitemap(); err != nil {
		return fmt.Errorf("failed generating sitemap: %w", err)
	}
	return nil
}

func (g *Generator) GenerateHTML(rates *fx.Rates) error {
	dataContent, err := os.ReadFile(g.DataPath)
	if err != nil {
		return fmt.Errorf("failed to read data file %s: %w", g.DataPath, err)
	}

	var plans []model.Plan
	if err := json.Unmarshal(dataContent, &plans); err != nil {
		return fmt.Errorf("failed to parse json from %s: %w", g.DataPath, err)
	}

	indexContent, err := os.ReadFile(g.IndexPath)
	if err != nil {
		return fmt.Errorf("failed to read index file %s: %w", g.IndexPath, err)
	}

	if rates == nil {
		rates = parseRatesFromREADME("README.md")
	}

	type calculatedPlan struct {
		Plan     model.Plan
		USDPrice float64
	}

	var calcPlans []calculatedPlan
	providersSet := make(map[string]bool)
	latestDate := "2026-01-01"

	for _, p := range plans {
		if p.Provider != "" {
			providersSet[p.Provider] = true
		}
		if p.UpdatedAt > latestDate {
			latestDate = p.UpdatedAt
		}

		origPrice, _ := strconv.ParseFloat(p.Price, 64)
		_, usdPrice, err := rates.ConvertToUSD(origPrice, p.Currency)
		if err != nil {
			usdPrice = origPrice
		}
		usdPrice = math.Round(usdPrice*100) / 100

		calcPlans = append(calcPlans, calculatedPlan{
			Plan:     p,
			USDPrice: usdPrice,
		})
	}

	// Sort plans by USD price ascending, then provider name
	sort.SliceStable(calcPlans, func(i, j int) bool {
		if calcPlans[i].USDPrice != calcPlans[j].USDPrice {
			return calcPlans[i].USDPrice < calcPlans[j].USDPrice
		}
		return calcPlans[i].Plan.Provider < calcPlans[j].Plan.Provider
	})

	// Pre-render Alert Banner
	eurRateDisplay := fx.DisplayRate(rates.EURToUSD)
	gbpRateDisplay := "1.32"
	if rateGBP, ok := rates.RatesToEUR["GBP"]; ok && rateGBP > 0 {
		gbpRateDisplay = fx.DisplayRate(rates.EURToUSD / rateGBP)
	}

	alertHTML := fmt.Sprintf(
		`<div class="alert alert-info text-center py-2 mb-3 shadow-sm" role="alert">`+
			`<i class="bi bi-patch-check-fill me-2 text-primary"></i>`+
			`<small>Prices verified from official sources on <strong>%s</strong>. Exchange rates: 1 EUR = %s USD, 1 GBP = %s USD.</small>`+
			`</div>`,
		latestDate, eurRateDisplay, gbpRateDisplay,
	)

	// Pre-render Provider Filter Options
	var providerNames []string
	for pName := range providersSet {
		providerNames = append(providerNames, pName)
	}
	sort.Strings(providerNames)

	var providerOptsBuilder strings.Builder
	providerOptsBuilder.WriteString("<option value=\"all\">All Providers</option>\n")
	for _, pName := range providerNames {
		providerOptsBuilder.WriteString(fmt.Sprintf("            <option value=\"%s\">%s</option>\n", htmlEscape(pName), htmlEscape(pName)))
	}
	providersHTML := strings.TrimRight(providerOptsBuilder.String(), "\n")

	// Pre-render Table Rows
	var rowsBuilder strings.Builder
	rowsBuilder.WriteString("<tbody>\n")
	for _, item := range calcPlans {
		p := item.Plan
		rowsBuilder.WriteString("          <tr>\n")
		rowsBuilder.WriteString(fmt.Sprintf("            <td>%s</td>\n", htmlEscape(p.Provider)))
		rowsBuilder.WriteString(fmt.Sprintf("            <td>%s</td>\n", htmlEscape(p.Memory)))
		rowsBuilder.WriteString(fmt.Sprintf("            <td>%s</td>\n", htmlEscape(p.CPU)))
		rowsBuilder.WriteString(fmt.Sprintf("            <td>%s</td>\n", htmlEscape(p.Disk)))
		rowsBuilder.WriteString(fmt.Sprintf("            <td>$%.2f</td>\n", item.USDPrice))
		rowsBuilder.WriteString(fmt.Sprintf("            <td>%s</td>\n", htmlEscape(p.UpdatedAt)))
		rowsBuilder.WriteString("          </tr>\n")
	}
	rowsBuilder.WriteString("        </tbody>")
	rowsHTML := rowsBuilder.String()

	htmlStr := string(indexContent)

	// Replace Alert
	htmlStr = replaceSection(htmlStr, "<!-- PRERENDERED_ALERT:START -->", "<!-- PRERENDERED_ALERT:END -->", alertHTML)

	// Replace Provider Options
	htmlStr = replaceSection(htmlStr, "<!-- PRERENDERED_PROVIDERS:START -->", "<!-- PRERENDERED_PROVIDERS:END -->", providersHTML)

	// Replace Table Rows
	htmlStr = replaceSection(htmlStr, "<!-- PRERENDERED_ROWS:START -->", "<!-- PRERENDERED_ROWS:END -->", rowsHTML)

	return os.WriteFile(g.IndexPath, []byte(htmlStr), 0644)
}

func (g *Generator) GenerateSitemap() error {
	dataContent, err := os.ReadFile(g.DataPath)
	if err != nil {
		return fmt.Errorf("failed to read data file %s: %w", g.DataPath, err)
	}

	var plans []model.Plan
	_ = json.Unmarshal(dataContent, &plans)

	latestDate := "2026-10-06"
	for _, p := range plans {
		if p.UpdatedAt > latestDate {
			latestDate = p.UpdatedAt
		}
	}

	sitemapContent := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url>
    <loc>https://vpsprices.lesichkov.co.uk/</loc>
    <lastmod>%s</lastmod>
    <changefreq>monthly</changefreq>
    <priority>1.0</priority>
  </url>
</urlset>
`, latestDate)

	return os.WriteFile(g.SitemapPath, []byte(sitemapContent), 0644)
}

func replaceSection(src, startMarker, endMarker, newContent string) string {
	startIdx := strings.Index(src, startMarker)
	endIdx := strings.Index(src, endMarker)
	if startIdx == -1 || endIdx == -1 || startIdx >= endIdx {
		return src
	}
	prefix := src[:startIdx+len(startMarker)]
	suffix := src[endIdx:]
	return prefix + "\n        " + newContent + "\n        " + suffix
}

func htmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}

func parseRatesFromREADME(readmePath string) *fx.Rates {
	defaultRates := &fx.Rates{
		Date:       "2026-10-05",
		Source:     "European Central Bank",
		EURToUSD:   1.12,
		RatesToEUR: map[string]float64{"EUR": 1.0, "USD": 1.12, "GBP": 0.8485},
	}

	content, err := os.ReadFile(readmePath)
	if err != nil {
		return defaultRates
	}

	reEUR := regexp.MustCompile(`1\s*EUR\s*=\s*([0-9\.]+)\s*USD`)
	reGBP := regexp.MustCompile(`1\s*GBP\s*=\s*([0-9\.]+)\s*USD`)

	matchesEUR := reEUR.FindStringSubmatch(string(content))
	matchesGBP := reGBP.FindStringSubmatch(string(content))

	if len(matchesEUR) > 1 {
		if v, err := strconv.ParseFloat(matchesEUR[1], 64); err == nil && v > 0 {
			defaultRates.EURToUSD = v
			defaultRates.RatesToEUR["USD"] = v
		}
	}

	if len(matchesGBP) > 1 {
		if gbpUSD, err := strconv.ParseFloat(matchesGBP[1], 64); err == nil && gbpUSD > 0 && defaultRates.EURToUSD > 0 {
			// 1 GBP = gbpUSD USD -> 1 EUR = eurUSD USD -> 1 EUR = (eurUSD / gbpUSD) GBP
			gbpRateToEUR := defaultRates.EURToUSD / gbpUSD
			defaultRates.RatesToEUR["GBP"] = gbpRateToEUR
		}
	}

	return defaultRates
}

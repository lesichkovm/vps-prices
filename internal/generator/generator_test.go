package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lesichkovm/vps-prices/internal/fx"
)

func TestGenerator(t *testing.T) {
	tmpDir := t.TempDir()

	testDataJSON := filepath.Join(tmpDir, "data.json")
	testIndexHTML := filepath.Join(tmpDir, "index.html")
	testSitemap := filepath.Join(tmpDir, "sitemap.xml")

	dataContent := `[
		{"provider": "ProviderA", "memory": "2", "cpu": "1", "disk": "40", "price": "5.00", "currency": "usd", "updated_at": "2026-10-06"},
		{"provider": "ProviderB", "memory": "4", "cpu": "2", "disk": "80", "price": "10.00", "currency": "eur", "updated_at": "2026-10-06"}
	]`
	if err := os.WriteFile(testDataJSON, []byte(dataContent), 0644); err != nil {
		t.Fatalf("Failed to write test data.json: %v", err)
	}

	indexContent := `<!DOCTYPE html>
<html>
<head><title>Test</title></head>
<body>
<!-- PRERENDERED_ALERT:START -->
<p>Old alert</p>
<!-- PRERENDERED_ALERT:END -->

<select id="provider-filter">
<!-- PRERENDERED_PROVIDERS:START -->
<option>Old Provider</option>
<!-- PRERENDERED_PROVIDERS:END -->
</select>

<table id="vps-table">
<!-- PRERENDERED_ROWS:START -->
<tbody><tr><td>Old Row</td></tr></tbody>
<!-- PRERENDERED_ROWS:END -->
</table>
</body>
</html>`
	if err := os.WriteFile(testIndexHTML, []byte(indexContent), 0644); err != nil {
		t.Fatalf("Failed to write test index.html: %v", err)
	}

	rates := &fx.Rates{
		Date:       "2026-10-06",
		Source:     "Test FX",
		EURToUSD:   1.12,
		RatesToEUR: map[string]float64{"EUR": 1.0, "USD": 1.12, "GBP": 0.85},
	}

	gen := NewGenerator(testDataJSON, testIndexHTML, testSitemap)
	if err := gen.GenerateAll(rates); err != nil {
		t.Fatalf("GenerateAll failed: %v", err)
	}

	// Verify HTML output
	updatedIndex, err := os.ReadFile(testIndexHTML)
	if err != nil {
		t.Fatalf("Failed to read updated index.html: %v", err)
	}
	htmlStr := string(updatedIndex)

	if !strings.Contains(htmlStr, "Prices verified from official sources on <strong>2026-10-06</strong>") {
		t.Errorf("Updated HTML missing alert banner. Got:\n%s", htmlStr)
	}
	if !strings.Contains(htmlStr, `<option value="ProviderA">ProviderA</option>`) || !strings.Contains(htmlStr, `<option value="ProviderB">ProviderB</option>`) {
		t.Errorf("Updated HTML missing provider options. Got:\n%s", htmlStr)
	}
	if !strings.Contains(htmlStr, "<td>ProviderA</td>") || !strings.Contains(htmlStr, "<td>ProviderB</td>") {
		t.Errorf("Updated HTML missing pre-rendered table rows. Got:\n%s", htmlStr)
	}
	if !strings.Contains(htmlStr, "<td>$11.20</td>") { // EUR 10.00 * 1.12 = 11.20
		t.Errorf("Updated HTML missing converted EUR price $11.20. Got:\n%s", htmlStr)
	}

	// Verify Sitemap output
	sitemapContent, err := os.ReadFile(testSitemap)
	if err != nil {
		t.Fatalf("Failed to read sitemap.xml: %v", err)
	}
	sitemapStr := string(sitemapContent)
	if !strings.Contains(sitemapStr, "<lastmod>2026-10-06</lastmod>") || !strings.Contains(sitemapStr, "https://vpsprices.lesichkov.co.uk/") {
		t.Errorf("Sitemap incorrect. Got:\n%s", sitemapStr)
	}
}

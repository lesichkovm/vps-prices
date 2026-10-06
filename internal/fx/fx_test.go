package fx

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const mockECBXML = `<?xml version="1.0" encoding="UTF-8"?>
<gesmes:Envelope xmlns:gesmes="http://www.gesmes.org/xml/2002-08-01" xmlns="http://www.ecb.int/vocabulary/2002-08-01/eurofxref">
	<Cube>
		<Cube time='2026-10-05'>
			<Cube currency='USD' rate='1.1204'/>
			<Cube currency='GBP' rate='0.8472'/>
		</Cube>
	</Cube>
</gesmes:Envelope>`

func TestFetchECB(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(mockECBXML))
	}))
	defer server.Close()
}

func TestConversion(t *testing.T) {
	rates := &Rates{
		Date:     "2026-10-05",
		Source:   "Test ECB",
		EURToUSD: 1.1204,
		RatesToEUR: map[string]float64{
			"EUR": 1.0,
			"USD": 1.1204,
			"GBP": 0.8472,
		},
	}

	// USD to USD
	_, usdVal, err := rates.ConvertToUSD(10.0, "USD")
	if err != nil || usdVal != 10.0 {
		t.Errorf("USD conversion failed: %v, got %f", err, usdVal)
	}

	// EUR to USD
	_, eurVal, err := rates.ConvertToUSD(10.0, "EUR")
	if err != nil || eurVal != 11.204 {
		t.Errorf("EUR conversion failed: %v, got %f", err, eurVal)
	}

	// GBP to USD
	gbpRate, gbpVal, err := rates.ConvertToUSD(10.0, "GBP")
	if err != nil {
		t.Fatalf("GBP conversion error: %v", err)
	}
	expectedGBPRate := 1.1204 / 0.8472
	if gbpRate != expectedGBPRate {
		t.Errorf("Expected GBP rate %f, got %f", expectedGBPRate, gbpRate)
	}
	if gbpVal != 10.0*expectedGBPRate {
		t.Errorf("Expected GBP USD value %f, got %f", 10.0*expectedGBPRate, gbpVal)
	}
}

func TestUpdateREADME(t *testing.T) {
	tempDir := t.TempDir()
	readmeFile := filepath.Join(tempDir, "README.md")

	initialReadme := `# Test Project

### Currency Standardization

Some old text here.

## Other section
`
	if err := os.WriteFile(readmeFile, []byte(initialReadme), 0644); err != nil {
		t.Fatal(err)
	}

	rates := &Rates{
		Date:     "2026-10-05",
		Source:   "European Central Bank",
		EURToUSD: 1.12,
		RatesToEUR: map[string]float64{
			"EUR": 1.0,
			"USD": 1.12,
			"GBP": 0.848,
		},
	}

	if err := UpdateREADMERates(readmeFile, rates); err != nil {
		t.Fatalf("UpdateREADMERates failed: %v", err)
	}

	updated, err := os.ReadFile(readmeFile)
	if err != nil {
		t.Fatal(err)
	}

	strUpdated := string(updated)
	if !strings.Contains(strUpdated, StartMarker) || !strings.Contains(strUpdated, EndMarker) {
		t.Errorf("README missing markers: %s", strUpdated)
	}
	if !strings.Contains(strUpdated, "1 EUR = 1.12 USD") {
		t.Errorf("README missing EUR rate: %s", strUpdated)
	}
}

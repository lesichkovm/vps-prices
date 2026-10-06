package fx

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

const (
	StartMarker = "<!-- FX-RATES:START -->"
	EndMarker   = "<!-- FX-RATES:END -->"
)

// UpdateREADMEUpdates updates the Currency Standardization section in README.md with the latest rates and fetch date.
func UpdateREADMERates(readmePath string, rates *Rates) error {
	content, err := os.ReadFile(readmePath)
	if err != nil {
		return fmt.Errorf("reading README: %w", err)
	}

	strContent := string(content)

	eurRate, _, _ := rates.ConvertToUSD(1.0, "EUR")
	gbpRate, _, _ := rates.ConvertToUSD(1.0, "GBP")

	sectionText := fmt.Sprintf(`%s
To allow a fair, apples-to-apples comparison across all providers, all prices originally quoted in Euros (EUR) or British Pounds (GBP) are converted to US Dollars (USD). The exchange rates fetched on %s from %s are:

- **1 EUR = %s USD**
- **1 GBP = %s USD**
%s`, StartMarker, rates.Date, rates.Source, DisplayRate(eurRate), DisplayRate(gbpRate), EndMarker)

	var newContent string
	if strings.Contains(strContent, StartMarker) && strings.Contains(strContent, EndMarker) {
		re := regexp.MustCompile("(?s)" + regexp.QuoteMeta(StartMarker) + ".*?" + regexp.QuoteMeta(EndMarker))
		newContent = re.ReplaceAllString(strContent, sectionText)
	} else {
		// Replace standard Currency Standardization section if markers aren't present yet
		targetHeader := "### Currency Standardization"
		if idx := strings.Index(strContent, targetHeader); idx != -1 {
			// Find end of section (next ## or end of file)
			nextSectionIdx := strings.Index(strContent[idx+len(targetHeader):], "\n## ")
			if nextSectionIdx != -1 {
				endIdx := idx + len(targetHeader) + nextSectionIdx
				newContent = strContent[:idx] + targetHeader + "\n\n" + sectionText + "\n\n" + strContent[endIdx:]
			} else {
				newContent = strContent[:idx] + targetHeader + "\n\n" + sectionText
			}
		} else {
			newContent = strContent + "\n\n" + targetHeader + "\n\n" + sectionText
		}
	}

	return os.WriteFile(readmePath, []byte(newContent), 0644)
}

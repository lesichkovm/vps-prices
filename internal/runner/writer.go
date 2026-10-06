package runner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/lesichkovm/vps-prices/internal/model"
)

// WriteDataJSONAtomically serializes plans with 2-space indentation and writes atomically to path.
func WriteDataJSONAtomically(filepath string, plans []model.Plan) error {
	// Sort plans deterministically: USDPrice asc, Provider asc, Memory asc, CPU asc, Disk asc
	sortPlans(plans)

	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)

	if err := enc.Encode(plans); err != nil {
		return fmt.Errorf("encoding JSON: %w", err)
	}

	return atomicWriteFile(filepath, buf.Bytes(), 0644)
}

func sortPlans(plans []model.Plan) {
	sort.SliceStable(plans, func(i, j int) bool {
		pi, pj := plans[i], plans[j]

		if pi.USDPrice != pj.USDPrice {
			return pi.USDPrice < pj.USDPrice
		}
		if pi.Provider != pj.Provider {
			return pi.Provider < pj.Provider
		}

		memI, _ := strconv.ParseFloat(pi.Memory, 64)
		memJ, _ := strconv.ParseFloat(pj.Memory, 64)
		if memI != memJ {
			return memI < memJ
		}

		cpuI, _ := strconv.ParseFloat(pi.CPU, 64)
		cpuJ, _ := strconv.ParseFloat(pj.CPU, 64)
		if cpuI != cpuJ {
			return cpuI < cpuJ
		}

		diskI, _ := strconv.ParseFloat(pi.Disk, 64)
		diskJ, _ := strconv.ParseFloat(pj.Disk, 64)
		return diskI < diskJ
	})
}

func atomicWriteFile(filename string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(filename)
	tmpFile, err := os.CreateTemp(dir, "tmp-*")
	if err != nil {
		return fmt.Errorf("creating temp file in %s: %w", dir, err)
	}
	tmpName := tmpFile.Name()
	defer os.Remove(tmpName)

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		return fmt.Errorf("writing to temp file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("closing temp file: %w", err)
	}

	if err := os.Chmod(tmpName, perm); err != nil {
		return fmt.Errorf("setting permissions on temp file: %w", err)
	}

	if err := os.Rename(tmpName, filename); err != nil {
		return fmt.Errorf("renaming temp file to %s: %w", filename, err)
	}

	return nil
}

// FindResearchFile finds the existing research file for a provider in research/ or returns a default path.
func FindResearchFile(providerSlug string) string {
	files, err := os.ReadDir("research")
	if err != nil {
		return filepath.Join("research", fmt.Sprintf("%s.md", providerSlug))
	}

	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".md") {
			name := strings.TrimSuffix(f.Name(), ".md")
			if strings.HasSuffix(name, "_"+providerSlug) || name == providerSlug {
				return filepath.Join("research", f.Name())
			}
		}
	}

	return filepath.Join("research", fmt.Sprintf("%s.md", providerSlug))
}

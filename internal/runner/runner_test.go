package runner

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/lesichkovm/vps-prices/internal/model"
)

func TestDataJSONByteFidelity(t *testing.T) {
	dataPath := filepath.Join("..", "..", "data.json")
	original, err := os.ReadFile(dataPath)
	if err != nil {
		t.Fatalf("Failed to read data.json: %v", err)
	}

	var plans []model.Plan
	if err := json.Unmarshal(original, &plans); err != nil {
		t.Fatalf("Failed to unmarshal data.json: %v", err)
	}

	// Re-serialize via Encoder
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)

	if err := enc.Encode(plans); err != nil {
		t.Fatalf("Failed to encode plans: %v", err)
	}

	reencoded := buf.Bytes()
	if !bytes.Equal(bytes.TrimSpace(original), bytes.TrimSpace(reencoded)) {
		t.Errorf("Byte fidelity failed! Original size: %d, Re-encoded size: %d", len(original), len(reencoded))
	}
}

func TestComparePlans(t *testing.T) {
	oldPlans := []model.Plan{
		{Provider: "TestProv", Memory: "1", CPU: "1", Disk: "20", Price: "5.00", Currency: "usd"},
		{Provider: "TestProv", Memory: "2", CPU: "2", Disk: "40", Price: "10.00", Currency: "usd"},
	}

	newPlans := []model.Plan{
		{Provider: "TestProv", Memory: "1", CPU: "1", Disk: "20", Price: "5.00", Currency: "usd"},  // Unchanged
		{Provider: "TestProv", Memory: "2", CPU: "2", Disk: "40", Price: "18.00", Currency: "usd"}, // Large change (>50%)
		{Provider: "TestProv", Memory: "4", CPU: "4", Disk: "80", Price: "20.00", Currency: "usd"}, // Added
	}

	added, updated, removed, largeChanges := comparePlans(oldPlans, newPlans)

	if added != 1 {
		t.Errorf("Expected 1 added plan, got %d", added)
	}
	if updated != 1 {
		t.Errorf("Expected 1 updated plan, got %d", updated)
	}
	if removed != 0 {
		t.Errorf("Expected 0 removed plans, got %d", removed)
	}
	if len(largeChanges) != 1 {
		t.Errorf("Expected 1 large price change, got %d", len(largeChanges))
	} else {
		if largeChanges[0].Change != 80.0 {
			t.Errorf("Expected +80%% change, got %f%%", largeChanges[0].Change)
		}
	}
}

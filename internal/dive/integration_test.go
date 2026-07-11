//go:build integration

package dive

import (
	"encoding/json"
	"os"
	"testing"
)

// TestDiveJSONCompatibility is run by scripts/verify-dive-compat.sh after a
// real dive --json invocation. Set DIVE_JSON_OUTPUT to the JSON file path.
func TestDiveJSONCompatibility(t *testing.T) {
	path := os.Getenv("DIVE_JSON_OUTPUT")
	if path == "" {
		t.Skip("DIVE_JSON_OUTPUT not set")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading dive output: %v", err)
	}

	var a Analysis
	if err := json.Unmarshal(data, &a); err != nil {
		t.Fatalf("parsing dive JSON output: %v", err)
	}
	if len(a.Layers) == 0 {
		t.Fatal("expected at least one layer")
	}
	if a.Image.SizeBytes <= 0 {
		t.Fatal("expected positive image sizeBytes")
	}
	if a.Image.EfficiencyScore < 0 || a.Image.EfficiencyScore > 1 {
		t.Fatalf("efficiencyScore out of range: %v", a.Image.EfficiencyScore)
	}
}

package dive

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func loadFixture(t *testing.T) *Analysis {
	t.Helper()
	data, err := os.ReadFile("testdata/sample.json")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}
	var a Analysis
	if err := json.Unmarshal(data, &a); err != nil {
		t.Fatalf("parsing fixture: %v", err)
	}
	return &a
}

func TestParseAnalysis(t *testing.T) {
	a := loadFixture(t)

	if len(a.Layers) != 5 {
		t.Fatalf("expected 5 layers, got %d", len(a.Layers))
	}
	if !strings.Contains(a.Layers[0].Command, "alpine-minirootfs") {
		t.Errorf("expected first layer command to mention alpine-minirootfs, got %q", a.Layers[0].Command)
	}

	if a.Image.SizeBytes != 8755204 {
		t.Errorf("SizeBytes = %d, want 8755204", a.Image.SizeBytes)
	}
	if a.Image.InefficientBytes != 102412 {
		t.Errorf("InefficientBytes = %d, want 102412", a.Image.InefficientBytes)
	}
	const wantEff = 0.988303413604069
	if diff := a.Image.EfficiencyScore - wantEff; diff > 1e-9 || diff < -1e-9 {
		t.Errorf("EfficiencyScore = %v, want %v", a.Image.EfficiencyScore, wantEff)
	}

	if len(a.Image.FileReference) != 2 {
		t.Fatalf("expected 2 file references, got %d", len(a.Image.FileReference))
	}
}

func TestTopWasted(t *testing.T) {
	a := loadFixture(t)

	top := TopWasted(a.Image.FileReference, 1)
	if len(top) != 1 {
		t.Fatalf("expected 1 result, got %d", len(top))
	}
	if top[0].File != "/tmp/big.bin" {
		t.Errorf("expected big.bin to be the top wasted file, got %q", top[0].File)
	}
	if top[0].TotalWastedBytes() != 204800 {
		t.Errorf("TotalWastedBytes = %d, want 204800", top[0].TotalWastedBytes())
	}

	all := TopWasted(a.Image.FileReference, 0)
	if len(all) != 2 {
		t.Fatalf("limit<=0 should return all entries, got %d", len(all))
	}

	tooMany := TopWasted(a.Image.FileReference, 100)
	if len(tooMany) != 2 {
		t.Fatalf("limit > len should return all entries, got %d", len(tooMany))
	}
}

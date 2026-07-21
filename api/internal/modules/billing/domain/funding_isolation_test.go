package domain

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestFundingIsolation proves that the billing module never reads from
// child_funding_records. This invariant is critical: child_funding_records
// is the sole source of truth, and billing reads exclusively via FundingLookup.
func TestFundingIsolation(t *testing.T) {
	billingDir := filepath.Join("..", "..", "..", "modules", "billing")

	err := filepath.Walk(billingDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		content := string(data)
		if strings.Contains(content, "child_funding_records") {
			t.Errorf("billing module file %s references child_funding_records — this violates the funding isolation invariant", path)
		}
		if strings.Contains(content, "ChildFundingRecord") {
			t.Errorf("billing module file %s references ChildFundingRecord — this violates the funding isolation invariant", path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk billing directory: %v", err)
	}
}

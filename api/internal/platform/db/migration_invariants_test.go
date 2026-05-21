package db

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestMigrationPairsExist(t *testing.T) {
	_, filePath, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to resolve test file path")
	}
	root := filepath.Join(filepath.Dir(filePath), "..", "..", "..", "db", "migrations")
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("read migrations dir: %v", err)
	}

	up := map[string]bool{}
	down := map[string]bool{}
	for _, entry := range entries {
		name := entry.Name()
		if len(name) < 8 {
			continue
		}
		if filepath.Ext(name) != ".sql" {
			continue
		}
		base := name[:len(name)-len(filepath.Ext(name))]
		if len(base) > 3 && base[len(base)-3:] == ".up" {
			up[base[:len(base)-3]] = true
			continue
		}
		if len(base) > 5 && base[len(base)-5:] == ".down" {
			down[base[:len(base)-5]] = true
			continue
		}
	}

	for name := range up {
		if !down[name] {
			t.Fatalf("missing down migration for %s", name)
		}
	}
	for name := range down {
		if !up[name] {
			t.Fatalf("missing up migration for %s", name)
		}
	}
}

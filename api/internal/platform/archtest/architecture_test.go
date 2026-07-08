package archtest

import (
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"
)

// forbiddenDomainImports are import prefixes that must never appear in domain-layer files.
var forbiddenDomainImports = []string{
	"/infrastructure/",
	"/interfaces/",
	"pgx",
	"database/sql",
	"github.com/gin-gonic/gin",
}

// allowedDomainExceptions are import paths allowed in domain despite the rules.
var allowedDomainExceptions = []string{
	"nursery-management-system/api/internal/platform/errors",
}

func isAllowedException(importPath string, exceptions []string) bool {
	for _, e := range exceptions {
		if importPath == e {
			return true
		}
	}
	return false
}

func TestDomainLayerDoesNotImportInfrastructure(t *testing.T) {
	modulesDir := filepath.Join("..", "..", "modules")
	violations := scanLayerViolations(t, modulesDir, "domain", forbiddenDomainImports, allowedDomainExceptions)
	for _, v := range violations {
		t.Errorf("domain import violation: %s imports %q", v.File, v.Import)
	}
}

func TestApplicationLayerDoesNotImportOtherModulesInfrastructure(t *testing.T) {
	modulesDir := filepath.Join("..", "..", "modules")
	violations := scanApplicationCrossModuleViolations(t, modulesDir)
	for _, v := range violations {
		t.Errorf("application cross-module violation: %s imports %q", v.File, v.Import)
	}
}

func TestNoModuleImportsGin(t *testing.T) {
	modulesDir := filepath.Join("..", "..", "modules")
	layers := []string{"domain", "application"}
	for _, layer := range layers {
		violations := scanLayerViolations(t, modulesDir, layer, []string{"github.com/gin-gonic/gin"}, nil)
		for _, v := range violations {
			t.Errorf("gin import in %s layer: %s imports %q", layer, v.File, v.Import)
		}
	}
}

type layerViolation struct {
	File   string
	Import string
}

// extractModuleName extracts the module name from a file path like
// .../modules/billing/domain/file.go -> "billing"
func extractModuleName(filePath, layer string) string {
	dir := filepath.Dir(filePath)
	parent := filepath.Dir(dir)
	return filepath.Base(parent)
}

func scanApplicationCrossModuleViolations(t *testing.T, baseDir string) []layerViolation {
	t.Helper()
	var violations []layerViolation

	pattern := filepath.Join(baseDir, "*", "application", "*.go")
	files, err := filepath.Glob(pattern)
	if err != nil {
		t.Fatalf("glob %s: %v", pattern, err)
	}

	fset := token.NewFileSet()
	for _, filePath := range files {
		if strings.HasSuffix(filePath, "_test.go") {
			continue
		}

		moduleName := extractModuleName(filePath, "application")

		absPath, err := filepath.Abs(filePath)
		if err != nil {
			t.Fatalf("abs path for %s: %v", filePath, err)
		}

		node, err := parser.ParseFile(fset, absPath, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse %s: %v", absPath, err)
		}

		for _, imp := range node.Imports {
			importPath := strings.Trim(imp.Path.Value, `"`)
			// Check if this imports another module's infrastructure
			if strings.Contains(importPath, "/infrastructure/") {
				// Extract the imported module name
				parts := strings.Split(importPath, "/modules/")
				if len(parts) >= 2 {
					importedModule := strings.Split(parts[1], "/")[0]
					if importedModule != moduleName {
						relPath, _ := filepath.Rel(".", filePath)
						violations = append(violations, layerViolation{
							File:   relPath,
							Import: importPath,
						})
					}
				}
			}
		}
	}

	return violations
}

func scanLayerViolations(t *testing.T, baseDir, layer string, forbidden []string, exceptions []string) []layerViolation {
	t.Helper()
	var violations []layerViolation

	pattern := filepath.Join(baseDir, "*", layer, "*.go")
	files, err := filepath.Glob(pattern)
	if err != nil {
		t.Fatalf("glob %s: %v", pattern, err)
	}

	fset := token.NewFileSet()
	for _, filePath := range files {
		if strings.HasSuffix(filePath, "_test.go") {
			continue
		}

		absPath, err := filepath.Abs(filePath)
		if err != nil {
			t.Fatalf("abs path for %s: %v", filePath, err)
		}

		node, err := parser.ParseFile(fset, absPath, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse %s: %v", absPath, err)
		}

		for _, imp := range node.Imports {
			importPath := strings.Trim(imp.Path.Value, `"`)
			if isAllowedException(importPath, exceptions) {
				continue
			}
			for _, forbidden := range forbidden {
				if strings.Contains(importPath, forbidden) {
					relPath, _ := filepath.Rel(".", filePath)
					violations = append(violations, layerViolation{
						File:   relPath,
						Import: importPath,
					})
					break
				}
			}
		}
	}

	return violations
}

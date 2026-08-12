// AST reads that identify a kind's registered Go kind name (parseGoKindName). Its
// wire:"data.*" tagged fields (parseDataFieldsFromAST, plus the shared goTypeExprStr type
// stringifier) are the sibling ast_datafields.go.
package kindscan

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// goIdentRE matches a legal TS/Go identifier. goKind is emitted as an unquoted
// TS object key in node-defs.ts, so a non-identifier name (hyphen, space, leading
// digit) would produce invalid TS; validate it at parse time and fail loudly.
var goIdentRE = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// parseGoKindName extracts the first string argument to Register (nodes/wire's
// wire.Register, or the pre-decompose monolithic Wiring package's Register) in pkgDir.
func parseGoKindName(pkgDir string) (string, error) {
	entries, err := os.ReadDir(pkgDir)
	if err != nil {
		return "", err
	}
	// RegisterBuilder is the SELF-CONSTRUCTION registration (build_args.go): a kind
	// that builds itself no longer calls wire.Register at all, and a generator that only
	// knew the old marker silently dropped it from NODE_DEFS — the editor then loses the
	// kind while the Go side works fine.
	markers := []string{`Wiring.RegisterBuilder("`}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(pkgDir, name))
		if err != nil {
			continue
		}
		s := string(data)
		for _, marker := range markers {
			_, rest, ok := strings.Cut(s, marker)
			if !ok {
				continue
			}
			name2, _, ok2 := strings.Cut(rest, `"`)
			if !ok2 {
				continue
			}
			if !goIdentRE.MatchString(name2) {
				return "", fmt.Errorf("kind name %q from %s in %s is not a legal identifier (must match [A-Za-z_][A-Za-z0-9_]*); it is emitted as an unquoted TS object key", name2, marker, pkgDir)
			}
			return name2, nil
		}
	}
	return "", fmt.Errorf("RegisterBuilder call not found in %s", pkgDir)
}

package kindscan

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var goIdentRE = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func parseGoKindName(pkgDir string) (string, error) {
	entries, err := os.ReadDir(pkgDir)
	if err != nil {
		return "", err
	}

	markers := []string{`Wiring.RegisterBuilder("`}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {  // path-resolution-ok: a package directory listing, not a scene path
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

package kindscan

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

var networkSegments = [2]string{"src", "Node"}

var kindsSegments = [2]string{"src", "NodeKinds"}

func NetworkDir(repoRoot string) string {
	return filepath.Join(repoRoot, networkSegments[0], networkSegments[1])
}

func KindsDir(repoRoot string) string {
	return filepath.Join(repoRoot, kindsSegments[0], kindsSegments[1])
}

func Kinds(repoRoot string) []KindEntry {
	nodesDir := KindsDir(repoRoot)
	kinds := CollectKinds(nodesDir)
	AssignKindIDs(kinds, nodesDir)
	sort.Slice(kinds, func(i, j int) bool {
		return kinds[i].GoKind < kinds[j].GoKind
	})
	return kinds
}

func KindsPkg(repoRoot string) string {
	modPath := filepath.Join(repoRoot, "go.mod")
	src, err := os.ReadFile(modPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "kindscan: read %s: %v\n", modPath, err)
		os.Exit(1)
	}
	var module string
	for _, line := range strings.Split(string(src), "\n") {
		if rest, ok := strings.CutPrefix(strings.TrimSpace(line), "module "); ok {
			module = strings.TrimSpace(rest)
			break
		}
	}
	if module == "" {
		fmt.Fprintf(os.Stderr, "kindscan: %s declares no module path, so the generated blank imports "+
			"would name a package that does not exist and every kind would vanish from the binary\n", modPath)
		os.Exit(1)
	}
	return path.Join(append([]string{module}, kindsSegments[:]...)...)
}

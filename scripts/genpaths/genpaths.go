package genpaths

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"

	"strings"

	"github.com/dtauraso/wirefold/scripts/kindscan"
)

func Roots() (repoRoot, srcRoot string) {
	cwd, err := os.Getwd()
	if err != nil {
		Fatalf("getwd: %v", err)
	}
	repoRoot = FindRepoRoot(cwd)
	if repoRoot == "" {
		Fatalf("could not locate repo root (no go.mod found from %s)", cwd)
	}
	return repoRoot, SrcRoot(repoRoot)
}

func Kinds(repoRoot string) []kindscan.KindEntry {
	nodesDir := KindsDir(repoRoot)
	kinds := kindscan.CollectKinds(nodesDir)
	kindscan.AssignKindIDs(kinds, nodesDir)
	sort.Slice(kinds, func(i, j int) bool {
		return kinds[i].GoKind < kinds[j].GoKind
	})
	return kinds
}

func NetworkDir(repoRoot string) string {
	return filepath.Join(repoRoot, networkSegments[0], networkSegments[1])
}

func KindsDir(repoRoot string) string {
	return filepath.Join(repoRoot, kindsSegments[0], kindsSegments[1])
}

func KindsPkg(repoRoot string) string {
	modPath := filepath.Join(repoRoot, "go.mod")
	src, err := os.ReadFile(modPath)
	if err != nil {
		Fatalf("read %s: %v", modPath, err)
	}
	var module string
	for _, line := range strings.Split(string(src), "\n") {
		if rest, ok := strings.CutPrefix(strings.TrimSpace(line), "module "); ok {
			module = strings.TrimSpace(rest)
			break
		}
	}
	if module == "" {
		Fatalf("%s declares no module path", modPath)
	}
	return path.Join(append([]string{module}, kindsSegments[:]...)...)
}

var networkSegments = [2]string{"src", "Node"}

var kindsSegments = [2]string{"src", "NodeKinds"}

func Fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, Self()+": "+format+"\n", args...)
	os.Exit(1)
}

func Announce(path string, count int, unit string) {
	fmt.Fprintf(os.Stderr, "%s: wrote %s (%d %s)\n", Self(), path, count, unit)
}

var Name = "gen"

func Self() string { return Name }

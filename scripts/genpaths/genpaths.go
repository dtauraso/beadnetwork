package genpaths

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/dtauraso/wirefold/scripts/kindscan"
)

func Roots() (repoRoot, srcRoot string) {
	cwd, err := os.Getwd()
	if err != nil {
		Fatalf("getwd: %v", err)
	}
	repoRoot = FindRepoRoot(cwd)
	if repoRoot == "" {
		Fatalf("could not locate repo root (no nodes/ dir found from %s)", cwd)
	}
	return repoRoot, SrcRoot(repoRoot)
}

func Kinds(repoRoot string) []kindscan.KindEntry {
	nodesDir := filepath.Join(repoRoot, "nodes")
	kinds := kindscan.CollectKinds(nodesDir)
	kindscan.AssignKindIDs(kinds, nodesDir)
	sort.Slice(kinds, func(i, j int) bool {
		return kinds[i].GoKind < kinds[j].GoKind
	})
	return kinds
}

func Fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, Self()+": "+format+"\n", args...)
	os.Exit(1)
}

func Announce(path string, count int, unit string) {
	fmt.Fprintf(os.Stderr, "%s: wrote %s (%d %s)\n", Self(), path, count, unit)
}

var Name = "gen"

func Self() string { return Name }

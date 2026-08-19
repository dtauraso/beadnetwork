// Package genpaths is what every generator needs before it can generate anything:
// where the repo is, where the npm package's src/ is, how to name itself when it
// dies, and the node kinds when it works from them.
//
// It lives under scripts/ because it serves the repo rather than one concern —
// each generator itself lives with the thing it generates.
package genpaths

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/dtauraso/wirefold/scripts/kindscan"
)

// Roots returns the repo root and the npm package's src/ directory, dying loudly
// if either cannot be found. Every generator starts here.
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

// Kinds collects the node kinds under nodes/ and assigns their IDs, sorted by Go
// kind name. Several generators work from this list, and each one runs in its own
// process, so the order and the IDs must depend only on what is on disk.
func Kinds(repoRoot string) []kindscan.KindEntry {
	nodesDir := filepath.Join(repoRoot, "nodes")
	kinds := kindscan.CollectKinds(nodesDir)
	kindscan.AssignKindIDs(kinds, nodesDir)
	sort.Slice(kinds, func(i, j int) bool {
		return kinds[i].GoKind < kinds[j].GoKind
	})
	return kinds
}

// Fatalf prints the message under the running generator's own name and exits
// non-zero, so a failure names which generator failed.
func Fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, Self()+": "+format+"\n", args...)
	os.Exit(1)
}

// Announce reports a written file under the running generator's own name.
func Announce(path string, count int, unit string) {
	fmt.Fprintf(os.Stderr, "%s: wrote %s (%d %s)\n", Self(), path, count, unit)
}

// Name is the running generator's own name, which it sets before doing anything.
// Under `go run ./src/Trace/gen` the binary is called "gen" and the working
// directory belongs to whoever invoked it, so neither can name the generator —
// it has to say who it is.
var Name = "gen"

// Self is the running generator's name.
func Self() string { return Name }

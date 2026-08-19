package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/dtauraso/wirefold/cmd/gen-node-defs/kindscan"
)

func main() {
	repoRoot, kinds := resolveRepoRootAndKinds()

	generateKindImports(repoRoot, kinds)
	generateNodeDefs(repoRoot, kinds)
	generateWireDefs(repoRoot)
	generateTraceKinds(repoRoot)
	generateNodeDims(repoRoot, kinds)
	generateNodeKindID(repoRoot, kinds)
	generateCurveParams(repoRoot)
	generateOverlayGen(repoRoot)
	generateShadingParams(repoRoot)
	generateColumnStreams(repoRoot)
	generateBufferLayout(repoRoot)
	generateFrameTags(repoRoot)
	generateInputLayout(repoRoot)
	generateScenes(repoRoot)
}

func resolveRepoRootAndKinds() (string, []kindscan.KindEntry) {
	cwd, err := os.Getwd()
	if err != nil {
		fatalf("getwd: %v", err)
	}
	repoRoot := findRepoRoot(cwd)
	if repoRoot == "" {
		fatalf("could not locate repo root (no nodes/ dir found from %s)", cwd)
	}

	nodesDir := filepath.Join(repoRoot, "nodes")
	kinds := kindscan.CollectKinds(nodesDir)
	kindscan.AssignKindIDs(kinds, nodesDir)

	sort.Slice(kinds, func(i, j int) bool {
		return kinds[i].GoKind < kinds[j].GoKind
	})
	return repoRoot, kinds
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "gen-node-defs: "+format+"\n", args...)
	os.Exit(1)
}

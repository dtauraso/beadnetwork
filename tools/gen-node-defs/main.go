// gen-node-defs walks nodes/*/ and emits src/schema/node-defs.ts.
// Port names and directions are derived from Go channel-typed struct fields.
// View metadata and per-port accent overrides are read from SPEC.md.
// Run: go run ../../tools/gen-node-defs (from tools/topology-vscode/)
//
// This is the SINGLE entry point for every generator pipeline in this package
// (node-defs, wire-defs, trace-kinds, node-dims/kind-id, curve/shading params,
// overlay-gen, buffer-layout). tools/buffer-schema/check-generated.sh derives its guarded-file
// list from this one invocation's "wrote <path>" stderr lines — do not split
// this into multiple binaries; add new pipelines as new files/functions in this
// package and call them from main() below.
//
// main() itself holds only the entry point and the call sequence; kind
// collection, kind-id assignment, and the shared KindEntry/Port/etc. types
// live in tools/gen-node-defs/kindscan, and repo-root resolution in
// repo_root.go.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/dtauraso/wirefold/tools/gen-node-defs/kindscan"
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
	generateBufferLayout(repoRoot)
	generateFrameTags(repoRoot)
	generateInputLayout(repoRoot)
}

// resolveRepoRootAndKinds resolves the repo root (walking up from cwd until a "nodes/"
// directory is found — generator is invoked from tools/topology-vscode/ via
// npm run gen:node-defs, so resolving the root relative to this file's location at compile
// time is not possible), then collects every node kind and assigns each a stable KindId.
//
// Finding B: KindId is a stable, assigned-once fact per kind (from SPEC.md's View "kindId"
// field), NOT a sort-derived index — adding a kind must never renumber an existing one.
// Resolve each kind's id, auto-assigning (and writing back into its SPEC.md) the next free
// id for any kind that doesn't have one yet, so a new kind is stable from the moment it's
// first generated.
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

	// Sort alphabetically by Go kind name (PascalCase spec kind) for stable,
	// human-readable emission ORDER only — this sort has no bearing on the id
	// VALUE, which comes from kindID above.
	sort.Slice(kinds, func(i, j int) bool {
		return kinds[i].GoKind < kinds[j].GoKind
	})
	return repoRoot, kinds
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "gen-node-defs: "+format+"\n", args...)
	os.Exit(1)
}

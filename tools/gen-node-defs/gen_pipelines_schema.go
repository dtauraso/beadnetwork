// Pipeline phase functions that emit the TS schema files derived directly from kind/wire/trace
// data: node-defs.ts, wire-defs.ts, trace-kinds.ts, plus the two kind-indexed Go/TS id tables
// (node_dims_gen.go, node_kind_id_gen.go). Split out of main.go by concern — main.go keeps only
// the entry point and the call sequence (see its header comment).
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dtauraso/wirefold/tools/gen-node-defs/kindscan"
)

// generateKindImports writes kinds_generated.go — blank imports that make each node
// package's init() (and thus its wire.Register call) run. Folded in from the former
// standalone tools/gen-kind-imports so ONE registration scan feeds every kind-derived
// output; a separate binary could silently diverge from this one (audit finding).
// check-generated.sh guards it via the "wrote" line below, so no dedicated guard is needed.
func generateKindImports(repoRoot string, kinds []kindscan.KindEntry) {
	kindImportsPath := filepath.Join(repoRoot, "kinds_generated.go")
	if err := kindscan.WriteKindImports(kindImportsPath, kinds); err != nil {
		fatalf("write %s: %v", kindImportsPath, err)
	}
	fmt.Fprintf(os.Stderr, "gen-node-defs: wrote %s (%d kinds)\n", kindImportsPath, len(kinds))
}

func generateNodeDefs(repoRoot string, kinds []kindscan.KindEntry) {
	outPath := filepath.Join(repoRoot, "tools", "topology-vscode", "src", "schema", "node-defs.ts")
	if err := writeNodeDefs(outPath, kinds); err != nil {
		fatalf("write %s: %v", outPath, err)
	}
	fmt.Fprintf(os.Stderr, "gen-node-defs: wrote %s (%d entries)\n", outPath, len(kinds))
}

func generateWireDefs(repoRoot string) {
	loaderPath := filepath.Join(repoRoot, "nodes", "Wiring", "loadspec", "topo_spec.go")
	wireProps, err := parseWirePropsFromFile(loaderPath)
	if err != nil {
		fatalf("parse wire props from loadspec/topo_spec.go: %v", err)
	}
	wireDefsPath := filepath.Join(repoRoot, "tools", "topology-vscode", "src", "schema", "wire-defs.ts")
	if err := writeWireDefs(wireDefsPath, wireProps); err != nil {
		fatalf("write %s: %v", wireDefsPath, err)
	}
	fmt.Fprintf(os.Stderr, "gen-node-defs: wrote %s (%d entries)\n", wireDefsPath, len(wireProps))
}

func generateTraceKinds(repoRoot string) {
	traceDir := filepath.Join(repoRoot, "Trace")
	traceKinds, err := parseTraceKinds(traceDir)
	if err != nil {
		fatalf("parse trace kinds: %v", err)
	}
	breadcrumbLabels, err := parseBreadcrumbLabels(traceDir)
	if err != nil {
		fatalf("parse breadcrumb labels: %v", err)
	}
	traceKindsPath := filepath.Join(repoRoot, "tools", "topology-vscode", "src", "schema", "trace-kinds.ts")
	if err := writeTraceKinds(traceKindsPath, traceKinds, breadcrumbLabels); err != nil {
		fatalf("write %s: %v", traceKindsPath, err)
	}
	fmt.Fprintf(os.Stderr, "gen-node-defs: wrote %s (%d kinds)\n", traceKindsPath, len(traceKinds))
}

func generateNodeDims(repoRoot string, kinds []kindscan.KindEntry) {
	nodeDimsGoPath := filepath.Join(repoRoot, "nodes", "Wiring", "nodegeom", "node_dims_gen.go")
	if err := writeNodeDims(nodeDimsGoPath, kinds); err != nil {
		fatalf("write %s: %v", nodeDimsGoPath, err)
	}
	fmt.Fprintf(os.Stderr, "gen-node-defs: wrote %s (%d kinds)\n", nodeDimsGoPath, len(kinds))
}

func generateNodeKindID(repoRoot string, kinds []kindscan.KindEntry) {
	nodeKindIDGoPath := filepath.Join(repoRoot, "Buffer", "node_kind_id_gen.go")
	if err := writeNodeKindID(nodeKindIDGoPath, kinds); err != nil {
		fatalf("write %s: %v", nodeKindIDGoPath, err)
	}
	fmt.Fprintf(os.Stderr, "gen-node-defs: wrote %s (%d kinds)\n", nodeKindIDGoPath, len(kinds))
}

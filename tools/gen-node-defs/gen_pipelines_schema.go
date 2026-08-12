package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dtauraso/wirefold/tools/gen-node-defs/kindscan"
)

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

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dtauraso/wirefold/tools/gen-node-defs/kindscan"
	"github.com/dtauraso/wirefold/tools/gen-node-defs/nodedefs"
	"github.com/dtauraso/wirefold/tools/gen-node-defs/tracekinds"
	"github.com/dtauraso/wirefold/tools/gen-node-defs/wiredefs"
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
	if err := nodedefs.WriteNodeDefs(outPath, kinds); err != nil {
		fatalf("write %s: %v", outPath, err)
	}
	fmt.Fprintf(os.Stderr, "gen-node-defs: wrote %s (%d entries)\n", outPath, len(kinds))
}

func generateWireDefs(repoRoot string) {
	loaderPath := filepath.Join(repoRoot, "nodes", "Wiring", "loadspec", "topo_spec.go")
	wireProps, err := wiredefs.ParseWirePropsFromFile(loaderPath)
	if err != nil {
		fatalf("parse wire props from loadspec/topo_spec.go: %v", err)
	}
	wireDefsPath := filepath.Join(repoRoot, "tools", "topology-vscode", "src", "schema", "wire-defs.ts")
	if err := wiredefs.WriteWireDefs(wireDefsPath, wireProps); err != nil {
		fatalf("write %s: %v", wireDefsPath, err)
	}
	fmt.Fprintf(os.Stderr, "gen-node-defs: wrote %s (%d entries)\n", wireDefsPath, len(wireProps))
}

func generateTraceKinds(repoRoot string) {
	traceDir := filepath.Join(repoRoot, "tools", "topology-vscode", "src", "Trace")
	traceKinds, err := tracekinds.ParseTraceKinds(traceDir)
	if err != nil {
		fatalf("parse trace kinds: %v", err)
	}
	breadcrumbLabels, err := tracekinds.ParseBreadcrumbLabels(traceDir)
	if err != nil {
		fatalf("parse breadcrumb labels: %v", err)
	}
	traceKindsPath := filepath.Join(repoRoot, "tools", "topology-vscode", "src", "Trace", "trace-kinds.ts")
	if err := tracekinds.WriteTraceKinds(traceKindsPath, traceKinds, breadcrumbLabels); err != nil {
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
	nodeKindIDGoPath := filepath.Join(repoRoot, "tools", "topology-vscode", "src", "Node", "node_kind_id_gen.go")
	if err := writeNodeKindID(nodeKindIDGoPath, kinds); err != nil {
		fatalf("write %s: %v", nodeKindIDGoPath, err)
	}
	fmt.Fprintf(os.Stderr, "gen-node-defs: wrote %s (%d kinds)\n", nodeKindIDGoPath, len(kinds))
}

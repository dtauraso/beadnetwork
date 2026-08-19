// The Trace generator: the trace kinds and breadcrumb labels, read from the
// Trace directory's own sources.
package main

//go:generate go run .

import (
	"path/filepath"

	"github.com/dtauraso/wirefold/scripts/genpaths"
	"github.com/dtauraso/wirefold/src/Trace/gen/tracekinds"
)

func main() {
	genpaths.Name = "Trace/gen"
	_, srcRoot := genpaths.Roots()

	traceDir := filepath.Join(srcRoot, "Trace")
	kinds, err := tracekinds.ParseTraceKinds(traceDir)
	if err != nil {
		genpaths.Fatalf("parse trace kinds: %v", err)
	}
	labels, err := tracekinds.ParseBreadcrumbLabels(traceDir)
	if err != nil {
		genpaths.Fatalf("parse breadcrumb labels: %v", err)
	}
	outPath := filepath.Join(traceDir, "trace-kinds.ts")
	if err := tracekinds.WriteTraceKinds(outPath, kinds, labels); err != nil {
		genpaths.Fatalf("write %s: %v", outPath, err)
	}
	genpaths.Announce(outPath, len(kinds), "kinds")
}

package main

//go:generate go run .

import (
	"path/filepath"

	"github.com/dtauraso/wirefold/scripts/genpaths"
	"github.com/dtauraso/wirefold/src/Node/nodegeom/gen/params"
)

func main() {
	genpaths.Name = "nodegeom/gen"
	repoRoot, srcRoot := genpaths.Roots()
	kinds := genpaths.Kinds(repoRoot)

	dimsPath := filepath.Join(genpaths.NetworkDir(repoRoot), "nodegeom", "node_dims_gen.go")
	if err := writeNodeDims(dimsPath, kinds); err != nil {
		genpaths.Fatalf("write %s: %v", dimsPath, err)
	}
	genpaths.Announce(dimsPath, len(kinds), "kinds")

	portsPath := filepath.Join(genpaths.KindsDir(repoRoot), "portwiring", "kind_ports_gen.go")
	if err := writeKindPorts(portsPath, kinds); err != nil {
		genpaths.Fatalf("write %s: %v", portsPath, err)
	}
	genpaths.Announce(portsPath, len(kinds), "kinds")

	kindIDPath := filepath.Join(srcRoot, "Node", "node_kind_id_gen.go")
	if err := writeNodeKindID(kindIDPath, kinds); err != nil {
		genpaths.Fatalf("write %s: %v", kindIDPath, err)
	}
	genpaths.Announce(kindIDPath, len(kinds), "kinds")

	generateShadingParams(repoRoot)
}

func generateShadingParams(repoRoot string) {
	goPath := filepath.Join(genpaths.NetworkDir(repoRoot), "nodegeom", "shading_params.go")
	shadingParams, err := params.ParseShadingParams(repoRoot, goPath)
	if err != nil {
		genpaths.Fatalf("parse shading params: %v", err)
	}
	tsPath := filepath.Join(genpaths.NetworkDir(repoRoot), "nodegeom", "shading-params.ts")
	if err := params.WriteShadingParams(tsPath, shadingParams); err != nil {
		genpaths.Fatalf("write %s: %v", tsPath, err)
	}
	genpaths.Announce(tsPath, len(shadingParams), "constants")
}

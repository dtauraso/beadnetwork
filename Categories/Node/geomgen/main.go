package main

//go:generate go run .

import (
	"github.com/dtauraso/wirefold/Categories/NodeKinds/gen/kindscan"
	"path/filepath"

	"github.com/dtauraso/wirefold/Categories/Node/geomgen/params"
)

func main() {
	genName = "Categories/Node/geomgen"
	repoRoot, srcRoot := roots()
	kinds := kindscan.Kinds(repoRoot)

	dimsPath := filepath.Join(srcRoot, "Node", "geom_node_dims_gen.go")
	if err := writeNodeDims(dimsPath, kinds); err != nil {
		fatalf("write %s: %v", dimsPath, err)
	}
	announce(dimsPath, len(kinds), "kinds")

	portsPath := filepath.Join(kindscan.KindsDir(repoRoot), "portwiring", "kind_ports_gen.go")
	if err := writeKindPorts(portsPath, kinds); err != nil {
		fatalf("write %s: %v", portsPath, err)
	}
	announce(portsPath, len(kinds), "kinds")

	kindIDPath := filepath.Join(srcRoot, "Node", "node_kind_id_gen.go")
	if err := writeNodeKindID(kindIDPath, kinds); err != nil {
		fatalf("write %s: %v", kindIDPath, err)
	}
	announce(kindIDPath, len(kinds), "kinds")

	dir := filepath.Join(srcRoot, "Node")
	goPath := filepath.Join(dir, "shading_params.go")
	shadingParams, err := params.ParseShadingParams(repoRoot, goPath)
	if err != nil {
		fatalf("parse shading params: %v", err)
	}
	tsPath := filepath.Join(dir, "shading-params.ts")
	if err := params.WriteShadingParams(tsPath, shadingParams, genName, "Categories/Node/shading_params.go"); err != nil {
		fatalf("write %s: %v", tsPath, err)
	}
	announce(tsPath, len(shadingParams), "constants")
}

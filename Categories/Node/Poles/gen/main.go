package main

//go:generate go run .

import (
	"path/filepath"

	"github.com/dtauraso/wirefold/Categories/Node/nodegeom/gen/params"
)

func main() {
	genName = "Categories/Node/Poles/gen"
	repoRoot, srcRoot := roots()

	dir := filepath.Join(srcRoot, "Node", "Poles")
	goPath := filepath.Join(dir, "shading_params.go")
	shadingParams, err := params.ParseShadingParams(repoRoot, goPath)
	if err != nil {
		fatalf("parse shading params: %v", err)
	}
	tsPath := filepath.Join(dir, "shading-params.ts")
	if err := params.WriteShadingParams(tsPath, shadingParams, genName, "Categories/Node/Poles/shading_params.go"); err != nil {
		fatalf("write %s: %v", tsPath, err)
	}
	announce(tsPath, len(shadingParams), "pole ring constants")
}

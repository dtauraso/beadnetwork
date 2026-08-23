package main

//go:generate go run .

import (
	"path/filepath"

	"github.com/dtauraso/wirefold/Categories/Node/geomgen/params"
)

func main() {
	genName = "Categories/Scene/Environment/gen"
	repoRoot, srcRoot := roots()

	dir := filepath.Join(srcRoot, "Scene", "Environment")
	goPath := filepath.Join(dir, "shading_params.go")
	shadingParams, err := params.ParseShadingParams(repoRoot, goPath)
	if err != nil {
		fatalf("parse shading params: %v", err)
	}
	tsPath := filepath.Join(dir, "shading-params.ts")
	if err := params.WriteShadingParams(tsPath, shadingParams, genName, "Categories/Scene/Environment/shading_params.go"); err != nil {
		fatalf("write %s: %v", tsPath, err)
	}
	announce(tsPath, len(shadingParams), "environment constants")
}

package main

//go:generate go run .

import (
	"path/filepath"

	Scene "github.com/dtauraso/wirefold/Categories/Scene"
	"github.com/dtauraso/wirefold/scripts/genpaths"
)

func main() {
	genpaths.SetName("Categories/Scene/values")
	_, srcRoot := genpaths.Roots()

	pathsDir := filepath.Join(srcRoot, "Scene", "paths")
	if err := writeValuePathFiles(pathsDir); err != nil {
		genpaths.Fatalf("write scene value paths: %v", err)
	}
	genpaths.Announce(pathsDir, len(Scene.SceneValues), "scene value paths")

	valuesPath := filepath.Join(srcRoot, "Scene", "scene-values-gen.ts")
	if err := writeValueNames(valuesPath); err != nil {
		genpaths.Fatalf("write %s: %v", valuesPath, err)
	}
	genpaths.Announce(valuesPath, len(Scene.SceneValues), "scene values")
}

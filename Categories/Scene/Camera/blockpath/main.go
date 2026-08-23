package main

//go:generate go run .

import (
	"path/filepath"

	"github.com/dtauraso/wirefold/scripts/genpaths"
)

func main() {
	genpaths.SetName("Categories/Scene/Camera/blockpath")
	_, srcRoot := genpaths.Roots()
	dir := filepath.Join(srcRoot, "Scene", "Camera")
	if err := writeCameraPaths(dir); err != nil {
		genpaths.Fatalf("write camera block path: %v", err)
	}
	genpaths.Announce(filepath.Join(dir, "paths"), 1, "camera block path")
}

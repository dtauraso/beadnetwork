package main

//go:generate go run .

import (
	"path/filepath"

	"github.com/dtauraso/wirefold/Categories/Input/gen/inputlayout"
)

func main() {
	genName = "Categories/Input/gen"
	repoRoot, srcRoot := roots()

	declDir := filepath.Join(srcRoot, "Input", "gen")
	inputFP, err := inputlayout.ParseInputLayoutFingerprintDir(declDir)
	if err != nil {
		fatalf("parse input layout fingerprint: %v", err)
	}
	goKindsPath := filepath.Join(srcRoot, "Input", "Drag", "event_kinds_gen.go")
	if err := inputlayout.WriteGoEventKinds(goKindsPath, inputFP); err != nil {
		fatalf("write %s: %v", goKindsPath, err)
	}
	announce(goKindsPath, 1, "event kinds")

	copyRecordReaders(repoRoot)
	copyWireVocabulary(repoRoot, inputFP)
	copyTSWireVocabulary(repoRoot, inputFP)
}

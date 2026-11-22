package StepStore

import (
	"path/filepath"
	"slices"
	"strings"

	"github.com/dtauraso/beadnetwork/Categories/Node/Links/Store"
)

const FileTreeOpen = "tree-open.bin"

type StepRow struct {
	Path  []int
	Label string
	Depth int
	Kids  int
	Open  bool
	Here  bool
	Shown bool
}

func PathDir(scene string, p []int) string {
	dir := scene
	for _, n := range p {
		dir = StepDir(dir, n)
	}
	return dir
}

func rowAt(dir string, parent []int, n int, here []int) StepRow {
	p := append(slices.Clone(parent), n)
	r := StepRow{Path: p, Label: LabelAt(dir, n), Depth: len(parent), Shown: slices.Equal(p, here)}
	r.Here = len(p) <= len(here) && slices.Equal(p, here[:len(p)])
	r.Kids = len(ReadStepNumbers(StepDir(dir, n)))
	return r
}

func ReadTree(scene string, parent, here []int) []StepRow {
	var out []StepRow
	dir := PathDir(scene, parent)
	for _, n := range ReadStepNumbers(dir) {
		r := rowAt(dir, parent, n, here)
		r.Open = r.Here && !r.Shown
		Store.ReadIfExists(filepath.Join(StepDir(dir, n), FileTreeOpen), &r.Open)
		r.Open = r.Open && r.Kids > 0
		out = append(out, r)
		if r.Open {
			out = append(out, ReadTree(scene, r.Path, here)...)
		}
	}
	return out
}

func findsIn(scene, find string, p []int) bool {
	if strings.Contains(strings.ToLower(LabelAt(PathDir(scene, p[:len(p)-1]), p[len(p)-1])), find) {
		return true
	}
	for _, n := range ReadStepNumbers(PathDir(scene, p)) {
		if findsIn(scene, find, append(slices.Clone(p), n)) {
			return true
		}
	}
	return false
}

func FindTree(scene, find string, parent, here []int) []StepRow {
	find = strings.ToLower(strings.TrimSpace(find))
	var out []StepRow
	dir := PathDir(scene, parent)
	for _, n := range ReadStepNumbers(dir) {
		r := rowAt(dir, parent, n, here)
		if !findsIn(scene, find, r.Path) {
			continue
		}
		r.Open = r.Kids > 0
		out = append(out, r)
		out = append(out, FindTree(scene, find, r.Path, here)...)
	}
	return out
}

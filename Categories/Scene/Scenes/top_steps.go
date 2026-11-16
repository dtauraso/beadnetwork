package Scenes

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
)

const (
	DirTopSteps   = "steps"
	DirOldSeqs    = "sequences"
	FileStepName  = "name.bin"
	FileOpenStep  = "open-step.bin"
	dirSceneNodes = "nodes"
)

func OldSequences(scenePath string) error {
	if _, err := os.Stat(filepath.Join(scenePath, DirOldSeqs)); err == nil {
		return fmt.Errorf("%s still keeps its node sets in %s/; run `go run ./scripts/migrate-top-steps %s` to move them to %s/", scenePath, DirOldSeqs, scenePath, DirTopSteps)
	}
	return nil
}

func TopStepsDir(scenePath string) string { return filepath.Join(scenePath, DirTopSteps) }

func TopStepDir(scenePath string, k int) string {
	return filepath.Join(TopStepsDir(scenePath), strconv.Itoa(k))
}

func OpenStepFilePath(scenePath string) string { return ViewFilePath(scenePath, FileOpenStep) }

func TopSteps(scenePath string) []int {
	if _, err := os.Stat(filepath.Join(scenePath, dirSceneNodes)); err == nil {
		return nil
	}
	entries, err := os.ReadDir(TopStepsDir(scenePath))
	if err != nil {
		return nil
	}
	var out []int
	for _, e := range entries {
		k, err := strconv.Atoi(e.Name())
		if err == nil && k > 0 && e.IsDir() { // path-resolution-ok: a top step is a numbered directory under steps/; other entries are skipped
			out = append(out, k)
		}
	}
	sort.Ints(out)
	return out
}

func OpenStepPath(scenePath string) []int {
	var raw string
	ReadIfExists(OpenStepFilePath(scenePath), &raw)
	var path []int
	for _, p := range strings.Split(raw, "/") {
		if n, err := strconv.Atoi(p); err == nil && n > 0 {
			path = append(path, n)
		}
	}
	return path
}

func OpenTopStep(scenePath string) int {
	tops := TopSteps(scenePath)
	if len(tops) == 0 {
		return 0
	}
	if p := OpenStepPath(scenePath); len(p) > 0 && slices.Contains(tops, p[0]) {
		return p[0]
	}
	return tops[0]
}

func NodesRoot(scenePath string) string {
	if k := OpenTopStep(scenePath); k > 0 {
		return TopStepDir(scenePath, k)
	}
	return scenePath
}

func OpenSFilePath(scenePath string) string { return SFilePath(NodesRoot(scenePath)) }

func TopStepName(scenePath string, k int) string {
	var name string
	ReadIfExists(filepath.Join(TopStepDir(scenePath, k), FileStepName), &name)
	return name
}

func TopStepLabel(scenePath string, k int) string {
	if name := TopStepName(scenePath, k); name != "" {
		return "step " + strconv.Itoa(k) + " · " + name
	}
	return "step " + strconv.Itoa(k)
}

func WriteOpenStep(scenePath string, path []int) error {
	parts := make([]string, len(path))
	for i, n := range path {
		parts[i] = strconv.Itoa(n)
	}
	return WriteAtomic(OpenStepFilePath(scenePath), strings.Join(parts, "/"))
}

func WriteTopStepName(scenePath string, k int, name string) error {
	return WriteAtomic(filepath.Join(TopStepDir(scenePath, k), FileStepName), name)
}

func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)
		if d.IsDir() { // path-resolution-ok: copying a top step's directory tree
			return os.MkdirAll(target, 0o755)
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, raw, 0o644)
	})
}

func RemoveTopStep(scenePath string, k int) error {
	if err := os.RemoveAll(TopStepDir(scenePath, k)); err != nil {
		return err
	}
	for i, s := range TopSteps(scenePath) {
		if want := i + 1; s != want {
			if err := os.Rename(TopStepDir(scenePath, s), TopStepDir(scenePath, want)); err != nil {
				return err
			}
		}
	}
	return nil
}

func InsertTopStep(scenePath string, after int, drop string) (int, error) {
	tops := TopSteps(scenePath)
	if !slices.Contains(tops, after) {
		return 0, fmt.Errorf("there is no top step %d to copy", after)
	}
	tmp := filepath.Join(TopStepsDir(scenePath), "inserting")
	if err := copyTree(TopStepDir(scenePath, after), tmp); err != nil {
		return 0, err
	}
	if drop != "" {
		if err := os.RemoveAll(filepath.Join(tmp, drop)); err != nil {
			return 0, err
		}
	}
	if err := os.Remove(filepath.Join(tmp, FileStepName)); err != nil && !os.IsNotExist(err) {
		return 0, err
	}
	for i := len(tops) - 1; i >= 0 && tops[i] > after; i-- {
		if err := os.Rename(TopStepDir(scenePath, tops[i]), TopStepDir(scenePath, tops[i]+1)); err != nil {
			return 0, err
		}
	}
	return after + 1, os.Rename(tmp, TopStepDir(scenePath, after+1))
}

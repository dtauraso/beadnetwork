package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var genName = "gen"

func roots() (repoRoot, srcRoot string) {
	cwd, err := os.Getwd()
	if err != nil {
		fatalf("getwd: %v", err)
	}
	repoRoot = findRepoRoot(cwd)
	if repoRoot == "" {
		fatalf("could not locate repo root (no go.mod found from %s)", cwd)
	}
	return repoRoot, srcRootOf(repoRoot)
}

func findRepoRoot(dir string) string {
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func srcRootOf(repoRoot string) string {
	var found []string
	err := filepath.WalkDir(repoRoot, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() { // path-resolution-ok: walking for the npm package root, not a scene path
			switch d.Name() {
			case "node_modules", "out", ".git", ".probe", ".beadnetwork-cache":
				return filepath.SkipDir
			}
			return nil
		}
		if d.Name() == "package.json" {
			found = append(found, filepath.Dir(p))
		}
		return nil
	})
	if err != nil {
		fatalf("find npm package root under %s: %v", repoRoot, err)
	}
	if len(found) != 1 {
		fatalf("expected exactly one package.json under %s (excluding node_modules), found %d: %s\n"+
			"  The generator locates its output by the npm package, so it cannot choose between several.",
			repoRoot, len(found), strings.Join(found, ", "))
	}
	return filepath.Join(found[0], "Categories")
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, genName+": "+format+"\n", args...)
	os.Exit(1)
}

func announce(path string, count int, unit string) {
	fmt.Fprintf(os.Stderr, "%s: wrote %s (%d %s)\n", genName, path, count, unit)
}

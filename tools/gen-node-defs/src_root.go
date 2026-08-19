package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func srcRoot(repoRoot string) string {
	var found []string
	err := filepath.WalkDir(repoRoot, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case "node_modules", "out", ".git", ".probe", ".wirefold-cache":
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
	src := filepath.Join(found[0], "src")
	if st, statErr := os.Stat(src); statErr != nil || !st.IsDir() {
		fatalf("npm package at %s has no src/ directory; generated files have nowhere to go", found[0])
	}
	return src
}

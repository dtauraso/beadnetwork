package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type portRequest struct {
	name string
	dir  string
	file string
}

var portRequestRe = regexp.MustCompile(`\.(In|Out|Broadcast|DriveOut)\("([^"]+)"\)`)

func collectPortRequests(pkgDir string) []portRequest {
	entries, err := os.ReadDir(pkgDir)
	if err != nil {
		return nil
	}
	var reqs []portRequest
	for _, e := range entries {

		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") { // path-resolution-ok: a package directory listing, not a scene path
			continue
		}
		data, err := os.ReadFile(filepath.Join(pkgDir, e.Name()))
		if err != nil {
			continue
		}
		for _, m := range portRequestRe.FindAllStringSubmatch(string(data), -1) {
			dir := "out"
			if m[1] == "In" {
				dir = "in"
			}
			reqs = append(reqs, portRequest{name: m[2], dir: dir, file: e.Name()})
		}
	}
	return reqs
}

func checkPortRequests(kindDir string, pkgDir string, ports []Port) {
	declared := map[string]string{}
	for _, p := range ports {
		declared[p.ID] = p.Direction
	}
	for _, r := range collectPortRequests(pkgDir) {
		dir, ok := declared[r.name]
		if !ok {
			fatalf("kind %q: %s binds port %q, which its SPEC.md ## Ports table does not declare (declares: %v)",
				kindDir, r.file, r.name, sortedKeys(declaredSet(ports)))
		}
		if dir != r.dir {
			fatalf("kind %q: %s binds port %q as %s, but its SPEC.md ## Ports table declares it %s",
				kindDir, r.file, r.name, r.dir, dir)
		}
	}
}

func declaredSet(ports []Port) map[string]bool {
	out := map[string]bool{}
	for _, p := range ports {
		out[p.ID] = true
	}
	return out
}

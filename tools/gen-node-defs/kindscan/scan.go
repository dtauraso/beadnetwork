package kindscan

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "gen-node-defs: "+format+"\n", args...)
	os.Exit(1)
}

func CollectKinds(nodesDir string) []KindEntry {
	entries, err := os.ReadDir(nodesDir)
	if err != nil {
		fatalf("readdir nodes: %v", err)
	}

	var kinds []KindEntry
	seenGoKind := map[string]string{}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		pkgDir := filepath.Join(nodesDir, e.Name())
		if !hasRegister(pkgDir) {
			continue
		}
		ports, err := parsePortsFromAST(pkgDir)
		if err != nil {
			fatalf("parse ports %s: %v", e.Name(), err)
		}

		embedded, err := parseEmbeddedPorts(nodesDir, pkgDir, map[string]bool{})
		if err != nil {
			fatalf("parse embedded ports %s: %v", e.Name(), err)
		}
		ports = append(ports, embedded...)

		if len(ports) == 0 {
			ports = parsePortsFromSpec(pkgDir)
		}
		dataFields, err := parseDataFieldsFromAST(pkgDir)
		if err != nil {
			fatalf("parse data fields %s: %v", e.Name(), err)
		}
		goKind, err := parseGoKindName(pkgDir)
		if err != nil {
			fatalf("parse go kind name %s: %v", e.Name(), err)
		}

		if prev, dup := seenGoKind[goKind]; dup {
			fatalf("duplicate kind name %q registered by both %q and %q", goKind, prev, e.Name())
		}
		seenGoKind[goKind] = e.Name()
		view, accentOverrides, edgeKindOverrides, optionalPorts, specPortNames, err := parseSpecMD(pkgDir)
		if err != nil {

			fatalf("kind %q registers a Go runtime but its SPEC.md View section is missing/broken: %v", e.Name(), err)
		}
		if view.Kind == "" {
			fatalf("kind %q registers a Go runtime but its SPEC.md ## View has an empty view.kind", e.Name())
		}

		astPortIDs := map[string]bool{}
		for _, p := range ports {
			astPortIDs[p.ID] = true
		}
		for name := range specPortNames {
			if !astPortIDs[name] {
				fatalf("kind %q: SPEC.md Ports table Name %q does not match any Go channel-typed port (got: %v)", e.Name(), name, sortedKeys(astPortIDs))
			}
		}

		for i, p := range ports {
			if a, ok := accentOverrides[p.ID]; ok && a != "" {
				ports[i].Accent = a
			}
			if ek, ok := edgeKindOverrides[p.ID]; ok && ek != "" {
				ports[i].EdgeKind = ek
			}
			if optionalPorts[p.ID] {
				ports[i].Optional = true
			}
		}
		defaultData := parseDefaultData(pkgDir)
		kinds = append(kinds, KindEntry{Kind: view.Kind, GoKind: goKind, Dir: e.Name(), View: view, Ports: ports, DataFields: dataFields, DefaultData: defaultData})
	}
	return kinds
}

func hasRegister(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {

		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}

		if bytes.Contains(data, []byte("Wiring.RegisterBuilder(")) {
			return true
		}
	}
	return false
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

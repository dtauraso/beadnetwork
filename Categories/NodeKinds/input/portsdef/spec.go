package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func parseGoKindName(pkgDir string) (string, error) {
	entries, err := os.ReadDir(pkgDir)
	if err != nil {
		return "", err
	}
	const marker = `BuilderFor("`
	goIdentRE := regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") { // path-resolution-ok: a package directory listing, not a scene path
			continue
		}
		data, err := os.ReadFile(filepath.Join(pkgDir, name))
		if err != nil {
			continue
		}
		_, rest, ok := strings.Cut(string(data), marker)
		if !ok {
			continue
		}
		kind, _, ok := strings.Cut(rest, `"`)
		if !ok {
			continue
		}
		if !goIdentRE.MatchString(kind) {
			return "", fmt.Errorf("kind name %q is not a legal identifier (must match [A-Za-z_][A-Za-z0-9_]*); it is emitted as an unquoted TS object key", kind)
		}
		return kind, nil
	}
	return "", fmt.Errorf("no BuilderFor call in %s", pkgDir)
}

type ViewDef struct {
	Kind, KindID, Bg, Border, Text, MinWidth string
	Shape, Fill, Stroke, Width, Height, Desc string
}

func parseView(pkgDir string) (ViewDef, error) {
	lines, err := readSpecMDLines(pkgDir)
	if err != nil {
		return ViewDef{}, err
	}
	viewLines := sectionLines(lines, "View")
	if viewLines == nil {
		return ViewDef{}, fmt.Errorf("no View section")
	}
	headers, rows := parseMDTable(viewLines)
	fieldIdx := indexOf(headers, "Field")
	valueIdx := indexOf(headers, "Value")
	if fieldIdx == -1 || valueIdx == -1 {
		return ViewDef{}, fmt.Errorf("view table missing Field/Value columns")
	}
	vmap := map[string]string{}
	for _, row := range rows {
		if fieldIdx < len(row) && valueIdx < len(row) {
			vmap[row[fieldIdx]] = row[valueIdx]
		}
	}
	return ViewDef{
		Kind: vmap["kind"], KindID: vmap["kindId"], Bg: vmap["bg"],
		Border: vmap["border"], Text: vmap["text"], MinWidth: vmap["minWidth"],
		Shape: vmap["shape"], Fill: vmap["fill"], Stroke: vmap["stroke"],
		Width: vmap["width"], Height: vmap["height"],
		Desc: firstParagraph(sectionLines(lines, "Description")),
	}, nil
}

func firstParagraph(sec []string) string {
	var out []string
	for _, l := range sec {
		t := strings.TrimSpace(l)
		if t == "" {
			if len(out) > 0 {
				break
			}
			continue
		}
		out = append(out, t)
	}
	return strings.Join(out, " ")
}

type Port struct {
	ID        string
	Direction string
	EdgeKind  string
	IsMulti   bool
}

func parsePortsFromSpec(pkgDir string) []Port {
	lines, err := readSpecMDLines(pkgDir)
	if err != nil {
		return nil
	}
	tableLines := sectionLines(lines, "Ports")
	if tableLines == nil {
		return nil
	}
	headers, rows := parseMDTable(tableLines)
	nameIdx := indexOf(headers, "Name")
	dirIdx := indexOf(headers, "Direction")
	edgeKindIdx := indexOf(headers, "EdgeKind")
	if nameIdx == -1 || dirIdx == -1 {
		return nil
	}
	var ports []Port
	for _, row := range rows {
		if nameIdx >= len(row) || dirIdx >= len(row) {
			continue
		}
		name := row[nameIdx]
		dir := row[dirIdx]
		if name == "" {
			continue
		}
		multi := dir == "broadcast"
		if multi {
			dir = "out"
		}
		if dir != "in" && dir != "out" {
			continue
		}
		var edgeKind string
		if edgeKindIdx != -1 && edgeKindIdx < len(row) {
			edgeKind = row[edgeKindIdx]
		}
		ports = append(ports, Port{ID: name, Direction: dir, EdgeKind: edgeKind, IsMulti: multi})
	}
	return ports
}

// SPEC.md parsing: the View table, the Ports table (accent/edgeKind/optional
// overrides and the fallback port list), and the Default data fenced block.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// parseSpecMD reads SPEC.md in pkgDir and returns the view definition,
// a map of port-name → accent override, a map of port-name → edgeKind,
// a set of optional port names from the Ports table, and the set of every
// "Name" cell that appears in the Ports table (used by callers to validate
// each one resolves to a real AST-derived port id — a typo previously
// dropped its override silently instead of failing).
// firstParagraph joins a section's leading paragraph into ONE line — the palette shows a
// kind's description on a single row, and a SPEC is written in wrapped prose. Stops at the
// first blank line, so a Description section can say more underneath without any of it
// reaching the menu.
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

func parseSpecMD(pkgDir string) (viewDef, map[string]string, map[string]string, map[string]bool, map[string]bool, error) {
	specPortNames := map[string]bool{}
	data, readErr := os.ReadFile(filepath.Join(pkgDir, "SPEC.md"))
	if readErr != nil {
		return viewDef{}, nil, nil, nil, nil, readErr
	}
	lines := strings.Split(string(data), "\n")

	sectionLines := func(heading string) []string {
		start := -1
		for i, l := range lines {
			if strings.TrimSpace(l) == "## "+heading {
				start = i
				break
			}
		}
		if start == -1 {
			return nil
		}
		end := len(lines)
		for i := start + 1; i < len(lines); i++ {
			if strings.HasPrefix(lines[i], "## ") {
				end = i
				break
			}
		}
		return lines[start+1 : end]
	}

	// Parse a markdown table into rows.
	parseTable := func(tableLines []string) ([]string, [][]string) {
		var rows []string
		var headers []string
		var result [][]string
		for _, l := range tableLines {
			if !strings.Contains(l, "|") {
				continue
			}
			rows = append(rows, l)
		}
		if len(rows) < 2 {
			return nil, nil
		}
		// First row is headers.
		parts := strings.Split(rows[0], "|")
		for _, p := range parts {
			h := strings.TrimSpace(p)
			if h != "" {
				headers = append(headers, h)
			}
		}
		for _, row := range rows[1:] {
			parts := strings.Split(row, "|")
			var cells []string
			for _, p := range parts {
				cells = append(cells, strings.TrimSpace(p))
			}
			// Remove leading/trailing empty cells from split on "|".
			if len(cells) > 0 && cells[0] == "" {
				cells = cells[1:]
			}
			if len(cells) > 0 && cells[len(cells)-1] == "" {
				cells = cells[:len(cells)-1]
			}
			// Skip separator rows.
			allSep := true
			for _, c := range cells {
				if !isSep(c) {
					allSep = false
					break
				}
			}
			if allSep {
				continue
			}
			result = append(result, cells)
		}
		return headers, result
	}

	// Parse View section.
	viewLines := sectionLines("View")
	if viewLines == nil {
		return viewDef{}, nil, nil, nil, nil, fmt.Errorf("no View section")
	}
	headers, rows := parseTable(viewLines)
	fieldIdx := indexOf(headers, "Field")
	valueIdx := indexOf(headers, "Value")
	if fieldIdx == -1 || valueIdx == -1 {
		return viewDef{}, nil, nil, nil, nil, fmt.Errorf("view table missing Field/Value columns")
	}
	vmap := map[string]string{}
	for _, row := range rows {
		if fieldIdx < len(row) && valueIdx < len(row) {
			vmap[row[fieldIdx]] = row[valueIdx]
		}
	}
	view := viewDef{
		kind:     vmap["kind"],
		kindID:   vmap["kindId"],
		bg:       vmap["bg"],
		border:   vmap["border"],
		text:     vmap["text"],
		minWidth: vmap["minWidth"],
		shape:    vmap["shape"],
		fill:     vmap["fill"],
		stroke:   vmap["stroke"],
		width:    vmap["width"],
		height:   vmap["height"],
		desc:     firstParagraph(sectionLines("Description")),
	}

	// Parse Ports section for accent, edgeKind overrides, and optional flags.
	accentOverrides := map[string]string{}
	edgeKindOverrides := map[string]string{}
	optionalPorts := map[string]bool{}
	portsLines := sectionLines("Ports")
	if portsLines != nil {
		headers, rows := parseTable(portsLines)
		nameIdx := indexOf(headers, "Name")
		accentIdx := indexOf(headers, "Accent")
		edgeKindIdx := indexOf(headers, "EdgeKind")
		optionalIdx := indexOf(headers, "Optional")
		for _, row := range rows {
			if nameIdx >= len(row) {
				continue
			}
			name := row[nameIdx]
			if name == "" {
				continue
			}
			// Record every Ports-table Name, even if it sets no override column,
			// so callers can validate it resolves to a real AST-derived port id
			// (a typo'd Name here previously dropped its override silently).
			specPortNames[name] = true
			if accentIdx != -1 && accentIdx < len(row) && row[accentIdx] != "" {
				accentOverrides[name] = row[accentIdx]
			}
			if edgeKindIdx != -1 && edgeKindIdx < len(row) && row[edgeKindIdx] != "" {
				edgeKindOverrides[name] = row[edgeKindIdx]
			}
			if optionalIdx != -1 && optionalIdx < len(row) && row[optionalIdx] == "yes" {
				optionalPorts[name] = true
			}
		}
	}

	return view, accentOverrides, edgeKindOverrides, optionalPorts, specPortNames, nil
}

// parsePortsFromSpec reads nodes/<Kind>/SPEC.md and returns ports derived from
// the Ports table (Name + Direction columns). Used as a fallback when AST
// parsing discovers 0 ports — e.g. when all ports live in an embedded struct
// from another package that the AST walker cannot follow.
func parsePortsFromSpec(pkgDir string) []port {
	data, err := os.ReadFile(filepath.Join(pkgDir, "SPEC.md"))
	if err != nil {
		return nil
	}
	lines := strings.Split(string(data), "\n")
	// Locate ## Ports section.
	start := -1
	for i, l := range lines {
		if strings.TrimSpace(l) == "## Ports" {
			start = i
			break
		}
	}
	if start == -1 {
		return nil
	}
	end := len(lines)
	for i := start + 1; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "## ") {
			end = i
			break
		}
	}
	tableLines := lines[start+1 : end]
	// Parse the markdown table.
	var rows []string
	for _, l := range tableLines {
		if strings.Contains(l, "|") {
			rows = append(rows, l)
		}
	}
	if len(rows) < 2 {
		return nil
	}
	// Parse header row.
	var headers []string
	for _, p := range strings.Split(rows[0], "|") {
		h := strings.TrimSpace(p)
		if h != "" {
			headers = append(headers, h)
		}
	}
	nameIdx := indexOf(headers, "Name")
	dirIdx := indexOf(headers, "Direction")
	if nameIdx == -1 || dirIdx == -1 {
		return nil
	}
	var ports []port
	for _, row := range rows[1:] {
		parts := strings.Split(row, "|")
		var cells []string
		for _, p := range parts {
			cells = append(cells, strings.TrimSpace(p))
		}
		if len(cells) > 0 && cells[0] == "" {
			cells = cells[1:]
		}
		if len(cells) > 0 && cells[len(cells)-1] == "" {
			cells = cells[:len(cells)-1]
		}
		// Skip separator rows.
		allSep := true
		for _, c := range cells {
			if !isSep(c) {
				allSep = false
				break
			}
		}
		if allSep {
			continue
		}
		if nameIdx >= len(cells) || dirIdx >= len(cells) {
			continue
		}
		name := cells[nameIdx]
		dir := cells[dirIdx]
		if name == "" || (dir != "in" && dir != "out") {
			continue
		}
		ports = append(ports, port{id: name, direction: dir})
	}
	return ports
}

// parseDefaultData reads nodes/<Kind>/SPEC.md and returns the JSON string from
// the first fenced code block inside ## Default data, or "" if absent.
func parseDefaultData(pkgDir string) string {
	data, err := os.ReadFile(filepath.Join(pkgDir, "SPEC.md"))
	if err != nil {
		return ""
	}
	lines := strings.Split(string(data), "\n")
	inSection := false
	inFence := false
	var jsonLines []string
	for _, l := range lines {
		if strings.TrimSpace(l) == "## Default data" {
			inSection = true
			continue
		}
		if inSection && strings.HasPrefix(l, "## ") {
			break
		}
		if inSection && !inFence && strings.TrimSpace(l) == "```json" {
			inFence = true
			continue
		}
		if inSection && inFence {
			if strings.TrimSpace(l) == "```" {
				break
			}
			jsonLines = append(jsonLines, l)
		}
	}
	return strings.TrimSpace(strings.Join(jsonLines, "\n"))
}

// isSep reports whether s is a markdown table separator cell (e.g. "---",
// ":--", "--:"): only '-', ':', and spaces.
func isSep(s string) bool {
	for _, c := range s {
		if c != '-' && c != ':' && c != ' ' {
			return false
		}
	}
	return len(s) > 0
}

func indexOf[T comparable](slice []T, val T) int {
	for i, v := range slice {
		if v == val {
			return i
		}
	}
	return -1
}

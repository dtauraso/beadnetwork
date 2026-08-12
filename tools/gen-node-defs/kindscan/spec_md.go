// SPEC.md parsing: the View table, the Ports table (accent/edgeKind/optional
// overrides and the fallback port list), and the Default data fenced block.
package kindscan

import (
	"fmt"
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

func parseSpecMD(pkgDir string) (ViewDef, map[string]string, map[string]string, map[string]bool, map[string]bool, error) {
	specPortNames := map[string]bool{}
	lines, err := readSpecMDLines(pkgDir)
	if err != nil {
		return ViewDef{}, nil, nil, nil, nil, err
	}

	// Parse View section.
	viewLines := sectionLines(lines, "View")
	if viewLines == nil {
		return ViewDef{}, nil, nil, nil, nil, fmt.Errorf("no View section")
	}
	headers, rows := parseMDTable(viewLines)
	fieldIdx := indexOf(headers, "Field")
	valueIdx := indexOf(headers, "Value")
	if fieldIdx == -1 || valueIdx == -1 {
		return ViewDef{}, nil, nil, nil, nil, fmt.Errorf("view table missing Field/Value columns")
	}
	vmap := map[string]string{}
	for _, row := range rows {
		if fieldIdx < len(row) && valueIdx < len(row) {
			vmap[row[fieldIdx]] = row[valueIdx]
		}
	}
	view := ViewDef{
		Kind:     vmap["kind"],
		KindID:   vmap["kindId"],
		Bg:       vmap["bg"],
		Border:   vmap["border"],
		Text:     vmap["text"],
		MinWidth: vmap["minWidth"],
		Shape:    vmap["shape"],
		Fill:     vmap["fill"],
		Stroke:   vmap["stroke"],
		Width:    vmap["width"],
		Height:   vmap["height"],
		Desc:     firstParagraph(sectionLines(lines, "Description")),
	}

	// Parse Ports section for accent, edgeKind overrides, and optional flags.
	accentOverrides := map[string]string{}
	edgeKindOverrides := map[string]string{}
	optionalPorts := map[string]bool{}
	portsLines := sectionLines(lines, "Ports")
	if portsLines != nil {
		headers, rows := parseMDTable(portsLines)
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

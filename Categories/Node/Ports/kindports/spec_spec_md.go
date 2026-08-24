package main

import (
	"fmt"
	"strings"
)

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

func parseSpecMD(pkgDir string) (ViewDef, map[string]string, map[string]string, map[string]bool, error) {
	lines, err := readSpecMDLines(pkgDir)
	if err != nil {
		return ViewDef{}, nil, nil, nil, err
	}

	viewLines := sectionLines(lines, "View")
	if viewLines == nil {
		return ViewDef{}, nil, nil, nil, fmt.Errorf("no View section")
	}
	headers, rows := parseMDTable(viewLines)
	fieldIdx := indexOf(headers, "Field")
	valueIdx := indexOf(headers, "Value")
	if fieldIdx == -1 || valueIdx == -1 {
		return ViewDef{}, nil, nil, nil, fmt.Errorf("view table missing Field/Value columns")
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

	return view, accentOverrides, edgeKindOverrides, optionalPorts, nil
}

package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

func exportField(field string) string {
	if field == "" {
		return field
	}
	return strings.ToUpper(field[:1]) + field[1:]
}

type overlayFlag struct {
	flag       string
	field      string
	method     string
	breadcrumb string
	accessor   bool
	defaultOn  bool
}

type overlayOverride struct {
	field, method, breadcrumb string
	accessor, defaultOff      bool
}

var overlayOverrides = map[string]overlayOverride{
	"tori":       {field: "sceneToriVisible", method: "SceneTori"},
	"scenePoles": {breadcrumb: "scene"},
	"nodePoles":  {breadcrumb: "nodes"},
	"overlays":   {method: "OverlaysVis"},
}

// OVERLAY_FLAGS_START / OVERLAY_FLAGS_END sentinels) and returns the flag metadata in

func parseOverlayFlags(messagesPath string) ([]overlayFlag, error) {
	data, err := os.ReadFile(messagesPath)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(data), "\n")

	// sentinel (e.g. "the flags below are fenced by OVERLAY_FLAGS_START/END", exactly the

	start, end := -1, -1
	for i, l := range lines {
		switch strings.TrimSpace(l) {
		case "// OVERLAY_FLAGS_START":
			if start == -1 {
				start = i
			}
		case "// OVERLAY_FLAGS_END":
			if start != -1 && end == -1 {
				end = i
			}
		}
		if end != -1 {
			break
		}
	}
	if start == -1 || end == -1 || end <= start {
		return nil, fmt.Errorf("OVERLAY_FLAGS_START/END sentinels not found in %s", messagesPath)
	}
	strLit := regexp.MustCompile(`"([A-Za-z][A-Za-z0-9]*)"`)
	var flags []overlayFlag
	seen := map[string]bool{}
	for _, l := range lines[start+1 : end] {
		m := strLit.FindStringSubmatch(l)
		if m == nil {
			continue
		}
		name := m[1]
		seen[name] = true
		of := overlayFlag{
			flag:      name,
			field:     name + "Visible",
			method:    strings.ToUpper(name[:1]) + name[1:],
			defaultOn: true,
		}
		if ov, ok := overlayOverrides[name]; ok {
			if ov.field != "" {
				of.field = ov.field
			}
			if ov.method != "" {
				of.method = ov.method
			}
			of.breadcrumb = ov.breadcrumb
			of.accessor = ov.accessor
			if ov.defaultOff {
				of.defaultOn = false
			}
		}
		flags = append(flags, of)
	}
	if len(flags) == 0 {
		return nil, fmt.Errorf("no overlay flags parsed from %s", messagesPath)
	}

	for key := range overlayOverrides {
		if !seen[key] {
			fatalf("overlayOverrides key %q is not a real overlay flag in OVERLAY_FLAG_NAMES", key)
		}
	}
	return flags, nil
}

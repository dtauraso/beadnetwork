// overlay_gen.go: parses OVERLAY_FLAG_NAMES out of messages.ts into the overlayFlag
// vocabulary this generator's overlay pipeline runs on. TS is the input here, Go the
// output — the one inverted-direction pipeline. Emitting the Go side from that
// vocabulary (writeOverlayGen) is overlay_write.go.
package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// exportField capitalizes an overlayState field name for OverlayState's exported surface —
// OverlayState now lives in nodes/Wiring/viewstate, a different package from the code that
// reads its fields directly (Wiring's stdin_dispatch.go/scene_overlays_persist.go), so the
// fields must be exported; this is that type's OWN surface (docs/planning/gesture-actor.md),
// not a second type forced open by a third file.
func exportField(field string) string {
	if field == "" {
		return field
	}
	return strings.ToUpper(field[:1]) + field[1:]
}

// overlayFlag is one entry of the OVERLAY_FLAG_NAMES vocabulary with the mechanical
// Go names derived from (or overridden for) its camelCase flag string.
type overlayFlag struct {
	flag       string // camelCase wire flag, e.g. "tori"
	field      string // overlayState bool field, e.g. "sceneToriVisible"
	method     string // Toggle/Emit/Trace method basename, e.g. "SceneTori"
	breadcrumb string // Breadcrumb scope arg on Toggle ("scene"/"nodes"); "" = uniform flip
	accessor   bool   // emit a bare bool accessor method
	defaultOn  bool   // startup default value
}

// overlayOverride names the per-flag deviations from the uniform derivation. Kept
// data-driven (per the task's option 1) so the deviating flags are still generated,
// just with their extra behavior. Any flag absent here is fully uniform.
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

// parseOverlayFlags reads the OVERLAY_FLAG_NAMES const in messages.ts (bounded by the
// OVERLAY_FLAGS_START / OVERLAY_FLAGS_END sentinels) and returns the flag metadata in
// source order, applying overlayOverrides for the deviating flags.
func parseOverlayFlags(messagesPath string) ([]overlayFlag, error) {
	data, err := os.ReadFile(messagesPath)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(data), "\n")
	// Match the sentinels ANCHORED: a comment line carrying the marker and nothing else.
	// strings.Contains is a trap here — the moment messages.ts's own prose names the
	// sentinel (e.g. "the flags below are fenced by OVERLAY_FLAGS_START/END", exactly the
	// style this repo uses), an unanchored scan opens the fence on that prose line and the
	// generator silently emits a WRONG flag set. Same class as the guards' fence bug.
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
	// Every overlayOverrides key MUST name a real flag in OVERLAY_FLAG_NAMES. A typo
	// in an override key (e.g. "tori" mistyped "toriz") would otherwise silently fall
	// back to the uniform derivation, generating a wrong Go field/method with a clean
	// build. fatalf naming the bad key closes that gap.
	for key := range overlayOverrides {
		if !seen[key] {
			fatalf("overlayOverrides key %q is not a real overlay flag in OVERLAY_FLAG_NAMES", key)
		}
	}
	return flags, nil
}

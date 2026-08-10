package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dtauraso/wirefold/tools/gen-node-defs/kindscan"
)

// node_defs_test.go — node_dims.go / node_defs.go: emitters over a hand-built
// kindscan.KindEntry.

func TestWriteNodeDims_EmitsPerKindDimensionsWithDefaults(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "node_dims_gen.go")
	kinds := []kindscan.KindEntry{
		{GoKind: "Alpha", View: kindscan.ViewDef{Width: "150", Height: "70"}},
		{GoKind: "Beta", View: kindscan.ViewDef{}}, // no width/height: defaults expected
	}
	if err := writeNodeDims(outPath, kinds); err != nil {
		t.Fatalf("writeNodeDims: %v", err)
	}
	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	out := string(data)
	if !strings.Contains(out, `{Width: 150, Height: 70}`) {
		t.Errorf("output missing Alpha dims with explicit values:\n%s", out)
	}
	if !strings.Contains(out, `{Width: 110, Height: 60}`) {
		t.Errorf("output missing Beta dims with default fallback (110x60):\n%s", out)
	}
	if !strings.Contains(out, `"Alpha"`) || !strings.Contains(out, `"Beta"`) {
		t.Errorf("output missing kind keys:\n%s", out)
	}
	if !strings.Contains(out, "package Wiring") {
		t.Errorf("output missing package clause:\n%s", out)
	}
}

func TestWriteNodeKindID_UsesStoredKindID(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "node_kind_id_gen.go")
	// Deliberately out of alphabetical order and with a non-contiguous id (5) to
	// prove the emitted id comes from the stored kindID field, not sort position
	// or input order (finding B: ids are stable, assigned-once facts).
	kinds := []kindscan.KindEntry{
		{GoKind: "Gamma", KindID: 5},
		{GoKind: "Alpha", KindID: 0},
		{GoKind: "Beta", KindID: 2},
	}
	if err := writeNodeKindID(outPath, kinds); err != nil {
		t.Fatalf("writeNodeKindID: %v", err)
	}
	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	out := string(data)
	// gofmt column-aligns the map literal, so match key and value on
	// separate substrings rather than an exact adjacent string.
	for _, e := range kinds {
		keyWant := `"` + e.GoKind + `":`
		if !strings.Contains(out, keyWant) {
			t.Errorf("output missing key %q:\n%s", keyWant, out)
			continue
		}
		valWant := itoa(int(e.KindID)) + ",\n"
		lineStart := strings.Index(out, keyWant)
		line := out[lineStart : lineStart+len(keyWant)+20]
		if !strings.Contains(line, valWant) {
			t.Errorf("kind %q: line %q does not contain its stored kindID %d", e.GoKind, line, e.KindID)
		}
	}
}

func itoa(i int) string {
	// Small helper to avoid importing strconv just for this.
	if i == 0 {
		return "0"
	}
	digits := ""
	neg := i < 0
	if neg {
		i = -i
	}
	for i > 0 {
		digits = string(rune('0'+i%10)) + digits
		i /= 10
	}
	if neg {
		digits = "-" + digits
	}
	return digits
}

func TestWriteNodeDefs_EveryKindPortAndFieldRoundTrips(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "node-defs.ts")
	kinds := []kindscan.KindEntry{
		{
			GoKind: "Alpha",
			View: kindscan.ViewDef{
				Bg: "#111111", Border: "#222222", Text: "#333333",
				Shape: "box", Fill: "#444444", Stroke: "#555555",
				Width: "100", Height: "50",
			},
			Ports: []kindscan.Port{
				{ID: "in1", Direction: "in", EdgeKind: "chain"},
				{ID: "out1", Direction: "out", EdgeKind: "value", IsMulti: true},
			},
		},
	}
	if err := writeNodeDefs(outPath, kinds); err != nil {
		t.Fatalf("writeNodeDefs: %v", err)
	}
	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	out := string(data)

	// RUNTIME_IMPLEMENTED_KINDS must include every goKind.
	if !strings.Contains(out, `"Alpha"`) {
		t.Errorf("RUNTIME_IMPLEMENTED_KINDS missing Alpha:\n%s", out)
	}
	// NODE_DEFS entry keyed by goKind, carrying view fields.
	if !strings.Contains(out, `Alpha: { bg: "#111111"`) {
		t.Errorf("NODE_DEFS missing Alpha entry with bg field:\n%s", out)
	}
	if !strings.Contains(out, `shape: "box"`) {
		t.Errorf("NODE_DEFS Alpha missing shape field:\n%s", out)
	}
	// Input port present in inputs[], output port present in outputs[] with isMulti true.
	if !strings.Contains(out, `inputs: [{ name: "in1", kind: "chain" }]`) {
		t.Errorf("NODE_DEFS Alpha missing typed inputs entry:\n%s", out)
	}
	if !strings.Contains(out, `outputs: [{ name: "out1", kind: "value", isMulti: true }]`) {
		t.Errorf("NODE_DEFS Alpha missing typed outputs entry with isMulti:\n%s", out)
	}
	// NODE_KIND_NAMES must carry the goKind string.
	if !strings.Contains(out, `NODE_KIND_NAMES: readonly string[] = [\n  "Alpha",`) &&
		!strings.Contains(out, "NODE_KIND_NAMES") {
		t.Errorf("output missing NODE_KIND_NAMES export:\n%s", out)
	}
}

func TestFilterPorts_SplitsByDirection(t *testing.T) {
	ports := []kindscan.Port{
		{ID: "a", Direction: "in"},
		{ID: "b", Direction: "out"},
		{ID: "c", Direction: "in"},
	}
	ins := filterPorts(ports, "in")
	outs := filterPorts(ports, "out")
	if len(ins) != 2 || ins[0].ID != "a" || ins[1].ID != "c" {
		t.Errorf("filterPorts(in) = %+v, want [a, c]", ins)
	}
	if len(outs) != 1 || outs[0].ID != "b" {
		t.Errorf("filterPorts(out) = %+v, want [b]", outs)
	}
}

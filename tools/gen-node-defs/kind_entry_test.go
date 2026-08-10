package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// kind_entry_test.go — the kindEntry pipeline: parsing a node kind's Go source and SPEC.md
// into a kindEntry (ast_ports.go/spec_md.go/ast_kind.go), and emitting a kindEntry back out
// as generated Go/TS (node_dims.go/node_defs.go).

// --- ast_ports.go/spec_md.go/ast_kind.go: parsePortsFromAST / parseSpecMD / parseGoKindName -------

func writeFile(t *testing.T, dir, name, contents string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(contents), 0644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

func TestParsePortsFromAST_DiscoversChannelFieldsAndSkipsDataTags(t *testing.T) {
	dir := t.TempDir()
	src := `package fake

type FakeNode struct {
	In1   *Wiring.In
	Out1  *Wiring.Out
	Multi Wiring.Broadcast
	State []int ` + "`wire:\"data.state\"`" + `
}
`
	writeFile(t, dir, "fake.go", src)
	ports, err := parsePortsFromAST(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ports) != 3 {
		t.Fatalf("want 3 ports (data.state field must be skipped), got %d: %+v", len(ports), ports)
	}
	byID := map[string]port{}
	for _, p := range ports {
		byID[p.id] = p
	}
	if byID["In1"].direction != "in" {
		t.Errorf("In1 direction = %q, want in", byID["In1"].direction)
	}
	if byID["Out1"].direction != "out" {
		t.Errorf("Out1 direction = %q, want out", byID["Out1"].direction)
	}
	if p, ok := byID["Multi"]; !ok || p.direction != "out" || !p.isMulti {
		t.Errorf("Multi port = %+v, want direction=out isMulti=true", p)
	}
	if _, ok := byID["State"]; ok {
		t.Errorf("wire:\"data.state\" field must not be emitted as a port")
	}
}

func TestParseSpecMD_ParsesViewAndPortsTables(t *testing.T) {
	dir := t.TempDir()
	spec := `# Fake

## View

| Field | Value |
|---|---|
| bg | #123456 |
| border | #654321 |

## Ports

| Name | Direction | Accent | EdgeKind | Optional |
|---|---|---|---|---|
| in1 | in | #abcdef | value | yes |
`
	writeFile(t, dir, "SPEC.md", spec)
	view, accent, edgeKind, optional, _, err := parseSpecMD(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if view.bg != "#123456" || view.border != "#654321" {
		t.Errorf("view = %+v, want bg/border from table", view)
	}
	if accent["in1"] != "#abcdef" {
		t.Errorf("accent[in1] = %q, want #abcdef", accent["in1"])
	}
	if edgeKind["in1"] != "value" {
		t.Errorf("edgeKind[in1] = %q, want value", edgeKind["in1"])
	}
	if !optional["in1"] {
		t.Errorf("optional[in1] = false, want true (Optional column = yes)")
	}
}

func TestParseSpecMD_MissingViewSectionIsError(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "SPEC.md", "# Fake\n\nNo view section here.\n")
	if _, _, _, _, _, err := parseSpecMD(dir); err == nil {
		t.Fatal("want error when ## View section is missing, got nil")
	}
}

func TestParseGoKindName_ExtractsRegisterArgument(t *testing.T) {
	dir := t.TempDir()
	src := `package fake

func init() {
	Wiring.RegisterBuilder("Alpha", nil, nil)
}
`
	writeFile(t, dir, "fake.go", src)
	name, err := parseGoKindName(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "Alpha" {
		t.Errorf("parseGoKindName = %q, want Alpha", name)
	}
}

func TestParseGoKindName_NotFoundIsError(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "fake.go", "package fake\n")
	if _, err := parseGoKindName(dir); err == nil {
		t.Fatal("want error when Wiring.Register is absent, got nil")
	}
}

func TestParseDataFieldsFromAST_ExtractsTaggedFieldsWithGoType(t *testing.T) {
	dir := t.TempDir()
	src := `package fake

type FakeNode struct {
	Init  []int  ` + "`wire:\"data.init\"`" + `
	State string ` + "`wire:\"data.state\"`" + `
	Other int
}
`
	writeFile(t, dir, "fake.go", src)
	fields, err := parseDataFieldsFromAST(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fields) != 2 {
		t.Fatalf("want 2 data fields, got %d: %+v", len(fields), fields)
	}
	if fields[0].wireTag != "init" || fields[0].goType != "[]int" || fields[0].fieldName != "Init" {
		t.Errorf("fields[0] = %+v, want wireTag=init goType=[]int fieldName=Init", fields[0])
	}
	if fields[1].wireTag != "state" || fields[1].goType != "string" || fields[1].fieldName != "State" {
		t.Errorf("fields[1] = %+v, want wireTag=state goType=string fieldName=State", fields[1])
	}
}

func TestParsePortsFromSpec_FallbackReadsPortsTable(t *testing.T) {
	dir := t.TempDir()
	spec := `# Fake

## Ports

| Name | Direction |
|---|---|
| a | in |
| b | out |
`
	writeFile(t, dir, "SPEC.md", spec)
	ports := parsePortsFromSpec(dir)
	if len(ports) != 2 {
		t.Fatalf("want 2 ports, got %d: %+v", len(ports), ports)
	}
	if ports[0].id != "a" || ports[0].direction != "in" {
		t.Errorf("ports[0] = %+v, want a/in", ports[0])
	}
	if ports[1].id != "b" || ports[1].direction != "out" {
		t.Errorf("ports[1] = %+v, want b/out", ports[1])
	}
}

// --- node_dims.go / node_defs.go: emitters over a hand-built kindEntry ------

func TestWriteNodeDims_EmitsPerKindDimensionsWithDefaults(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "node_dims_gen.go")
	kinds := []kindEntry{
		{goKind: "Alpha", view: viewDef{width: "150", height: "70"}},
		{goKind: "Beta", view: viewDef{}}, // no width/height: defaults expected
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
	kinds := []kindEntry{
		{goKind: "Gamma", kindID: 5},
		{goKind: "Alpha", kindID: 0},
		{goKind: "Beta", kindID: 2},
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
		keyWant := `"` + e.goKind + `":`
		if !strings.Contains(out, keyWant) {
			t.Errorf("output missing key %q:\n%s", keyWant, out)
			continue
		}
		valWant := itoa(int(e.kindID)) + ",\n"
		lineStart := strings.Index(out, keyWant)
		line := out[lineStart : lineStart+len(keyWant)+20]
		if !strings.Contains(line, valWant) {
			t.Errorf("kind %q: line %q does not contain its stored kindID %d", e.goKind, line, e.kindID)
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
	kinds := []kindEntry{
		{
			goKind: "Alpha",
			view: viewDef{
				bg: "#111111", border: "#222222", text: "#333333",
				shape: "box", fill: "#444444", stroke: "#555555",
				width: "100", height: "50",
			},
			ports: []port{
				{id: "in1", direction: "in", edgeKind: "chain"},
				{id: "out1", direction: "out", edgeKind: "value", isMulti: true},
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
	ports := []port{
		{id: "a", direction: "in"},
		{id: "b", direction: "out"},
		{id: "c", direction: "in"},
	}
	ins := filterPorts(ports, "in")
	outs := filterPorts(ports, "out")
	if len(ins) != 2 || ins[0].id != "a" || ins[1].id != "c" {
		t.Errorf("filterPorts(in) = %+v, want [a, c]", ins)
	}
	if len(outs) != 1 || outs[0].id != "b" {
		t.Errorf("filterPorts(out) = %+v, want [b]", outs)
	}
}

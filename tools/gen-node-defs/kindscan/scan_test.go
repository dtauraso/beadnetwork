package kindscan

import (
	"os"
	"path/filepath"
	"testing"
)

// scan_test.go — the KindEntry pipeline: parsing a node kind's Go source and
// SPEC.md into a KindEntry (ast_ports.go/spec_md.go/ast_kind.go).

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
	byID := map[string]Port{}
	for _, p := range ports {
		byID[p.ID] = p
	}
	if byID["In1"].Direction != "in" {
		t.Errorf("In1 direction = %q, want in", byID["In1"].Direction)
	}
	if byID["Out1"].Direction != "out" {
		t.Errorf("Out1 direction = %q, want out", byID["Out1"].Direction)
	}
	if p, ok := byID["Multi"]; !ok || p.Direction != "out" || !p.IsMulti {
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
	if view.Bg != "#123456" || view.Border != "#654321" {
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
	if fields[0].WireTag != "init" || fields[0].GoType != "[]int" || fields[0].FieldName != "Init" {
		t.Errorf("fields[0] = %+v, want wireTag=init goType=[]int fieldName=Init", fields[0])
	}
	if fields[1].WireTag != "state" || fields[1].GoType != "string" || fields[1].FieldName != "State" {
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
	if ports[0].ID != "a" || ports[0].Direction != "in" {
		t.Errorf("ports[0] = %+v, want a/in", ports[0])
	}
	if ports[1].ID != "b" || ports[1].Direction != "out" {
		t.Errorf("ports[1] = %+v, want b/out", ports[1])
	}
}

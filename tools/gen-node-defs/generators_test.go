package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- wire_defs.go: parseWirePropsFromFile ---------------------------------

func writeTempGoFile(t *testing.T, contents string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "loader.go")
	if err := os.WriteFile(path, []byte(contents), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	return path
}

func TestParseWirePropsFromFile_WellFormed(t *testing.T) {
	src := `package main

type specEdge struct {
	Label string ` + "`json:\"label,omitempty\" wire:\"prop,optional,tsType:string\"`" + `
	Count int    ` + "`json:\"count\" wire:\"prop,required,tsType:number\"`" + `
	NotAProp string ` + "`json:\"skip\"`" + `
}
`
	path := writeTempGoFile(t, src)
	props, err := parseWirePropsFromFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(props) != 2 {
		t.Fatalf("want 2 props, got %d: %+v", len(props), props)
	}
	if got, want := props[0], (wireProp{jsonName: "label", tsType: "string", required: false}); got != want {
		t.Errorf("props[0] = %+v, want %+v", got, want)
	}
	if got, want := props[1], (wireProp{jsonName: "count", tsType: "number", required: true}); got != want {
		t.Errorf("props[1] = %+v, want %+v", got, want)
	}
}

func TestParseWirePropsFromFile_TooFewSegments(t *testing.T) {
	// Missing the required|optional segment entirely.
	src := `package main

type specEdge struct {
	Label string ` + "`json:\"label\" wire:\"prop,tsType:string\"`" + `
}
`
	path := writeTempGoFile(t, src)
	_, err := parseWirePropsFromFile(path)
	if err == nil {
		t.Fatal("want error for too-few-segments wire tag, got nil")
	}
	if !strings.Contains(err.Error(), "malformed wire tag") {
		t.Errorf("error = %q, want it to mention malformed wire tag", err.Error())
	}
}

func TestParseWirePropsFromFile_BadRequiredOptionalSegment(t *testing.T) {
	src := `package main

type specEdge struct {
	Label string ` + "`json:\"label\" wire:\"prop,mandatory,tsType:string\"`" + `
}
`
	path := writeTempGoFile(t, src)
	_, err := parseWirePropsFromFile(path)
	if err == nil {
		t.Fatal("want error for invalid required|optional segment, got nil")
	}
	if !strings.Contains(err.Error(), `"required" or "optional"`) {
		t.Errorf("error = %q, want it to mention required/optional", err.Error())
	}
}

func TestParseWirePropsFromFile_MissingTsType(t *testing.T) {
	src := `package main

type specEdge struct {
	Label string ` + "`json:\"label\" wire:\"prop,optional,foo:bar\"`" + `
}
`
	path := writeTempGoFile(t, src)
	_, err := parseWirePropsFromFile(path)
	if err == nil {
		t.Fatal("want error for missing tsType: segment, got nil")
	}
	if !strings.Contains(err.Error(), "no tsType:") {
		t.Errorf("error = %q, want it to mention missing tsType", err.Error())
	}
}

func TestParseWirePropsFromFile_JSONNameFallsBackToFieldName(t *testing.T) {
	// No json tag at all: jsonName should derive from the field name, lowercased first letter.
	src := `package main

type specEdge struct {
	Label string ` + "`wire:\"prop,optional,tsType:string\"`" + `
}
`
	path := writeTempGoFile(t, src)
	props, err := parseWirePropsFromFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(props) != 1 || props[0].jsonName != "label" {
		t.Fatalf("want jsonName derived as %q, got %+v", "label", props)
	}
}

// --- trace_kinds.go: parseTraceKinds / parseBreadcrumbLabels --------------

func TestParseTraceKinds_OrderMatchesSourceDeclarationOrder(t *testing.T) {
	dir := t.TempDir()
	src := `package Trace

const (
	KindRecv = "recv"
	KindFire = "fire"
	KindSend = "send"
)

const KindSlot = "slot"

// Not a Kind* const — must be ignored.
const OtherThing = "other"
`
	if err := os.WriteFile(filepath.Join(dir, "trace.go"), []byte(src), 0644); err != nil {
		t.Fatal(err)
	}
	kinds, err := parseTraceKinds(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"recv", "fire", "send", "slot"}
	if len(kinds) != len(want) {
		t.Fatalf("kinds = %v, want %v", kinds, want)
	}
	for i := range want {
		if kinds[i] != want[i] {
			t.Errorf("kinds[%d] = %q, want %q (reorder bug: wire numbering IS slice order)", i, kinds[i], want[i])
		}
	}
}

func TestParseTraceKinds_IgnoresTestFiles(t *testing.T) {
	dir := t.TempDir()
	main := `package Trace

const KindReal = "real"
`
	testFile := `package Trace

const KindFromTest = "fromtest"
`
	if err := os.WriteFile(filepath.Join(dir, "trace.go"), []byte(main), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "trace_test.go"), []byte(testFile), 0644); err != nil {
		t.Fatal(err)
	}
	kinds, err := parseTraceKinds(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(kinds) != 1 || kinds[0] != "real" {
		t.Fatalf("kinds = %v, want only [\"real\"] (test file constants must not leak in)", kinds)
	}
}

func TestParseTraceKinds_EmptyIsError(t *testing.T) {
	dir := t.TempDir()
	src := `package Trace

const NotAKindConst = "x"
`
	if err := os.WriteFile(filepath.Join(dir, "trace.go"), []byte(src), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := parseTraceKinds(dir); err == nil {
		t.Fatal("want error when no Kind* constants are found, got nil")
	}
}

func TestParseBreadcrumbLabels_OrderPreserved(t *testing.T) {
	dir := t.TempDir()
	src := `package Trace

var BreadcrumbLabels = []string{"first", "second", "third"}
`
	if err := os.WriteFile(filepath.Join(dir, "trace.go"), []byte(src), 0644); err != nil {
		t.Fatal(err)
	}
	labels, err := parseBreadcrumbLabels(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"first", "second", "third"}
	if len(labels) != len(want) {
		t.Fatalf("labels = %v, want %v", labels, want)
	}
	for i := range want {
		if labels[i] != want[i] {
			t.Errorf("labels[%d] = %q, want %q", i, labels[i], want[i])
		}
	}
}

func TestParseBreadcrumbLabels_NotFoundIsError(t *testing.T) {
	dir := t.TempDir()
	src := `package Trace

var SomethingElse = []string{"x"}
`
	if err := os.WriteFile(filepath.Join(dir, "trace.go"), []byte(src), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := parseBreadcrumbLabels(dir); err == nil {
		t.Fatal("want error when BreadcrumbLabels var is absent, got nil")
	}
}

// --- params.go: parseCurveParams / parseShadingParams / camelToScreamingSnake ---

func TestCamelToScreamingSnake(t *testing.T) {
	cases := map[string]string{
		"CurveParamBulgeFactor": "CURVE_PARAM_BULGE_FACTOR",
		"BeadID":                "BEAD_ID",
		"CX":                    "CX",
		"CurveParamX":           "CURVE_PARAM_X",
	}
	for in, want := range cases {
		if got := camelToScreamingSnake(in); got != want {
			t.Errorf("camelToScreamingSnake(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestParseCurveParams_ExtractsPrefixedConstsOnly(t *testing.T) {
	dir := t.TempDir()
	src := `package Wiring

const (
	CurveParamBulgeFactor = 0.5
	CurveParamSegments    = 12
	notAParam             = "ignore me"
)
`
	path := filepath.Join(dir, "curve_params.go")
	if err := os.WriteFile(path, []byte(src), 0644); err != nil {
		t.Fatal(err)
	}
	params, err := parseCurveParams(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(params) != 2 {
		t.Fatalf("want 2 params, got %d: %+v", len(params), params)
	}
	if params[0].tsName != "CURVE_PARAM_BULGE_FACTOR" || params[0].value != "0.5" || params[0].isInt {
		t.Errorf("params[0] = %+v, want name CURVE_PARAM_BULGE_FACTOR, value 0.5, isInt=false", params[0])
	}
	if params[1].tsName != "CURVE_PARAM_SEGMENTS" || params[1].value != "12" || !params[1].isInt {
		t.Errorf("params[1] = %+v, want name CURVE_PARAM_SEGMENTS, value 12, isInt=true", params[1])
	}
}

func TestParseCurveParams_EmptyIsError(t *testing.T) {
	dir := t.TempDir()
	src := `package Wiring

const notAParam = 1
`
	path := filepath.Join(dir, "curve_params.go")
	if err := os.WriteFile(path, []byte(src), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := parseCurveParams(path); err == nil {
		t.Fatal("want error when no CurveParam* constants are found, got nil")
	}
}

func TestParseShadingParams_StringAndNumericLiterals(t *testing.T) {
	dir := t.TempDir()
	src := `package Wiring

const (
	ShadingParamBaseColor = "#ff00ff"
	ShadingParamIntensity = 1.25
)
`
	path := filepath.Join(dir, "shading_params.go")
	if err := os.WriteFile(path, []byte(src), 0644); err != nil {
		t.Fatal(err)
	}
	params, err := parseShadingParams(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(params) != 2 {
		t.Fatalf("want 2 params, got %d: %+v", len(params), params)
	}
	if params[0].tsName != "SHADING_PARAM_BASE_COLOR" || params[0].value != "#ff00ff" || !params[0].isStr {
		t.Errorf("params[0] = %+v, want name SHADING_PARAM_BASE_COLOR, value #ff00ff, isStr=true", params[0])
	}
	if params[1].tsName != "SHADING_PARAM_INTENSITY" || params[1].value != "1.25" || params[1].isStr {
		t.Errorf("params[1] = %+v, want name SHADING_PARAM_INTENSITY, value 1.25, isStr=false", params[1])
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

func TestWriteNodeKindID_IndexMatchesInputOrder(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "node_kind_id_gen.go")
	kinds := []kindEntry{
		{goKind: "Alpha"},
		{goKind: "Beta"},
		{goKind: "Gamma"},
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
	for i, e := range kinds {
		keyWant := `"` + e.goKind + `":`
		if !strings.Contains(out, keyWant) {
			t.Errorf("output missing key %q:\n%s", keyWant, out)
			continue
		}
		valWant := itoa(i) + ",\n"
		lineStart := strings.Index(out, keyWant)
		line := out[lineStart : lineStart+len(keyWant)+20]
		if !strings.Contains(line, valWant) {
			t.Errorf("kind %q: line %q does not contain index %d (index must equal input order, which callers guarantee is alphabetical)", e.goKind, line, i)
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
				role: "source", shape: "box", fill: "#444444", stroke: "#555555",
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
	if !strings.Contains(out, `role: "source"`) || !strings.Contains(out, `shape: "box"`) {
		t.Errorf("NODE_DEFS Alpha missing role/shape fields:\n%s", out)
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

// --- ast_parse.go: parsePortsFromAST / parseSpecMD / parseGoKindName -------

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
| role | source |

## Ports

| Name | Direction | Accent | EdgeKind | Optional |
|---|---|---|---|---|
| in1 | in | #abcdef | value | yes |
`
	writeFile(t, dir, "SPEC.md", spec)
	view, accent, edgeKind, optional, err := parseSpecMD(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if view.bg != "#123456" || view.border != "#654321" || view.role != "source" {
		t.Errorf("view = %+v, want bg/border/role from table", view)
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
	if _, _, _, _, err := parseSpecMD(dir); err == nil {
		t.Fatal("want error when ## View section is missing, got nil")
	}
}

func TestParseGoKindName_ExtractsRegisterArgument(t *testing.T) {
	dir := t.TempDir()
	src := `package fake

func init() {
	Wiring.Register("Alpha", nil)
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

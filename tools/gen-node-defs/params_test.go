package main

import (
	"os"
	"path/filepath"
	"testing"
)

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
	params, err := parseShadingParams(dir, path)
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

// TestParseShadingParams_EvaluatesCrossPackageExpression is the regression guard for the
// actual bug this change closes: a ShadingParam* const written as an EXPRESSION —
// including one that references a const in another package, the exact shape
// ShadingParamBeadRadius uses for wire.BeadTorusOuterR — must still show up in the
// generated TS mirror with the correct evaluated value. Before constexpr.go, parseShadingParams
// only recognized a plain *ast.BasicLit and silently DROPPED anything else (see the git
// history of ShadingParamBeadRadius, which was written as a hand-computed literal for
// exactly this reason); this test fails loudly the same way a real regression would — the
// derived const would vanish from params, not merely mismatch — if that ever comes back.
func TestParseShadingParams_EvaluatesCrossPackageExpression(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/fixture\n\ngo 1.25.0\n"), 0644); err != nil {
		t.Fatal(err)
	}
	otherDir := filepath.Join(root, "other")
	if err := os.Mkdir(otherDir, 0755); err != nil {
		t.Fatal(err)
	}
	otherSrc := `package other

const Base = 8.0
`
	if err := os.WriteFile(filepath.Join(otherDir, "other.go"), []byte(otherSrc), 0644); err != nil {
		t.Fatal(err)
	}
	consumerDir := filepath.Join(root, "consumer")
	if err := os.Mkdir(consumerDir, 0755); err != nil {
		t.Fatal(err)
	}
	consumerSrc := `package consumer

import (
	other "example.com/fixture/other"
)

const ShadingParamRatio = 0.12

const ShadingParamDerived = other.Base / (1 + ShadingParamRatio)
`
	path := filepath.Join(consumerDir, "shading_params.go")
	if err := os.WriteFile(path, []byte(consumerSrc), 0644); err != nil {
		t.Fatal(err)
	}
	params, err := parseShadingParams(root, path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(params) != 2 {
		t.Fatalf("want 2 params (Ratio literal + Derived expression), got %d: %+v", len(params), params)
	}
	want := "7.142857142857143" // 8.0 / 1.12, formatted the way strconv.FormatFloat(f, 'g', -1, 64) does
	if params[1].tsName != "SHADING_PARAM_DERIVED" || params[1].value != want || params[1].isStr {
		t.Errorf("params[1] = %+v, want name SHADING_PARAM_DERIVED, value %s, isStr=false", params[1], want)
	}
}

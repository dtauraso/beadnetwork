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

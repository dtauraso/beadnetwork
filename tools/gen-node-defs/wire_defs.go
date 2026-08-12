// wire-defs.ts pipeline: parses wire:"prop,..." tags off specEdge in
// nodes/Wiring/loadspec/topo_spec.go (this file) and emits WIRE_PROPS/WireProps from the
// parsed result (the sibling wire_defs_write.go).
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// wireProp represents one wire:"prop,..." tagged field on specEdge.
type wireProp struct {
	jsonName string // from json:"..." tag
	tsType   string // from tsType:... in wire tag
	required bool   // false if "optional", true if "required"
}

// parseWirePropsFromFile parses wire:"prop,..." tags on fields of specEdge
// in the given Go source file and returns them in declaration order.
func parseWirePropsFromFile(filePath string) ([]wireProp, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filePath, nil, 0)
	if err != nil {
		return nil, err
	}
	var props []wireProp
	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok || typeSpec.Name.Name != "specEdge" {
				continue
			}
			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}
			for _, field := range structType.Fields.List {
				if field.Tag == nil {
					continue
				}
				rawTag := strings.Trim(field.Tag.Value, "`")
				// Extract wire tag value (text after `wire:"` up to the closing quote).
				_, wireVal, hasWire := strings.Cut(rawTag, `wire:"`)
				if !hasWire {
					continue
				}
				wireVal, _, _ = strings.Cut(wireVal, `"`)
				if !strings.HasPrefix(wireVal, "prop,") {
					continue
				}
				// Parse segments: prop, optional|required, tsType:X
				//
				// A malformed prop tag is an ERROR, not a skip. These `continue`s used to be
				// silent: a field tagged wire:"prop,tsType:string" (no optional|required
				// segment) was dropped, the generator still printed "wrote wire-defs.ts
				// (1 entries)" and exited 0, and the prop simply never reached TS. The tag
				// looked correct in review and nothing anywhere said otherwise. If a field
				// says it is a wire prop, either emit it or say why not.
				fieldName := "<anonymous>"
				if len(field.Names) > 0 {
					fieldName = field.Names[0].Name
				}
				segments := strings.Split(wireVal, ",")
				if len(segments) < 3 {
					return nil, fmt.Errorf(
						"specEdge.%s: malformed wire tag %q — want prop,<optional|required>,tsType:<T> (got %d segments, need at least 3)",
						fieldName, wireVal, len(segments))
				}
				if segments[1] != "required" && segments[1] != "optional" {
					return nil, fmt.Errorf(
						"specEdge.%s: wire tag %q has second segment %q — want \"required\" or \"optional\"",
						fieldName, wireVal, segments[1])
				}
				required := segments[1] == "required"
				tsType := ""
				for _, seg := range segments[2:] {
					if after, found := strings.CutPrefix(seg, "tsType:"); found {
						tsType = after
					}
				}
				if tsType == "" {
					return nil, fmt.Errorf(
						"specEdge.%s: wire tag %q has no tsType:<T> segment", fieldName, wireVal)
				}
				// Extract json name.
				jsonName := ""
				_, after, found := strings.Cut(rawTag, `json:"`)
				if found {
					jsonName, _, _ = strings.Cut(after, `"`)
					// Strip omitempty etc.
					jsonName, _, _ = strings.Cut(jsonName, ",")
				}
				if jsonName == "" && len(field.Names) > 0 {
					jsonName = strings.ToLower(field.Names[0].Name[:1]) + field.Names[0].Name[1:]
				}
				props = append(props, wireProp{jsonName: jsonName, tsType: tsType, required: required})
			}
		}
	}
	return props, nil
}

package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

type wireProp struct {
	jsonName string
	tsType   string
	required bool
}

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

				_, wireVal, hasWire := strings.Cut(rawTag, `wire:"`)
				if !hasWire {
					continue
				}
				wireVal, _, _ = strings.Cut(wireVal, `"`)
				if !strings.HasPrefix(wireVal, "prop,") {
					continue
				}

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

				jsonName := ""
				_, after, found := strings.Cut(rawTag, `json:"`)
				if found {
					jsonName, _, _ = strings.Cut(after, `"`)

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

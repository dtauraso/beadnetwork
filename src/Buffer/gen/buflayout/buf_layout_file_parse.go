package buflayout

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

func parseBufferLayoutFile(layoutPath string) (BufLayoutSchema, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, layoutPath, nil, 0)
	if err != nil {
		return BufLayoutSchema{}, err
	}

	var schema BufLayoutSchema

	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			if d.Tok == token.CONST {
				for _, spec := range d.Specs {
					vs, ok := spec.(*ast.ValueSpec)
					if !ok {
						continue
					}
					for i, nm := range vs.Names {
						if i >= len(vs.Values) {
							continue
						}
						lit, ok := vs.Values[i].(*ast.BasicLit)
						if !ok || lit.Kind != token.INT {
							continue
						}
						var ival int
						fmt.Sscan(lit.Value, &ival)
						switch {
						case nm.Name == "BufLayoutVersion":
							schema.version = ival
						case nm.Name == "BufInteriorSlotsPerNode":
							schema.interiorSlotsPerNode = ival
						}
					}
				}
			} else if d.Tok == token.TYPE {
				for _, spec := range d.Specs {
					ts, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}
					if !strings.HasPrefix(ts.Name.Name, "bufLayout") {
						continue
					}
					st, ok := ts.Type.(*ast.StructType)
					if !ok {
						continue
					}
					blockName := ts.Name.Name[len("bufLayout"):]
					block := bufBlock{name: blockName}
					offset := 0
					for _, field := range st.Fields.List {
						if field.Tag == nil || len(field.Names) == 0 {
							continue
						}
						rawTag := strings.Trim(field.Tag.Value, "`")
						_, after, ok := strings.Cut(rawTag, `buf:"`)
						if !ok {
							continue
						}
						bufType, _, _ := strings.Cut(after, `"`)
						sz, err := bufTypeSize(bufType)
						if err != nil {
							return BufLayoutSchema{}, fmt.Errorf("block %s field %s: %w", blockName, field.Names[0].Name, err)
						}
						block.columns = append(block.columns, bufCol{
							name:    field.Names[0].Name,
							bufType: bufType,
							offset:  offset,
						})
						offset += sz
					}
					block.stride = offset
					schema.Blocks = append(schema.Blocks, block)
				}
			}
		}
	}

	return schema, nil
}

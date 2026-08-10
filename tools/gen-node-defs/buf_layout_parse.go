package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// bufCol is one column within a buffer block (from a buf: struct tag).
type bufCol struct {
	name    string // Go field name, e.g. "X"
	bufType string // "f32" | "i32" | "u32" | "u8"
	offset  int    // byte offset within one row (packed, no padding)
}

// bufBlock is one column block from a bufLayout* struct definition.
type bufBlock struct {
	name    string   // e.g. "Bead", "Node", "Edge", "Camera", "Overlay"
	columns []bufCol // in declaration order
	stride  int      // total bytes per row
}

// bufLayoutSchema is the parsed content of Buffer/layout.go.
type bufLayoutSchema struct {
	version              int
	blocks               []bufBlock
	interiorSlotsPerNode int
}

// parseBufferLayout reads Buffer/layout.go and returns the schema:
// - BufLayoutVersion const → version
// - bufLayout* struct types in source order → blocks with columns + strides
func parseBufferLayout(layoutPath string) (bufLayoutSchema, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, layoutPath, nil, 0)
	if err != nil {
		return bufLayoutSchema{}, err
	}

	var schema bufLayoutSchema

	// Walk declarations in source order to preserve relative ordering of consts
	// and struct types (they are interleaved intentionally — version first, then
	// blocks).
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
							return bufLayoutSchema{}, fmt.Errorf("block %s field %s: %w", blockName, field.Names[0].Name, err)
						}
						block.columns = append(block.columns, bufCol{
							name:    field.Names[0].Name,
							bufType: bufType,
							offset:  offset,
						})
						offset += sz
					}
					block.stride = offset
					schema.blocks = append(schema.blocks, block)
				}
			}
		}
	}

	if schema.version == 0 {
		return bufLayoutSchema{}, fmt.Errorf("BufLayoutVersion const not found in %s", layoutPath)
	}
	if schema.interiorSlotsPerNode == 0 {
		return bufLayoutSchema{}, fmt.Errorf("BufInteriorSlotsPerNode const not found in %s", layoutPath)
	}
	if len(schema.blocks) == 0 {
		return bufLayoutSchema{}, fmt.Errorf("no bufLayout* struct types found in %s", layoutPath)
	}
	return schema, nil
}

// buildBufFingerprint builds a deterministic fingerprint string from the schema.
// Both generated files embed this as a comment; the parity guard greps and compares.
func buildBufFingerprint(schema bufLayoutSchema) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("version:%d", schema.version))
	parts = append(parts, fmt.Sprintf("interiorSlotsPerNode:%d", schema.interiorSlotsPerNode))
	for _, blk := range schema.blocks {
		var cols []string
		for _, c := range blk.columns {
			cols = append(cols, fmt.Sprintf("%s:%s:%d", c.name, c.bufType, c.offset))
		}
		parts = append(parts, fmt.Sprintf("block:%s[%s]:stride:%d", blk.name, strings.Join(cols, ","), blk.stride))
	}
	return strings.Join(parts, "|")
}

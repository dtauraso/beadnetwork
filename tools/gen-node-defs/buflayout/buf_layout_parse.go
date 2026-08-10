package buflayout

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"sort"
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

// BufLayoutSchema is the parsed content of Buffer/layout.go.
type BufLayoutSchema struct {
	version              int
	Blocks               []bufBlock
	interiorSlotsPerNode int
}

// bufBlockOrder is the FIXED, canonical ordering of bufLayout* blocks that
// BUF_LAYOUT_FINGERPRINT was built from back when they all lived in source
// order in one file (Buffer/layout.go). Splitting the schema across several
// files (Buffer/layout_*.go, one struct per file, scanned by
// ParseBufferLayoutDir below) makes filesystem scan order no longer the
// source of block order, so this list is what keeps the fingerprint byte-
// identical across that split: every block found is re-sorted into this
// order before the fingerprint is built. A block name missing from this list
// is a loud error (fatalf), never a silent append — the same "scan, don't
// hardcode a path, but never guess an order" shape as
// parseInputLayoutFingerprintDir (memory/feedback_guards_hardcoding_single_file_break_on_split.md).
var bufBlockOrder = []string{
	"Node", "ChainBead", "Interior", "Edge", "Camera", "Overlay", "Scene", "Event",
}

// ParseBufferLayoutDir scans every *.go file directly in dir (Buffer/) for the
// BufLayoutVersion/BufInteriorSlotsPerNode consts and bufLayout* struct types,
// and returns the accumulated schema. Scanning the directory rather than a
// single named file is what lets Buffer/layout.go's schema be split into
// several files, one struct per file, without the generator losing track of
// any of them (memory/feedback_guards_hardcoding_single_file_break_on_split.md,
// the same shape as parseInputLayoutFingerprintDir in input_layout.go). Block
// ORDER is then normalized via bufBlockOrder (above), independent of scan
// order, so the fingerprint stays exactly what it was when everything lived
// in declaration order in one file.
func ParseBufferLayoutDir(dir string) (BufLayoutSchema, error) {
	paths, err := filepath.Glob(filepath.Join(dir, "*.go"))
	if err != nil {
		return BufLayoutSchema{}, err
	}
	sort.Strings(paths)

	var schema BufLayoutSchema
	var blocks []bufBlock
	for _, p := range paths {
		if strings.HasSuffix(p, "_test.go") {
			continue
		}
		fileSchema, err := parseBufferLayoutFile(p)
		if err != nil {
			return BufLayoutSchema{}, fmt.Errorf("%s: %w", p, err)
		}
		if fileSchema.version != 0 {
			if schema.version != 0 && schema.version != fileSchema.version {
				return BufLayoutSchema{}, fmt.Errorf("%s: BufLayoutVersion %d conflicts with earlier %d", p, fileSchema.version, schema.version)
			}
			schema.version = fileSchema.version
		}
		if fileSchema.interiorSlotsPerNode != 0 {
			schema.interiorSlotsPerNode = fileSchema.interiorSlotsPerNode
		}
		blocks = append(blocks, fileSchema.Blocks...)
	}

	if schema.version == 0 {
		return BufLayoutSchema{}, fmt.Errorf("BufLayoutVersion const not found under %s", dir)
	}
	if schema.interiorSlotsPerNode == 0 {
		return BufLayoutSchema{}, fmt.Errorf("BufInteriorSlotsPerNode const not found under %s", dir)
	}
	if len(blocks) == 0 {
		return BufLayoutSchema{}, fmt.Errorf("no bufLayout* struct types found under %s", dir)
	}

	byName := map[string]bufBlock{}
	for _, b := range blocks {
		byName[b.name] = b
	}
	if len(byName) != len(blocks) {
		return BufLayoutSchema{}, fmt.Errorf("duplicate bufLayout block name found under %s", dir)
	}
	for _, name := range bufBlockOrder {
		b, ok := byName[name]
		if !ok {
			return BufLayoutSchema{}, fmt.Errorf("bufBlockOrder names block %q but no bufLayout%s struct was found under %s", name, name, dir)
		}
		schema.Blocks = append(schema.Blocks, b)
		delete(byName, name)
	}
	if len(byName) != 0 {
		var leftover []string
		for name := range byName {
			leftover = append(leftover, name)
		}
		sort.Strings(leftover)
		return BufLayoutSchema{}, fmt.Errorf("bufLayout block(s) %v found under %s but missing from bufBlockOrder — add them there", leftover, dir)
	}

	return schema, nil
}

// parseBufferLayoutFile parses one file for its BufLayoutVersion/
// BufInteriorSlotsPerNode consts (zero value if this file declares neither)
// and its bufLayout* struct types (nil if this file declares none) — the
// per-file half of ParseBufferLayoutDir's directory scan.
func parseBufferLayoutFile(layoutPath string) (BufLayoutSchema, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, layoutPath, nil, 0)
	if err != nil {
		return BufLayoutSchema{}, err
	}

	var schema BufLayoutSchema

	// Walk declarations in source order to preserve relative ordering of consts
	// and struct types (they are interleaved intentionally — version first, then
	// blocks — within a file that declares more than one).
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

// buildBufFingerprint builds a deterministic fingerprint string from the schema.
// Both generated files embed this as a comment; the parity guard greps and compares.
func buildBufFingerprint(schema BufLayoutSchema) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("version:%d", schema.version))
	parts = append(parts, fmt.Sprintf("interiorSlotsPerNode:%d", schema.interiorSlotsPerNode))
	for _, blk := range schema.Blocks {
		var cols []string
		for _, c := range blk.columns {
			cols = append(cols, fmt.Sprintf("%s:%s:%d", c.name, c.bufType, c.offset))
		}
		parts = append(parts, fmt.Sprintf("block:%s[%s]:stride:%d", blk.name, strings.Join(cols, ","), blk.stride))
	}
	return strings.Join(parts, "|")
}

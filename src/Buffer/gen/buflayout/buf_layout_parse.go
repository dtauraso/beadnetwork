package buflayout

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type bufCol struct {
	name    string
	bufType string
	offset  int
}

type bufBlock struct {
	name    string
	columns []bufCol
	stride  int

	dir string
}

type BufLayoutSchema struct {
	version              int
	Blocks               []bufBlock
	interiorSlotsPerNode int
}

var bufBlockOrder = []string{
	"Node", "Interior", "Edge", "EdgeBead", "NodeRingPoint", "BeadRingPoint", "TiltArrow", "ChannelVector", "Scene",
	"Recv", "Fire", "Send", "Arrive", "Breadcrumb",
	"SpeedPanel", "TiltPanel", "OverlaysPill", "FitChip",
	"PointerTarget",
}

func ParseBufferLayoutTree(root string) (BufLayoutSchema, error) {
	var paths []string
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() { // path-resolution-ok: a generator walking its own source tree, not a scene path
			switch d.Name() {
			case "node_modules", "out", ".git", ".probe", ".wirefold-cache":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(p, ".go") || strings.HasSuffix(p, "_test.go") {
			return nil
		}
		b, readErr := os.ReadFile(p)
		if readErr != nil {
			return readErr
		}
		if bytes.Contains(b, []byte("bufLayout")) || bytes.Contains(b, []byte("BufLayoutVersion")) {
			paths = append(paths, p)
		}
		return nil
	})
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
		for _, b := range fileSchema.Blocks {
			b.dir = filepath.Dir(p)
			blocks = append(blocks, b)
		}
	}

	if schema.version == 0 {
		return BufLayoutSchema{}, fmt.Errorf("BufLayoutVersion const not found under %s", root)
	}
	if schema.interiorSlotsPerNode == 0 {
		return BufLayoutSchema{}, fmt.Errorf("BufInteriorSlotsPerNode const not found under %s", root)
	}
	if len(blocks) == 0 {
		return BufLayoutSchema{}, fmt.Errorf("no bufLayout* struct types found under %s", root)
	}

	byName := map[string]bufBlock{}
	for _, b := range blocks {
		byName[b.name] = b
	}
	if len(byName) != len(blocks) {
		return BufLayoutSchema{}, fmt.Errorf("duplicate bufLayout block name found under %s", root)
	}
	for _, name := range bufBlockOrder {
		b, ok := byName[name]
		if !ok {
			return BufLayoutSchema{}, fmt.Errorf("bufBlockOrder names block %q but no bufLayout%s struct was found under %s", name, name, root)
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
		return BufLayoutSchema{}, fmt.Errorf("bufLayout block(s) %v found under %s but missing from bufBlockOrder — add them there", leftover, root)
	}

	return schema, nil
}

func fnv1aHash(s string) uint32 {
	const offset32 = 2166136261
	const prime32 = 16777619
	h := uint32(offset32)
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= prime32
	}
	return h
}

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

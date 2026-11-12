package Builder

import (
	"os"
	"path/filepath"
	"strings"
)

const OpFolderFold = "folder-fold"

func (m *Model) ToggleFolder(rel string) {
	if m.FoldersOpen == nil {
		m.FoldersOpen = map[string]bool{}
	}
	m.FoldersOpen[rel] = !m.folderOpen(rel)
}

func (m Model) folderOpen(rel string) bool {
	if open, ok := m.FoldersOpen[rel]; ok {
		return open
	}
	return rel == "."
}

func folderGroups(m Model, rel string, depth int) []Group {
	entries, err := os.ReadDir(filepath.Join(m.Root, rel))
	title := filepath.Base(m.Root) + "/"
	if rel != "." {
		title = filepath.Base(rel) + "/"
	}
	open := m.folderOpen(rel)
	g := Group{Title: strings.Repeat("    ", depth) + foldTitle(!open, title), Fold: Act{Op: OpFolderFold, Node: rel}}
	if err != nil {
		g.Rows = [][]Cell{{Text(err.Error())}}
		return []Group{g}
	}
	if !open {
		return []Group{g}
	}
	var files []string
	var dirs []string
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".") {
			continue
		}
		if e.IsDir() { // path-resolution-ok: listing the sequence folder as it is on disk
			dirs = append(dirs, e.Name())
		} else {
			files = append(files, e.Name())
		}
	}
	if len(files) > 0 {
		g.Rows = [][]Cell{{Text(strings.Repeat("    ", depth+1) + strings.Join(files, "  "))}}
		g.Wraps = true
	}
	out := []Group{g}
	for _, d := range dirs {
		out = append(out, folderGroups(m, filepath.Join(rel, d), depth+1)...)
	}
	return out
}

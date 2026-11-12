package Builder

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	OpFolderFold = "folder-fold"
	OpFolderGo   = "folder-go"
	OpFolderNew  = "folder-new"
)

func (m *Model) ToggleFolder(dir string) {
	if m.FoldersOpen == nil {
		m.FoldersOpen = map[string]bool{}
	}
	m.FoldersOpen[dir] = !m.folderOpen(dir)
}

func (m Model) cwd() string {
	if m.Cwd != "" {
		return m.Cwd
	}
	return m.Root
}

func (m Model) folderOpen(dir string) bool {
	if open, ok := m.FoldersOpen[dir]; ok {
		return open
	}
	return dir == m.cwd()
}

func (m *Model) GoFolder(dir string) {
	m.Cwd = filepath.Clean(dir)
}

func nextNumber(dir string) string {
	entries, _ := os.ReadDir(dir)
	n := 0
	for _, e := range entries {
		if k, err := strconv.Atoi(e.Name()); err == nil && e.IsDir() { // path-resolution-ok: the next free number among a folder's numbered folders
			n = max(n, k)
		}
	}
	return strconv.Itoa(n + 1)
}

func (m *Model) NewFolder(dir, name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		name = nextNumber(dir)
	}
	if strings.ContainsAny(name, `/\`) || name == "." || name == ".." {
		return fmt.Sprintf("%q can't be a folder name, so nothing was made", name), nil
	}
	path := filepath.Join(dir, name)
	if _, err := os.Stat(path); err == nil {
		return path + " is already there, so nothing was made", nil
	}
	if err := os.Mkdir(path, 0o755); err != nil {
		return "", err
	}
	if m.FoldersOpen == nil {
		m.FoldersOpen = map[string]bool{}
	}
	m.FoldersOpen[dir] = true
	if m.Root != "" && strings.HasPrefix(path, m.Root+string(filepath.Separator)) {
		return "", m.rereadSteps()
	}
	return "", nil
}

func pathRow(m Model) Group {
	dir := m.cwd()
	row := []Cell{Button("↑", Act{Op: OpFolderGo, Node: filepath.Dir(dir)}, false).Tipped("go up one folder")}
	parts := strings.Split(strings.Trim(dir, string(filepath.Separator)), string(filepath.Separator))
	at := string(filepath.Separator)
	row = append(row, Button(at, Act{Op: OpFolderGo, Node: at}, false))
	for i, p := range parts {
		if p == "" {
			continue
		}
		at = filepath.Join(at, p)
		row = append(row, Button(p, Act{Op: OpFolderGo, Node: at}, i == len(parts)-1))
	}
	return Group{Rows: [][]Cell{row}, Slides: true, Table: "folder path"}
}

func folderBrowser(m Model) []Group {
	return append([]Group{pathRow(m)}, folderGroups(m, m.cwd(), 0)...)
}

func folderGroups(m Model, dir string, depth int) []Group {
	pad := strings.Repeat("    ", depth)
	open := m.folderOpen(dir)
	title := []Cell{
		Text(pad + foldTitle(!open, filepath.Base(dir)+"/")).Acting(Act{Op: OpFolderFold, Node: dir}),
		Button("→", Act{Op: OpFolderGo, Node: dir}, false).Tipped("go into this folder"),
		Field("+ folder", "", Act{Op: OpFolderNew, Node: dir}, true, false).Fitting("a folder name").Outline(false).Wording().Tipped("type a name and press Enter, or press Enter for the next number"),
	}
	g := Group{Rows: [][]Cell{title}}
	entries, err := os.ReadDir(dir)
	if err != nil {
		g.Rows = append(g.Rows, []Cell{Text(pad + "    " + err.Error())})
		return []Group{g}
	}
	if !open {
		return []Group{g}
	}
	var files, dirs []string
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".") {
			continue
		}
		if e.IsDir() { // path-resolution-ok: listing a folder as it is on disk
			dirs = append(dirs, e.Name())
		} else {
			files = append(files, e.Name())
		}
	}
	out := []Group{g}
	if len(files) > 0 {
		out = append(out, Group{Rows: [][]Cell{{Text(pad + "    " + strings.Join(files, "  "))}}, Wraps: true})
	}
	for _, d := range dirs {
		out = append(out, folderGroups(m, filepath.Join(dir, d), depth+1)...)
	}
	return out
}

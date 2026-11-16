package Builder

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
	"github.com/dtauraso/beadnetwork/Categories/Scene/Scenes"
)

const (
	OpTreeGo     = "tree-go"
	OpTreeFold   = "tree-fold"
	FileTreeOpen = "tree-open.bin"
)

type StepRow struct {
	Path  []int
	Label string
	Depth int
	Kids  int
	Open  bool
	Here  bool
	Shown bool
}

func PathKey(p []int) string {
	parts := make([]string, len(p))
	for i, n := range p {
		parts[i] = strconv.Itoa(n)
	}
	return strings.Join(parts, "/")
}

func ParsePath(s string) []int {
	var p []int
	for _, part := range strings.Split(s, "/") {
		if n, err := strconv.Atoi(part); err == nil && n > 0 {
			p = append(p, n)
		}
	}
	return p
}

func LevelPath(level string) []int {
	var p []int
	parts := strings.Split(filepath.ToSlash(level), "/")
	for i := 1; i < len(parts); i += 2 {
		if n, err := strconv.Atoi(parts[i]); err == nil && parts[i-1] == DirSteps {
			p = append(p, n)
		}
	}
	return p
}

func LevelOf(p []int) string {
	var parts []string
	for _, n := range p {
		parts = append(parts, DirSteps, strconv.Itoa(n))
	}
	return filepath.Join(parts...)
}

func (m Model) pathDir(p []int) string {
	dir := m.Scene
	for _, n := range p {
		dir = stepDir(dir, n)
	}
	return dir
}

func (m Model) innerPath() []int {
	p := LevelPath(m.Level)
	if n := m.StepShown(); n > 0 {
		p = append(p, n)
	}
	return p
}

func (m Model) OpenPath() []int {
	if m.OpenTop > 0 {
		return append([]int{m.OpenTop}, m.innerPath()...)
	}
	return m.innerPath()
}

func (m *Model) ReadTree() {
	m.Tree = nil
	if m.Scene != "" && strings.TrimSpace(m.TreeFind) != "" {
		m.walkFind(nil, m.OpenPath())
	} else if m.Scene != "" {
		m.walkTree(nil, m.OpenPath())
	}
}

func (m *Model) walkTree(parent, here []int) {
	dir := m.pathDir(parent)
	for _, n := range ReadStepNumbers(dir) {
		p := append(slices.Clone(parent), n)
		r := StepRow{Path: p, Label: stepLabelAt(dir, n), Depth: len(parent), Shown: slices.Equal(p, here)}
		r.Here = len(p) <= len(here) && slices.Equal(p, here[:len(p)])
		r.Kids = len(ReadStepNumbers(stepDir(dir, n)))
		r.Open = r.Here && !r.Shown
		Links.ReadIfExists(filepath.Join(stepDir(dir, n), FileTreeOpen), &r.Open)
		r.Open = r.Open && r.Kids > 0
		m.Tree = append(m.Tree, r)
		if r.Open {
			m.walkTree(p, here)
		}
	}
}

func (m *Model) ToggleTreeFold(p []int) error {
	for _, r := range m.Tree {
		if slices.Equal(r.Path, p) {
			if err := Links.WriteAtomic(filepath.Join(m.pathDir(p), FileTreeOpen), !r.Open); err != nil {
				return err
			}
		}
	}
	m.ReadTree()
	return nil
}

func (m Model) Crosses(p []int) bool { return m.OpenTop > 0 && (len(p) == 0 || p[0] != m.OpenTop) }

func (m *Model) TreeGo(p []int) error {
	inner := p
	if m.OpenTop > 0 {
		inner = p[1:]
	}
	if len(inner) == 0 {
		inner = Scenes.OpenStepPath(m.Root)
	}
	if _, err := os.Stat(filepath.Join(m.Root, LevelOf(inner))); err != nil || len(inner) == 0 {
		inner = []int{1}
	}
	if err := m.WriteSteps(); err != nil {
		return err
	}
	m.Level, m.ShownStep = LevelOf(inner[:len(inner)-1]), inner[len(inner)-1]
	m.StepClosed = false
	err := m.rereadSteps()
	return err
}

func (m Model) SaveOpen() error {
	if m.Scene == "" || m.Root == "" {
		return nil
	}
	err := Scenes.WriteOpenStep(m.Scene, m.OpenPath())
	if m.Root != m.Scene {
		err = errors.Join(err, Scenes.WriteOpenStep(m.Root, m.innerPath()))
	}
	return err
}

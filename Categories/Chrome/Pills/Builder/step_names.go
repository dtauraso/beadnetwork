package Builder

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Board"
	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

const (
	OpStepName   = "step-name"
	FileStepName = "name.bin"
)

func (m Model) stepStrip(add Act) []Board.Chip {
	out := []Board.Chip{{Label: "START", Act: Act{Op: OpStartFold}, Lit: m.StartOpen, Tip: "open or close where each link begins"}}
	ns := m.stepNums()
	for _, n := range ns {
		c := Board.Chip{Label: m.StepLabel(n), Act: Act{Op: OpStepShow, I: n}, Lit: n == m.StepShown(), Tip: "show step " + itoa(n) + " of " + itoa(len(ns)) + " in the scene"}
		out = append(out, c)
	}
	if add.Op != "" {
		out = append(out, Board.Chip{Label: "＋", Act: add, Tip: "add a step at the end, open"})
	}
	return out
}

func (m *Model) readStepNames() {
	m.StepNames = map[int]string{}
	for _, n := range m.stepNums() {
		var name string
		if Links.ReadIfExists(filepath.Join(stepDir(m.lvl(), n), FileStepName), &name) && name != "" {
			m.StepNames[n] = name
		}
	}
}

func (m Model) StepLabel(n int) string {
	if name := m.StepNames[n]; name != "" {
		return strconv.Itoa(n) + " · " + name
	}
	return strconv.Itoa(n)
}

func (m *Model) NameStep(n int, name string) string {
	if name = strings.TrimSpace(name); name == "step "+strconv.Itoa(n) {
		name = ""
	}
	if m.Root == "" {
		return "steps are kept in the scene's steps/ folder, and this scene has none, so the name was not kept"
	}
	path := filepath.Join(stepDir(m.lvl(), n), FileStepName)
	var err error
	if name == "" {
		err = os.Remove(path)
		if os.IsNotExist(err) {
			err = nil
		}
	} else {
		err = Links.WriteAtomic(path, name)
	}
	if err != nil {
		return "step " + strconv.Itoa(n) + ": " + err.Error()
	}
	if m.StepNames == nil {
		m.StepNames = map[int]string{}
	}
	m.StepNames[n] = name
	m.readLevels()
	return ""
}

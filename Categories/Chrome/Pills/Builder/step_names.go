package Builder

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

const (
	OpStepName   = "step-name"
	FileStepName = "name.bin"
	stepNameFit  = "a step name here"
)

func (m *Model) readStepNames() {
	m.StepNames = map[int]string{}
	for _, n := range m.stepNums() {
		var name string
		if Links.ReadIfExists(filepath.Join(stepDir(m.Root, n), FileStepName), &name) && name != "" {
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
	name = strings.TrimSpace(name)
	if m.Root == "" {
		return "steps are kept in the scene's steps/ folder, and this scene has none, so the name was not kept"
	}
	path := filepath.Join(stepDir(m.Root, n), FileStepName)
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
	return ""
}

func stepNameCell(m Model, n int) Cell {
	return Field("name", m.StepNames[n], Act{Op: OpStepName, I: n}, true, false).Fitting(stepNameFit).Wording()
}

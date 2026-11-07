package Builder

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

const (
	OpStepPick   = "step-pick"
	DirStepPicks = "picks"
	FileExtPick  = ".bin"
)

var StepNumbers = []int{1, 2}

type StepEdge struct {
	Step     int
	From, To string
}

func (m Model) StepPicked(step int, l Links.Link) bool {
	return m.StepPicks[StepEdge{step, l.From, l.To}]
}

func (m *Model) ToggleStepPick(step int, from, to string) {
	if m.StepPicks == nil {
		m.StepPicks = map[StepEdge]bool{}
	}
	k := StepEdge{step, from, to}
	m.StepPicks[k] = !m.StepPicks[k]
}

func stepPicksDir(stepDir string, step int) string {
	return filepath.Join(stepDir, DirStepPicks, "step-"+strconv.Itoa(step))
}

func (m *Model) ReadStepPicks() {
	if m.StepDir == "" {
		return
	}
	m.StepPicks = map[StepEdge]bool{}
	for _, step := range StepNumbers {
		dir := stepPicksDir(m.StepDir, step)
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			from, to, ok := strings.Cut(strings.TrimSuffix(e.Name(), FileExtPick), "To")
			if e.IsDir() || !ok { // path-resolution-ok: a step pick is a <from>To<to>.bin leaf; skipping strays
				continue
			}
			var on bool
			if Links.ReadIfExists(filepath.Join(dir, e.Name()), &on) && on {
				m.StepPicks[StepEdge{step, from, to}] = true
			}
		}
	}
}

func (m Model) writeStepPicks() error {
	for k, on := range m.StepPicks {
		dir := stepPicksDir(m.StepDir, k.Step)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
		if err := Links.WriteAtomicIfChanged(filepath.Join(dir, k.From+"To"+k.To+FileExtPick), on); err != nil {
			return err
		}
	}
	return nil
}

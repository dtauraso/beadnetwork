package Builder

import (
	"strconv"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
)

const (
	OpRename  = "rename"
	OpLinkRow = "link-row"

	renameStep = "step"
)

func itoa(n int) string { return strconv.Itoa(n) }

func (m Model) Rename(a Act, e *Sheet.Edit) {
	if c, ok := m.RenameCell(a); ok {
		Sheet.StartDraft(e, c)
	}
}

func (m Model) RenameCell(a Act) (Cell, bool) {
	switch a.K {
	case renameStep:
		return Field("", m.StepNames[a.I], Act{Op: OpStepName, I: a.I}, true, false).Wording(), true
	case renamePath:
		return Field("", m.PathName(ParsePath(a.Node)), Act{Op: OpTreeName, Node: a.Node}, true, false).Wording(), true
	}
	return Cell{}, false
}

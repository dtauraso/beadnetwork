package viewstate

import (
	"github.com/dtauraso/wirefold/Categories/Chrome/Panels/PolarRulesPanel"
)

func (ui *UIState) ToggleSharedRow(row int32) {
	if ui.RuleSharedRow == row {
		ui.RuleSharedRow = -1
		return
	}
	ui.RuleSharedRow = row
}

func (ui *UIState) StartThetaDraft(row int32, self bool) {
	ui.RuleEdit = PolarRulesPanel.Edit{
		Active:  true,
		NodeRow: row,
		Self:    self,
		Draft:   "1/2",
	}
}

type ThetaCommit struct {
	Commit  bool
	NodeRow int32
	Self    bool
	Turns   float64
}

func (ui *UIState) RuleKey(key string) ThetaCommit {
	e := &ui.RuleEdit
	if !e.Active {
		return ThetaCommit{}
	}
	defer ui.EmitViewFrame(nil)

	switch key {
	case "Escape":
		*e = PolarRulesPanel.Edit{}
	case "Enter":
		row, self, draft := e.NodeRow, e.Self, e.Draft
		*e = PolarRulesPanel.Edit{}
		if turns, ok := PolarRulesPanel.ParsePiDraft(draft); ok {
			return ThetaCommit{Commit: true, NodeRow: row, Self: self, Turns: turns}
		}
	case "Backspace":
		if len(e.Draft) > 0 {
			e.Draft = e.Draft[:len(e.Draft)-1]
		}
	default:
		if len(key) == 1 {
			e.Draft += key
		}
	}
	return ThetaCommit{}
}

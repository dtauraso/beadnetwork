package PolarRulesPanel

type ThetaCommit struct {
	Commit  bool
	NodeRow int32
	Self    bool
	Turns   float64
}

func ToggleSharedRow(sharedRow *int32, row int32) {
	if *sharedRow == row {
		*sharedRow = -1
		return
	}
	*sharedRow = row
}

func StartThetaDraft(e *Edit, row int32, self bool) {
	*e = Edit{
		Active:  true,
		NodeRow: row,
		Self:    self,
		Draft:   "1/2",
	}
}

func RuleKey(e *Edit, key string, redraw func()) ThetaCommit {
	if !e.Active {
		return ThetaCommit{}
	}
	defer redraw()

	switch key {
	case "Escape":
		*e = Edit{}
	case "Enter":
		row, self, draft := e.NodeRow, e.Self, e.Draft
		*e = Edit{}
		if turns, ok := ParsePiDraft(draft); ok {
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

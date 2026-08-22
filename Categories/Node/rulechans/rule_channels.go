package rulechans

import "github.com/dtauraso/wirefold/Categories/Node/rulenode"

type RuleChannels struct {
	EditsByNodeRow []chan<- rulenode.Edit

	KindTogglesByNodeRow []chan<- struct{}

	TogglesByEdgeRow []chan<- struct{}
}

func (rc *RuleChannels) SizeByNodeRows(rows int) {
	rc.EditsByNodeRow = make([]chan<- rulenode.Edit, rows)
	rc.KindTogglesByNodeRow = make([]chan<- struct{}, rows)
}

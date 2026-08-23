package rulenode

type RuleChannels struct {
	EditsByNodeRow []chan<- Edit

	KindTogglesByNodeRow []chan<- struct{}

	TogglesByEdgeRow []chan<- struct{}
}

func (rc *RuleChannels) SizeByNodeRows(rows int) {
	rc.EditsByNodeRow = make([]chan<- Edit, rows)
	rc.KindTogglesByNodeRow = make([]chan<- struct{}, rows)
}

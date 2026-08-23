package Node

type RuleChannels struct {
	EditsByNodeRow []chan<- RuleEdit

	KindTogglesByNodeRow []chan<- struct{}

	TogglesByEdgeRow []chan<- struct{}
}

func (rc *RuleChannels) SizeByNodeRows(rows int) {
	rc.EditsByNodeRow = make([]chan<- RuleEdit, rows)
	rc.KindTogglesByNodeRow = make([]chan<- struct{}, rows)
}

package loadspec

import (
	"strconv"

	"github.com/dtauraso/wirefold/Chrome/Panels/TiltPanel"
)

func (spec TopoSpec) TiltPanelRows() (rows []int32, labels []string) {
	byRow := make([]string, spec.RowCount)
	for _, n := range spec.Nodes {
		id, err := strconv.Atoi(n.ID)
		if err != nil || id-1 < 0 || id-1 >= spec.RowCount {
			continue
		}
		if !TiltPanel.KindWantsVectorChannel(n.Type) {
			continue
		}
		byRow[id-1] = n.ID
	}
	for row, label := range byRow {
		if label == "" {
			continue
		}
		rows = append(rows, int32(row))
		labels = append(labels, label)
	}
	return rows, labels
}

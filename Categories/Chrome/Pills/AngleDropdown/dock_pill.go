package AngleDropdown

import (
	"strconv"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
)

const OpLattice = "lattice-points"

func DockPill(latticePoints int32) Sheet.Pill {
	field := Sheet.Field(LatticeName, strconv.Itoa(int(latticePoints)), Sheet.Act{Op: OpLattice}, true, true)
	return Sheet.Pill{Label: Label, Bottom: true, Groups: []Sheet.Group{
		{Title: "angles", Rows: [][]Sheet.Cell{
			{field, Sheet.Text("steps of " + strconv.Itoa(LatticePointsStep) + ", from " + strconv.Itoa(LatticePointsMin) + " to " + strconv.Itoa(LatticePointsMax))},
		}},
	}}
}

package Node

type Seed struct {
	ID, Label, Kind    string
	CX, CY, CZ, Radius float64

	Row int
}

func NewSeed(id string, g NodeGeom, row int) Seed {
	label := g.Label
	if label == "" {
		label = id
	}
	var cx, cy, cz float64
	if g.HasPos {
		c := NodeWorldPos(g)
		cx, cy, cz = c.X, c.Y, c.Z
	}
	return Seed{
		ID: id, Label: label, Kind: g.Kind,
		CX: cx, CY: cy, CZ: cz,
		Radius: NodeRadius(g.Kind),
		Row:    row,
	}
}

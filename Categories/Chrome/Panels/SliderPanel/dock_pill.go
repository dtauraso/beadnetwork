package SliderPanel

import "github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"

const (
	Label   = "Speed"
	OpSpeed = "speed"
)

func (s Setting) Shown() string {
	if s.Den == "" {
		return s.Num
	}
	return s.Num + "/" + s.Den
}

func DockPill(speed float64) Sheet.Pill {
	row := []Sheet.Cell{Sheet.Text("speed")}
	for i, s := range Settings {
		row = append(row, Sheet.Button(s.Shown(), Sheet.Act{Op: OpSpeed, I: i}, s.Speed == speed))
	}
	return Sheet.Pill{Label: Label, Bottom: true, Groups: []Sheet.Group{
		{Title: "speed — how fast the network runs", Rows: [][]Sheet.Cell{row}},
	}}
}

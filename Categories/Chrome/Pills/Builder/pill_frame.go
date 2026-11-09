package Builder

import "strconv"

const FrameLabel = "Frame"

const (
	OpFrame  = "frame"
	OpTickMs = "tick-ms"

	FrameRho    = "rho"
	FrameTphi   = "tphi"
	FrameTtheta = "ttheta"
	FrameCr     = "cr"
	FrameS      = "s"
)

func FramePill(m Model) Pill {
	f := m.Frame
	act := func(k string) Act { return Act{Op: OpFrame, K: k} }
	return Pill{Label: FrameLabel, Bottom: true, Groups: []Group{
		{Title: "the frame every node is placed in — a change reloads the scene", Rows: [][]Cell{
			{Field("ρ", Num(f.Rho), act(FrameRho), true, false), Text("length of one unit of a link's r")},
			{Field("T_φ", strconv.Itoa(f.Tphi), act(FrameTphi), true, false), Text("index steps a turn in φ")},
			{Field("T_θ", strconv.Itoa(f.Ttheta), act(FrameTtheta), true, false), Text("index steps a turn in θ")},
			{Field("c_r", Num(f.Cr), act(FrameCr), true, false), Text("length of one radial index step")},
			{Field("s", strconv.Itoa(m.S), act(FrameS), true, false), Text("the largest node s — typing it sets every node's s")},
		}},
		{Title: "time — each node loops at machine speed; its machine ticks become one human tick", Rows: [][]Cell{
			{Field("Δt", strconv.Itoa(m.TickMs), Act{Op: OpTickMs}, true, true), Text("ms one human tick lasts")},
			{Text("behind " + strconv.Itoa(m.Behind()) + "/" + strconv.Itoa(len(m.Nodes))), Text("nodes that have not reached h−1")},
		}},
	}}
}

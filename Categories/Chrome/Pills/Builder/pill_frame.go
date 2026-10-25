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
	return Pill{Label: FrameLabel, Groups: []Group{
		{Title: "the frame every node is placed in — a change reloads the scene", Rows: [][]Cell{
			{Field("ρ", Num(f.Rho), act(FrameRho), true, false), Text("length of one unit of a link's r")},
			{Field("T_φ", strconv.Itoa(f.Tphi), act(FrameTphi), true, false), Text("index steps a turn in φ")},
			{Field("T_θ", strconv.Itoa(f.Ttheta), act(FrameTtheta), true, false), Text("index steps a turn in θ")},
			{Field("c_r", Num(f.Cr), act(FrameCr), true, false), Text("length of one radial index step")},
			{Field("s", strconv.Itoa(m.S), act(FrameS), true, false), Text("ticks a twelfth of a turn, for a node with no s of its own")},
		}},
		{Title: "time — each node loops at machine speed; its machine ticks become one human tick", Rows: [][]Cell{
			{Field("Δt", strconv.Itoa(m.TickMs), Act{Op: OpTickMs}, true, true), Text("ms one human tick lasts")},
			{Text("h " + strconv.FormatInt(m.Now, 10)), Text("human ticks since the scene started")},
			{Text("n min " + m.NMin()), Text("machine ticks per human tick of the slowest node — near 0, it is about to miss human ticks")},
			{Text("behind " + strconv.Itoa(m.Behind()) + "/" + strconv.Itoa(len(m.Nodes))), Text("nodes that have not reached h−1")},
		}},
	}}
}

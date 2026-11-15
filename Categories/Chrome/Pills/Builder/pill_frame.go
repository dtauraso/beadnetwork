package Builder

import "strconv"

const (
	OpFrame  = "frame"
	OpTickMs = "tick-ms"

	OpFrameFold = "frame-fold"

	FrameRho    = "rho"
	FrameTphi   = "tphi"
	FrameTtheta = "ttheta"
	FrameCr     = "cr"
	FrameS      = "s"
)

func frameAct(k string) Act { return Act{Op: OpFrame, K: k} }

func allSeqsGroup(m Model) Group {
	const title = "all sequences — the frame every node is placed in (a change reloads the scene), and time"
	fold := Act{Op: OpFrameFold}
	if m.FrameFolded {
		return Group{Title: "›  " + title, Fold: fold}
	}
	f := m.Frame
	return Group{Title: "⌄  " + title, Fold: fold, Rows: [][]Cell{
		{Field("ρ", Num(f.Rho), frameAct(FrameRho), true, false), Text("length of one unit of a link's r")},
		{Field("T_φ", strconv.Itoa(f.Tphi), frameAct(FrameTphi), true, false), Text("index steps a turn in φ")},
		{Field("T_θ", strconv.Itoa(f.Ttheta), frameAct(FrameTtheta), true, false), Text("index steps a turn in θ")},
		{Field("c_r", Num(f.Cr), frameAct(FrameCr), true, false), Text("length of one radial index step")},
		{Field("Δt", strconv.Itoa(m.TickMs), Act{Op: OpTickMs}, true, true), Text("ms one human tick lasts — each node's machine ticks become one human tick")},
	}}
}

package selectright

func clearWindow(g *GateNode, w *gateWindow) {
	g.FromLeft.Breadcrumb("window_clear", "")
	g.HasLeft = false
	g.HasRight = false
	w.t0Set = false
}

func openWindowIfNeeded(g *GateNode, w *gateWindow, now func() int64) {
	if (g.HasLeft || g.HasRight) && !w.t0Set {
		w.t0 = now()
		w.t0Set = true

		g.FromLeft.Breadcrumb("window_open", "")
	}
}

func tryFireOnDwell(g *GateNode, w *gateWindow, now func() int64, fireResult func(*GateNode) int) bool {
	if !(g.HasLeft && g.HasRight) {
		return false
	}

	if !w.dwellSet {
		w.dwellStart = now()
		w.dwellSet = true

		g.FromLeft.Breadcrumb("dwell_start", "")
	}
	if now()-w.dwellStart < fireDwellTicks {
		return false
	}
	result := fireResult(g)
	if g.Fire != nil {
		g.Fire()
	}
	g.HasLeft = false
	g.HasRight = false
	w.t0Set = false
	w.dwellSet = false
	emitInputs(g)

	g.ToPassed.PlaceDrivenAt(result, now())
	return true
}

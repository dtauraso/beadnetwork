package gatecommon

// clearWindow discards both held inputs without firing: resets the has-input flags
// and the window-open state. Breadcrumb on FromLeft (the consistent logging point).
func clearWindow(g *GateNode, w *gateWindow) {
	g.FromLeft.Breadcrumb("window_clear", "")
	g.HasLeft = false
	g.HasRight = false
	w.t0Set = false
}

// openWindowIfNeeded opens the coincidence window on the first input to arrive.
func openWindowIfNeeded(g *GateNode, w *gateWindow, now func() int64) {
	if (g.HasLeft || g.HasRight) && !w.t0Set {
		w.t0 = now()
		w.t0Set = true
		// Breadcrumb the window-open instant. t0 is now captured against the
		// clock, so an observer that waits for this before advancing the sim
		// clock can't race the t0 = now() read (deterministic test sync).
		g.FromLeft.Breadcrumb("window_open", "")
	}
}

// tryFireOnDwell handles the both-inputs-held case: it starts the fire-dwell timer
// on first entry, and once the dwell has elapsed, fires the fireResult and resets
// the held/window/dwell state. Returns true if it fired (caller should `continue`
// its loop iteration without also running the window-timeout check).
func tryFireOnDwell(g *GateNode, w *gateWindow, now func() int64, fireResult func(*GateNode) int) bool {
	if !(g.HasLeft && g.HasRight) {
		return false
	}
	// Both inputs held: dwell so both interior beads are visible before the gate
	// resolves. Once committed to the dwell, the window-timeout is gated off so it
	// can't clip the fire.
	if !w.dwellSet {
		w.dwellStart = now()
		w.dwellSet = true
		// Breadcrumb the dwell-start instant. dwellStart is now captured against
		// the clock, so an observer can wait for this before advancing the sim
		// clock without racing the dwellStart = now() read.
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
	// Place the fire result without walking it to delivery. The wire's driver
	// (its source node's mover) times its traversal — the gate goroutine is
	// never parked across the output traversal. now() is this goroutine's own
	// clock (injected by the caller, gate.go's Update loop) — read once here
	// and stamped as this bead's placementTick.
	g.ToPassed.PlaceDrivenAt(result, now())
	return true
}

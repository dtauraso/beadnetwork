package gatecommon

import wire "github.com/dtauraso/wirefold/nodes/wire"

// drainLatestReal consumes ALL queued beads on a side and returns the most-recent
// REAL value (discarding NoValue placeholders). got=false when nothing real was queued.
//
// Drain-until-empty, transitively bounded by this In's own wire's declared channel
// capacity — no iteration cap; see nodes/wire/paced_wire_drive.go's drainPlacements doc
// comment for the full reasoning shared by every drain-until-empty loop in this repo.
func drainLatestReal(in *wire.In) (int, bool) {
	v, got := NoValue, false
	for {
		nv, ok := in.PollRecv()
		if !ok {
			break
		}
		if nv != NoValue {
			v = nv
			got = true
		}
	}
	return v, got
}

// captureLeft drains FromLeft and, if a real value arrived, applies the invertLeft
// inversion and stores it (only if it changed the held value/presence). Returns
// true when the held state changed and inputs should be re-emitted.
func captureLeft(g *GateNode, invertLeft bool) bool {
	v, got := drainLatestReal(g.FromLeft)
	if !got {
		return false
	}
	var stored int
	if invertLeft {
		stored = 1 - v // NOT the left input
	} else {
		stored = v
	}
	if !g.HasLeft || g.Left != stored {
		g.Left = stored
		g.HasLeft = true
		return true
	}
	return false
}

// captureRight drains FromRight and, if a real value arrived, applies the
// complementary (NOT invertLeft) inversion and stores it. Returns true when the
// held state changed and inputs should be re-emitted.
func captureRight(g *GateNode, invertLeft bool) bool {
	v, got := drainLatestReal(g.FromRight)
	if !got {
		return false
	}
	var stored int
	if !invertLeft {
		stored = 1 - v // NOT the right input
	} else {
		stored = v
	}
	if !g.HasRight || g.Right != stored {
		g.Right = stored
		g.HasRight = true
		return true
	}
	return false
}

// captureRawLeft drains FromLeft and stores the raw value with NO inversion.
// Returns true when the held state changed and inputs should be re-emitted.
// Used by RunGateAccept (direct pattern-match acceptance, no NOT gates).
func captureRawLeft(g *GateNode) bool {
	v, got := drainLatestReal(g.FromLeft)
	if !got {
		return false
	}
	if !g.HasLeft || g.Left != v {
		g.Left = v
		g.HasLeft = true
		return true
	}
	return false
}

// captureRawRight drains FromRight and stores the raw value with NO inversion.
// Returns true when the held state changed and inputs should be re-emitted.
// Used by RunGateAccept (direct pattern-match acceptance, no NOT gates).
func captureRawRight(g *GateNode) bool {
	v, got := drainLatestReal(g.FromRight)
	if !got {
		return false
	}
	if !g.HasRight || g.Right != v {
		g.Right = v
		g.HasRight = true
		return true
	}
	return false
}

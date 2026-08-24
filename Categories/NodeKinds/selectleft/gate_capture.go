package selectleft

import beadanimation "github.com/dtauraso/beadnetwork/Categories/Node/BeadAnimation"

func drainLatestReal(in *beadanimation.Receiver) (int, bool) {
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

func captureLeft(g *GateNode, invertLeft bool) bool {
	v, got := drainLatestReal(g.FromLeft)
	if !got {
		return false
	}
	var stored int
	if invertLeft {
		stored = 1 - v
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

func captureRight(g *GateNode, invertLeft bool) bool {
	v, got := drainLatestReal(g.FromRight)
	if !got {
		return false
	}
	var stored int
	if !invertLeft {
		stored = 1 - v
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

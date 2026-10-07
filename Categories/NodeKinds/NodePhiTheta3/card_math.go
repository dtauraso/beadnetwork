package NodePhiTheta3

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/CardPanel"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/TiltPanel"
)

type Vec = CardPanel.Vec

const (
	poleLow  = 0
	poleMid  = 6
	poleHigh = 12
	qtLow    = 3
	qtHigh   = 9
	jump     = 1
)

func msgOf(v Vec) TiltPanel.TiltVectorMsg {
	return TiltPanel.TiltVectorMsg{PhiIdx: int32(v.Phi), ThetaIdx: int32(v.Theta), RIdx: int32(v.R)}
}

func vecOf(m TiltPanel.TiltVectorMsg) Vec {
	return Vec{Phi: int(m.PhiIdx), Theta: int(m.ThetaIdx), R: int(m.RIdx)}
}

func scale(k int, v Vec) Vec { return Vec{Phi: k * v.Phi, Theta: k * v.Theta, R: k * v.R} }

func add(a, b Vec) Vec { return Vec{Phi: a.Phi + b.Phi, Theta: a.Theta + b.Theta, R: a.R + b.R} }

func pickOne(k1, k2 int, a1, a2 Vec) Vec {
	if k1^k2 != 1 {
		return Vec{}
	}
	return add(scale(k1, a1), scale(k2, a2))
}

func down(arrival, pole, qt int) int {
	if pole < arrival && arrival < qt {
		return -jump
	}
	return 0
}

func up(arrival, qt, pole int) int {
	if qt < arrival && arrival < pole {
		return jump
	}
	return 0
}

func dirDown(a Vec, p, qt int, offset Vec) Vec {
	return Vec{Phi: down(a.Phi, p+offset.Phi, qt), Theta: down(a.Theta, p+offset.Theta, qt)}
}

func dirUp(a Vec, qt, p int, offset Vec) Vec {
	return Vec{Phi: up(a.Phi, qt, p-offset.Phi), Theta: up(a.Theta, qt, p-offset.Theta)}
}

func step(a, offset Vec) Vec {
	sum := add(add(dirDown(a, poleLow, qtLow, offset), dirUp(a, qtLow, poleMid, offset)),
		add(dirDown(a, poleMid, qtHigh, offset), dirUp(a, qtHigh, poleHigh, offset)))
	sum.R = a.R
	return sum
}

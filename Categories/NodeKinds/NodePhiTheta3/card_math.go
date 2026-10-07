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

func down(arrival, pole, qt, m int) int {
	if pole < arrival && arrival < qt {
		return -m
	}
	return 0
}

func up(arrival, qt, pole, m int) int {
	if qt < arrival && arrival < pole {
		return m
	}
	return 0
}

func dirDown(a Vec, p, qt int, offset Vec, m int) Vec {
	return Vec{Phi: down(a.Phi, p+offset.Phi, qt, m), Theta: down(a.Theta, p+offset.Theta, qt, m)}
}

func dirUp(a Vec, qt, p int, offset Vec, m int) Vec {
	return Vec{Phi: up(a.Phi, qt, p-offset.Phi, m), Theta: up(a.Theta, qt, p-offset.Theta, m)}
}

func step(a, offset Vec, m int) Vec {
	sum := add(add(dirDown(a, poleLow, qtLow, offset, m), dirUp(a, qtLow, poleMid, offset, m)),
		add(dirDown(a, poleMid, qtHigh, offset, m), dirUp(a, qtHigh, poleHigh, offset, m)))
	sum.R = a.R
	return sum
}

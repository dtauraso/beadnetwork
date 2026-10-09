package NodePhiTheta3

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/CardPanel"
)

type Vec = CardPanel.Vec

type Angles struct{ Phi, Theta int }

const (
	poleLow  = 0
	poleMid  = 6
	poleHigh = 12
	qtLow    = 3
	qtHigh   = 9
)

func anglesOf(v Vec) Angles { return Angles{Phi: v.Phi, Theta: v.Theta} }

func scale(k int, a Angles) Angles { return Angles{Phi: k * a.Phi, Theta: k * a.Theta} }

func add(a, b Angles) Angles { return Angles{Phi: a.Phi + b.Phi, Theta: a.Theta + b.Theta} }

func sub(a, b Angles) Angles { return Angles{Phi: a.Phi - b.Phi, Theta: a.Theta - b.Theta} }

func pickOne(k1, k2 int, a1, a2 Angles) Angles {
	if k1^k2 != 1 {
		return Angles{}
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

func dirDown(a Angles, p, qt int, offset Angles, m int) Angles {
	return Angles{Phi: down(a.Phi, p+offset.Phi, qt, m), Theta: down(a.Theta, p+offset.Theta, qt, m)}
}

func dirUp(a Angles, qt, p int, offset Angles, m int) Angles {
	return Angles{Phi: up(a.Phi, qt, p-offset.Phi, m), Theta: up(a.Theta, qt, p-offset.Theta, m)}
}

func step(a Angles, offset Vec, m, s int) Angles {
	off := scale(s, anglesOf(offset))
	return add(a, add(add(dirDown(a, poleLow*s, qtLow*s, off, m), dirUp(a, qtLow*s, poleMid*s, off, m)),
		add(dirDown(a, poleMid*s, qtHigh*s, off, m), dirUp(a, qtHigh*s, poleHigh*s, off, m))))
}

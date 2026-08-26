package NodePhiTheta

import "github.com/dtauraso/beadnetwork/Categories/Vectors/polarindex"

type Turn = polarindex.Index

func mod(x, tau int) int {
	if tau <= 0 {
		return x
	}
	return (x%tau + tau) % tau
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func Bottom(top, tau Turn) Turn {
	return Turn{
		Phi:   mod(top.Phi+tau.Phi/2, tau.Phi),
		Theta: mod(top.Theta+tau.Theta/2, tau.Theta),
		R:     top.R,
	}
}

func Distance(end, arrival Turn) Turn {
	return Turn{
		Phi:   abs(end.Phi - arrival.Phi),
		Theta: abs(end.Theta - arrival.Theta),
		R:     abs(end.R - arrival.R),
	}
}

func offsetOn(distanceTop, distanceBottom, tau int) int {
	switch {
	case distanceTop == 0 && distanceBottom == 0:
		return 0
	case distanceTop < tau/4:
		return -1
	case distanceBottom < tau/4:
		return +1
	default:
		return 0
	}
}

func Offset(top, arrival, tau Turn) Turn {
	distanceTop := Distance(top, arrival)
	distanceBottom := Distance(Bottom(top, tau), arrival)
	return Turn{
		Phi:   offsetOn(distanceTop.Phi, distanceBottom.Phi, tau.Phi),
		Theta: offsetOn(distanceTop.Theta, distanceBottom.Theta, tau.Theta),
		R:     0,
	}
}

func Add(center, offset, tau Turn) Turn {
	return Turn{
		Phi:   mod(center.Phi+offset.Phi, tau.Phi),
		Theta: mod(center.Theta+offset.Theta, tau.Theta),
		R:     center.R,
	}
}

func CenterNext(center, top, arrival, tau Turn) Turn {
	return Add(center, Offset(top, arrival, tau), tau)
}

func AtRest(top, arrival, tau Turn) bool {
	o := Offset(top, arrival, tau)
	return o.Phi == 0 && o.Theta == 0
}

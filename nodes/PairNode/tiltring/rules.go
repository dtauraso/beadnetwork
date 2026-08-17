package tiltring

func Mod(x, tau int32) int32 { return (x%tau + tau) % tau }

func Abs(x int32) int32 {
	if x < 0 {
		return -x
	}
	return x
}

func Bottom(top, tau int32) int32 { return Mod(top+tau/2, tau) }

func DistanceTop(top, arrival int32) int32 { return Abs(top - arrival) }

func DistanceBottom(top, arrival, tau int32) int32 { return Abs(Bottom(top, tau) - arrival) }

func Offset(top, arrival, tau int32) int32 {
	distanceTop := DistanceTop(top, arrival)
	distanceBottom := DistanceBottom(top, arrival, tau)
	switch {
	case distanceTop == 0 || distanceBottom == 0:
		return 0
	case distanceTop < tau/4:
		return +1
	case distanceBottom < tau/4:
		return -1
	default:
		return 0
	}
}

func TopNext(top, arrival, tau int32) int32 { return Mod(top+Offset(top, arrival, tau), tau) }

func BottomNext(top, arrival, tau int32) int32 {
	return Mod(Bottom(top, tau)+Offset(top, arrival, tau), tau)
}

func Sent(top, tau int32) int32 { return Mod(top+tau/4, tau) }

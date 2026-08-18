package lattice

func BeadFraction(nowTick, placementTick, crossTicks float64) float64 {
	if crossTicks <= 0 {
		return 0
	}
	target := nowTick
	if nowTick >= placementTick+crossTicks {
		target = placementTick + crossTicks
	}
	t := (target - placementTick) / crossTicks
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	return t
}

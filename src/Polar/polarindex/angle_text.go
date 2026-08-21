package polarindex

import "fmt"

func AngleDenom(points int32) int32 {
	if points/2 < 1 {
		return 1
	}
	return points / 2
}

func AngleText(idx int32, points int32) string {
	if idx == 0 {
		return "0"
	}
	sign := ""
	if idx < 0 {
		sign = "-"
		idx = -idx
	}
	return fmt.Sprintf("%s%dπ/%d", sign, idx, AngleDenom(points))
}

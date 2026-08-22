package Camera

import "math"

const FocalPixels = 965.0

const (
	ReferenceHeightPx = 900.0
	ReferenceFovDeg   = 50.0
)

func FovDegForHeight(heightPx float64) float64 {
	if !(heightPx > 0) {
		return ReferenceFovDeg
	}
	return 2 * math.Atan(heightPx/(2*FocalPixels)) * 180 / math.Pi
}

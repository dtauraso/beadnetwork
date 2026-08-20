package interior

import "github.com/dtauraso/wirefold/src/Node/spatial"

type vec3 = spatial.Vec3

const (
	interiorBeadR         = 5.0
	interiorTorusTubeFrac = 0.12
	interiorBeadGap       = 0.2
)

const InteriorTorusOuterR = interiorBeadR * (1 + interiorTorusTubeFrac)

const InteriorSlot = InteriorTorusOuterR + interiorBeadGap/2

func InteriorSlotOffset(row, col int) vec3 {
	slot := InteriorSlot
	pitch := 2 * slot
	return vec3{
		X: (float64(col) - 0.5) * pitch,
		Y: (0.5 - float64(row)) * pitch,
		Z: 0,
	}
}

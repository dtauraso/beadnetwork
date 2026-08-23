package TiltVectors

import (
	"fmt"
)

var TiltValueNames = buildTiltValueNames()

func buildTiltValueNames() []string {
	names := []string{"received"}
	for m := range 16 {
		names = append(names, fmt.Sprintf("tiltShaftM%d", m))
	}
	for m := range 16 {
		names = append(names, fmt.Sprintf("tiltHeadM%d", m))
	}
	return names
}

func ShaftName(m int) string { return TiltValueNames[1+m] }
func HeadName(m int) string  { return TiltValueNames[17+m] }

package ChannelVectors

import (
	"fmt"
)

var VectorValueNames = buildVectorValueNames()

func buildVectorValueNames() []string {
	names := make([]string, 0, 32)
	for m := 0; m < 16; m++ {
		names = append(names, fmt.Sprintf("channelShaftM%d", m))
	}
	for m := 0; m < 16; m++ {
		names = append(names, fmt.Sprintf("channelHeadM%d", m))
	}
	return names
}

func ShaftName(m int) string { return VectorValueNames[m] }
func HeadName(m int) string  { return VectorValueNames[16+m] }

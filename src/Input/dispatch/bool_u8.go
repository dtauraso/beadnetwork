package dispatch

func boolU8(b bool) uint8 {
	if b {
		return 1
	}
	return 0
}

package Drag

func enumAt(list []string, i byte) string {
	if int(i) >= len(list) {
		return ""
	}
	return list[i]
}

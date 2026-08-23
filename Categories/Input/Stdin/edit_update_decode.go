package Stdin

func decodeEditUpdate(rec []byte) (StdinMsg, bool) {
	if len(rec) < 3 {
		return StdinMsg{}, false
	}
	entity := enumAt(InUpdateKinds, rec[1])
	decode, ok := updateDecoders[entity]
	if !ok {
		return StdinMsg{}, false
	}
	return decode(rec[3:], rec[2])
}

package Panel

var UpdateAttrs = []string{"toggle"}

func attrIndex(attr string) byte {
	for i, a := range UpdateAttrs {
		if a == attr {
			return byte(i)
		}
	}
	panic("Panel.attrIndex: no wire byte exists for update attribute " + attr +
		"; Panel.UpdateAttrs does not carry it, so an edit naming it could never be " +
		"encoded. Add it there and regenerate, in the same commit as the decoder that reads it.")
}

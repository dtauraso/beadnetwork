package Codec

func AttrIndex(attr string) byte {
	for i, a := range InUpdateAttrs {
		if a == attr {
			return byte(i)
		}
	}
	panic("Codec.AttrIndex: " + attr + " is not in the updateAttrs list of INPUT_LAYOUT_FINGERPRINT, " +
		"so nothing on the wire can carry it; add it to the fingerprint in the same commit as the decoder")
}

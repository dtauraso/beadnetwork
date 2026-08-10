// bool_u8.go — boolU8 converts a bool to the buffer's canonical 0/1 byte encoding. A
// SEPARATE unexported copy of the same one-liner lives in package interior
// (interior_stream.go's own doc comment) and in Buffer (unexported there too) — each
// package keeps its own trivial copy rather than importing another package for one
// function, the same precedent Buffer.boolU8's original comment already set.
package Wiring

func boolU8(b bool) uint8 {
	if b {
		return 1
	}
	return 0
}

package bufschema

// bufLayoutEdgeBead is one in-flight bead on an edge, at its WORLD position.
//
// The chain-bead block that preceded it was per NODE, and held an offset from
// that node's centre, because the node worked out where its neighbour was and
// laid a chain toward it. An edge already knows both of its endpoints, so a
// bead on it is placed along its own segment and needs no origin to be read
// against.
type bufLayoutEdgeBead struct {
	X float32 `buf:"f32"`
	Y float32 `buf:"f32"`
	Z float32 `buf:"f32"`

	Value int32 `buf:"i32"`
}

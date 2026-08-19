package Edge

var _ = bufLayoutEdge{}

type bufLayoutEdge struct {
	SX float32 `buf:"f32"`
	SY float32 `buf:"f32"`
	SZ float32 `buf:"f32"`
	EX float32 `buf:"f32"`
	EY float32 `buf:"f32"`
	EZ float32 `buf:"f32"`

	SrcNodeRow int32   `buf:"i32"`
	DstNodeRow int32   `buf:"i32"`
	DeltaR     float32 `buf:"f32"`

	DragActive uint8 `buf:"u8"`

	Label []byte `buf:"bytes"`
}

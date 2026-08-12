package Buffer

type bufLayoutEdge struct {
	SX float32 `buf:"f32"`
	SY float32 `buf:"f32"`
	SZ float32 `buf:"f32"`
	EX float32 `buf:"f32"`
	EY float32 `buf:"f32"`
	EZ float32 `buf:"f32"`

	EdgeLabelOff uint32 `buf:"u32"`
	EdgeLabelLen uint32 `buf:"u32"`
}

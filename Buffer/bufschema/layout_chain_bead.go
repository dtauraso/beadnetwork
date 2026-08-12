package bufschema

type bufLayoutChainBead struct {
	OX float32 `buf:"f32"`
	OY float32 `buf:"f32"`
	OZ float32 `buf:"f32"`

	Lit uint8 `buf:"u8"`

	LitValue int32 `buf:"i32"`
}

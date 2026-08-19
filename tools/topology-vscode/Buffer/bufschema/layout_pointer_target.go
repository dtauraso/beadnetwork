package bufschema

type bufLayoutPointerTarget struct {
	X float32 `buf:"f32"`
	Y float32 `buf:"f32"`
	W float32 `buf:"f32"`
	H float32 `buf:"f32"`

	Kind uint8 `buf:"u8"`

	TipX    float32 `buf:"f32"`
	TipY    float32 `buf:"f32"`
	TipText []byte  `buf:"bytes"`
}

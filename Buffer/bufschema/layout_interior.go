package bufschema

type bufLayoutInterior struct {
	Present uint8   `buf:"u8"`
	Value   int32   `buf:"i32"`
	OX      float32 `buf:"f32"`
	OY      float32 `buf:"f32"`
	OZ      float32 `buf:"f32"`
}

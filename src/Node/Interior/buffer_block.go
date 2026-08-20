package interior

var _ = bufLayoutInterior{}

type bufLayoutInterior struct {
	Present uint8   `buf:"u8"`
	Value   int32   `buf:"i32"`
	X       float32 `buf:"f32"`
	Y       float32 `buf:"f32"`
	Z       float32 `buf:"f32"`
}

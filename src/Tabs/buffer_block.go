package Tabs

var _ = bufLayoutTabStrip{}

type bufLayoutTabStrip struct {
	StripX float32 `buf:"f32"`
	StripY float32 `buf:"f32"`
	StripW float32 `buf:"f32"`
	StripH float32 `buf:"f32"`

	TabX float32 `buf:"f32"`
	TabY float32 `buf:"f32"`
	TabW float32 `buf:"f32"`
	TabH float32 `buf:"f32"`

	TabNameText []byte `buf:"bytes"`
	TabNameLen  uint32 `buf:"u32"`

	TabSelected uint8 `buf:"u8"`
}

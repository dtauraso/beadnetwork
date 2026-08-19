package TiltPanel

var _ = bufLayoutTiltPanel{}

type bufLayoutTiltPanel struct {
	BoxX float32 `buf:"f32"`
	BoxY float32 `buf:"f32"`
	BoxW float32 `buf:"f32"`
	BoxH float32 `buf:"f32"`

	StartX float32 `buf:"f32"`
	StartY float32 `buf:"f32"`
	StartW float32 `buf:"f32"`
	StartH float32 `buf:"f32"`

	ResetX float32 `buf:"f32"`
	ResetY float32 `buf:"f32"`
	ResetW float32 `buf:"f32"`
	ResetH float32 `buf:"f32"`

	StartText []byte `buf:"bytes"`
	ResetText []byte `buf:"bytes"`

	ColNodeRow int32 `buf:"i32"`

	ColLabelText []byte `buf:"bytes"`
	ColLabelLen  uint32 `buf:"u32"`

	HeadX float32 `buf:"f32"`
	HeadY float32 `buf:"f32"`
	HeadW float32 `buf:"f32"`
	HeadH float32 `buf:"f32"`

	RoundsX float32 `buf:"f32"`
	RoundsY float32 `buf:"f32"`
	RoundsW float32 `buf:"f32"`
	RoundsH float32 `buf:"f32"`

	MsgsX float32 `buf:"f32"`
	MsgsY float32 `buf:"f32"`
	MsgsW float32 `buf:"f32"`
	MsgsH float32 `buf:"f32"`
}

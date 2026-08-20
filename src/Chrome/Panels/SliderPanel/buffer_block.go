package SliderPanel

var _ = bufLayoutSpeedPanel{}

type bufLayoutSpeedPanel struct {
	BoxX float32 `buf:"f32"`
	BoxY float32 `buf:"f32"`
	BoxW float32 `buf:"f32"`
	BoxH float32 `buf:"f32"`

	RectX float32 `buf:"f32"`
	RectY float32 `buf:"f32"`
	RectW float32 `buf:"f32"`
	RectH float32 `buf:"f32"`

	Selected uint8 `buf:"u8"`

	NumText []byte `buf:"bytes"`
	NumLen  uint32 `buf:"u32"`

	DenText []byte `buf:"bytes"`
	DenLen  uint32 `buf:"u32"`

	TrackX float32 `buf:"f32"`
	TrackY float32 `buf:"f32"`
	TrackW float32 `buf:"f32"`
	TrackH float32 `buf:"f32"`
}

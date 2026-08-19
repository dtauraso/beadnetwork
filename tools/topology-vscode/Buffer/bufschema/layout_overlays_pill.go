package bufschema

type bufLayoutOverlaysPill struct {
	ScrollY float32 `buf:"f32"`

	PillX float32 `buf:"f32"`
	PillY float32 `buf:"f32"`
	PillW float32 `buf:"f32"`
	PillH float32 `buf:"f32"`

	Open   uint8 `buf:"u8"`
	Active uint8 `buf:"u8"`

	PopoverX float32 `buf:"f32"`
	PopoverY float32 `buf:"f32"`
	PopoverW float32 `buf:"f32"`
	PopoverH float32 `buf:"f32"`

	LabelText []byte `buf:"bytes"`

	RowKind  uint8 `buf:"u8"`
	RowDepth uint8 `buf:"u8"`

	RowX float32 `buf:"f32"`
	RowY float32 `buf:"f32"`
	RowW float32 `buf:"f32"`
	RowH float32 `buf:"f32"`

	RowTextData []byte `buf:"bytes"`
	RowTextLen  uint32 `buf:"u32"`

	RowIconData []byte `buf:"bytes"`
	RowIconLen  uint32 `buf:"u32"`

	RowOn       uint8 `buf:"u8"`
	RowDisabled uint8 `buf:"u8"`

	RowCountOn  uint32 `buf:"u32"`
	RowCountAll uint32 `buf:"u32"`

	CountX float32 `buf:"f32"`
	CountY float32 `buf:"f32"`
	CountW float32 `buf:"f32"`
	CountH float32 `buf:"f32"`
}

package NodesDropdown

var _ = bufLayoutNodesPill{}

type bufLayoutNodesPill struct {
	PillX float32 `buf:"f32"`
	PillY float32 `buf:"f32"`
	PillW float32 `buf:"f32"`
	PillH float32 `buf:"f32"`

	Open uint8 `buf:"u8"`

	PopoverX float32 `buf:"f32"`
	PopoverY float32 `buf:"f32"`
	PopoverW float32 `buf:"f32"`
	PopoverH float32 `buf:"f32"`

	LabelText []byte `buf:"bytes"`

	RowX float32 `buf:"f32"`
	RowY float32 `buf:"f32"`
	RowW float32 `buf:"f32"`
	RowH float32 `buf:"f32"`

	RowOpen uint8 `buf:"u8"`

	RowKindText []byte `buf:"bytes"`
	RowKindLen  uint32 `buf:"u32"`

	RowFillText   []byte `buf:"bytes"`
	RowFillLen    uint32 `buf:"u32"`
	RowStrokeText []byte `buf:"bytes"`
	RowStrokeLen  uint32 `buf:"u32"`

	SwatchX float32 `buf:"f32"`
	SwatchY float32 `buf:"f32"`
	SwatchW float32 `buf:"f32"`
	SwatchH float32 `buf:"f32"`

	RowDescText []byte `buf:"bytes"`
	RowDescLen  uint32 `buf:"u32"`

	DescX float32 `buf:"f32"`
	DescY float32 `buf:"f32"`
	DescW float32 `buf:"f32"`
	DescH float32 `buf:"f32"`

	RefusedCount uint32  `buf:"u32"`
	RefusedX     float32 `buf:"f32"`
	RefusedY     float32 `buf:"f32"`
	RefusedW     float32 `buf:"f32"`
	RefusedH     float32 `buf:"f32"`
	RefusedText  []byte  `buf:"bytes"`
}

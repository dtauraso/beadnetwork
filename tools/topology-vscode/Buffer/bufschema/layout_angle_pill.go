package bufschema

type bufLayoutAnglePill struct {
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

	StepX float32 `buf:"f32"`
	StepY float32 `buf:"f32"`
	StepW float32 `buf:"f32"`
	StepH float32 `buf:"f32"`

	StepNameText []byte `buf:"bytes"`
	StepNameLen  uint32 `buf:"u32"`

	StepShownText []byte `buf:"bytes"`
	StepShownLen  uint32 `buf:"u32"`

	StepValueRow int32 `buf:"i32"`
	StepDenom    int32 `buf:"i32"`

	StepUpX float32 `buf:"f32"`
	StepUpY float32 `buf:"f32"`
	StepUpW float32 `buf:"f32"`
	StepUpH float32 `buf:"f32"`

	StepDownX float32 `buf:"f32"`
	StepDownY float32 `buf:"f32"`
	StepDownW float32 `buf:"f32"`
	StepDownH float32 `buf:"f32"`

	StepUpOn   uint8 `buf:"u8"`
	StepDownOn uint8 `buf:"u8"`

	GroupX float32 `buf:"f32"`
	GroupY float32 `buf:"f32"`
	GroupW float32 `buf:"f32"`
	GroupH float32 `buf:"f32"`

	GroupOpen uint8 `buf:"u8"`

	GroupHeadText []byte `buf:"bytes"`
	GroupHeadLen  uint32 `buf:"u32"`
}

package bufschema

type bufLayoutRulesPanel struct {
	ClipY   float32 `buf:"f32"`
	ClipH   float32 `buf:"f32"`
	ScrollY float32 `buf:"f32"`

	BoxX float32 `buf:"f32"`
	BoxY float32 `buf:"f32"`
	BoxW float32 `buf:"f32"`
	BoxH float32 `buf:"f32"`

	Open uint8 `buf:"u8"`

	ToggleX float32 `buf:"f32"`
	ToggleY float32 `buf:"f32"`
	ToggleW float32 `buf:"f32"`
	ToggleH float32 `buf:"f32"`

	ToggleText []byte `buf:"bytes"`

	RowKind  uint8 `buf:"u8"`
	RowDepth uint8 `buf:"u8"`

	RowX float32 `buf:"f32"`
	RowY float32 `buf:"f32"`
	RowW float32 `buf:"f32"`
	RowH float32 `buf:"f32"`

	RowTextData []byte `buf:"bytes"`
	RowTextLen  uint32 `buf:"u32"`

	RowGlyphData []byte `buf:"bytes"`
	RowGlyphLen  uint32 `buf:"u32"`

	RowFree uint8 `buf:"u8"`

	RowNodeRow int32 `buf:"i32"`
	RowEdgeRow int32 `buf:"i32"`

	RowCheck  uint8   `buf:"u8"`
	RowCheckX float32 `buf:"f32"`
	RowCheckY float32 `buf:"f32"`
	RowCheckW float32 `buf:"f32"`
	RowCheckH float32 `buf:"f32"`

	RowValue  uint8   `buf:"u8"`
	RowValueX float32 `buf:"f32"`
	RowValueY float32 `buf:"f32"`
	RowValueW float32 `buf:"f32"`
	RowValueH float32 `buf:"f32"`

	RowSharedX float32 `buf:"f32"`
	RowSharedY float32 `buf:"f32"`
	RowSharedW float32 `buf:"f32"`
	RowSharedH float32 `buf:"f32"`

	RowEditing uint8 `buf:"u8"`

	DraftText []byte  `buf:"bytes"`
	DraftX    float32 `buf:"f32"`
	DraftY    float32 `buf:"f32"`
	DraftW    float32 `buf:"f32"`
	DraftH    float32 `buf:"f32"`

	MenuOpen      uint8   `buf:"u8"`
	MenuAnchorRow int32   `buf:"i32"`
	MenuX         float32 `buf:"f32"`
	MenuY         float32 `buf:"f32"`
	MenuW         float32 `buf:"f32"`
	MenuH         float32 `buf:"f32"`

	MenuRowX float32 `buf:"f32"`
	MenuRowY float32 `buf:"f32"`
	MenuRowW float32 `buf:"f32"`
	MenuRowH float32 `buf:"f32"`

	MenuCheckX float32 `buf:"f32"`
	MenuCheckY float32 `buf:"f32"`
	MenuCheckW float32 `buf:"f32"`
	MenuCheckH float32 `buf:"f32"`

	MenuLabelData []byte `buf:"bytes"`
	MenuLabelLen  uint32 `buf:"u32"`

	MenuNodeRow int32 `buf:"i32"`
}

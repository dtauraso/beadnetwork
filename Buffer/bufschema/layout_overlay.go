package bufschema

type bufLayoutOverlay struct {
	SceneTori    uint8 `buf:"u8"`
	ScenePoles   uint8 `buf:"u8"`
	NodePoles    uint8 `buf:"u8"`
	Handholds    uint8 `buf:"u8"`
	LabelsGlobal uint8 `buf:"u8"`
	OverlaysVis  uint8 `buf:"u8"`

	NodeBody      uint8 `buf:"u8"`
	NodeRing      uint8 `buf:"u8"`
	RingPick      uint8 `buf:"u8"`
	SelectionRing uint8 `buf:"u8"`
	HoverRing     uint8 `buf:"u8"`
	ReachSphere   uint8 `buf:"u8"`

	SceneVectors uint8 `buf:"u8"`

	RuleChannels uint8 `buf:"u8"`

	DragNodeRow int32 `buf:"i32"`

	EditRefused uint32 `buf:"u32"`

	SceneEditable uint8 `buf:"u8"`

	SceneKinds uint32 `buf:"u32"`

	GroupLenTime  float32 `buf:"f32"`
	GroupLenInput float32 `buf:"f32"`
	GroupLenGate  float32 `buf:"f32"`

	Speed float32 `buf:"f32"`
}

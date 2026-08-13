package bufschema

type bufLayoutOverlay struct {
	SceneTori      uint8 `buf:"u8"`
	ScenePoles     uint8 `buf:"u8"`
	NodePoles      uint8 `buf:"u8"`
	SelSpherePoles uint8 `buf:"u8"`
	Handholds      uint8 `buf:"u8"`
	LabelsGlobal   uint8 `buf:"u8"`
	OverlaysVis    uint8 `buf:"u8"`

	NodeBody      uint8 `buf:"u8"`
	NodeRing      uint8 `buf:"u8"`
	RingPick      uint8 `buf:"u8"`
	SelectionRing uint8 `buf:"u8"`
	HoverRing     uint8 `buf:"u8"`
	ReachSphere   uint8 `buf:"u8"`

	// SceneVectors draws each node's stored scene vector — the line from the
	// scene sphere's centre to that node's centre.
	SceneVectors uint8 `buf:"u8"`

	// CommEdges draws the edges a node uses to tell another node where the
	// constraint it holds puts it — an input node's outgoing paths, which
	// carry a position to each neighbour rather than a bead. They take the
	// place of the animation edge along those same paths while it is on.
	CommEdges uint8 `buf:"u8"`

	DragNodeRow int32 `buf:"i32"`

	EditRefused uint32 `buf:"u32"`

	SceneEditable uint8 `buf:"u8"`

	SceneKinds uint32 `buf:"u32"`

	GroupLenTime  float32 `buf:"f32"`
	GroupLenInput float32 `buf:"f32"`
	GroupLenGate  float32 `buf:"f32"`

	Speed float32 `buf:"f32"`
}

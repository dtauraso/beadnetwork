package Buffer

// bufLayoutOverlay defines the overlay visibility column block (always 1 row).
// Matched from KindSceneTori/ScenePoles/…/OverlaysVis trace events.
// Field order matches overlayState in overlay_gen.go.
type bufLayoutOverlay struct {
	SceneTori      uint8 `buf:"u8"` // 1 = polar-guide tori visible
	ScenePoles     uint8 `buf:"u8"` // 1 = scene-center pole frame visible
	NodePoles      uint8 `buf:"u8"` // 1 = per-node pole frames visible
	SelSpherePoles uint8 `buf:"u8"` // 1 = selection-sphere pole axes visible
	Handholds      uint8 `buf:"u8"` // 1 = rotation grab-sphere handholds visible
	LabelsGlobal   uint8 `buf:"u8"` // 1 = all node labels visible
	OverlaysVis    uint8 `buf:"u8"` // 1 = master overlays toggle on
	// The NODE-LOCAL drawings (OVERLAY_FLAG_NAMES' second half). Everything above is scene
	// furniture drawn around the nodes; these six are the node itself. Order matches the
	// flag list, same as the block's first half.
	NodeBody      uint8 `buf:"u8"` // 1 = node sphere drawn
	NodeRing      uint8 `buf:"u8"` // 1 = per-node border ring drawn
	RingPick      uint8 `buf:"u8"` // 1 = ring click-band drawn (it is hit-testable either way)
	SelectionRing uint8 `buf:"u8"` // 1 = selected node's ring + halo drawn
	HoverRing     uint8 `buf:"u8"` // 1 = hovered node's ring drawn
	ReachSphere   uint8 `buf:"u8"` // 1 = selected node's reach-sphere rings drawn
	// DragNodeRow is the row index (into the Node block) of the node currently
	// being dragged by the gesture FSM (nodes/Wiring/gesture.go g.dragNode),
	// or -1 when no drag is in progress. Identity rides row index, not a
	// name/id sidecar; TS resolves the human name from that row's Label.
	DragNodeRow int32 `buf:"i32"` // dragged node's row index, -1 = not dragging
	// EditRefused COUNTS refused structural edits (scene_structure.go: a drop that cannot be
	// connected, a delete of a row holding no node). It is a counter, not a flag, because a
	// second refusal must be distinguishable from the first: a flag that is already 1 says
	// nothing when the same mistake is made twice, and the whole point of this column is that
	// the gesture did nothing and the person needs telling. TS shows a message when the count
	// it last saw goes up.
	//
	// The REASON is not here. It rides stderr into the output channel and
	// .probe/go-errors.jsonl, where the detail belongs; the screen only has to say that the
	// edit was refused, which is the part a person cannot otherwise see.
	EditRefused uint32 `buf:"u32"` // count of refused structural edits this run
	// SceneEditable is SceneTab.Editable for the tree actually loaded — whether this scene
	// takes structural edits at all. Streamed rather than inferred in TS from a scene name,
	// for the same reason the tab LIST is streamed: which scenes exist and what they allow
	// is Go's, and the editor renders what it is told.
	SceneEditable uint8 `buf:"u8"` // 1 = this scene takes node create/delete
	// SceneKinds is the BITMASK of kind ids this scene accepts (bit N = the kind whose
	// KindId is N). All bits set means no restriction. The palette offers exactly these, so
	// a kind a scene has no place for is never draggable in it — rather than draggable and
	// then refused, which teaches nothing and looks broken.
	SceneKinds uint32 `buf:"u32"` // kind-id bitmask this scene accepts
	// GroupLenTime/GroupLenInput/GroupLenGate are the "distance home button" toolbar
	// panel's 3 read-only group max-pair-lengths (nodes/Wiring/distance_groups.go's
	// distanceGroupOrder: time, input, gate). Computed fresh every VIEW-frame emit from
	// the live node centers (max over the group's pairs of |center(target)-center(source)|,
	// mirroring reachRFromPolar's max-over-edges loop) — Go owns the group definitions and
	// the math; the panel only reflects these 3 numbers and fires an edit-update on an
	// arrow click. 0 when a group's centers aren't resolvable yet.
	GroupLenTime  float32 `buf:"f32"` // time-nodes group's current max pair length
	GroupLenInput float32 `buf:"f32"` // input-node group's current max pair length
	GroupLenGate  float32 `buf:"f32"` // gate/pulse-nodes group's current max pair length
	// Speed is the current playback-speed multiplier (one of the SpeedSlider's six table
	// values: 0, 0.25, 0.5, 0.75, 1, 2). Go owns it (RunStdinReader's clock/speed edit
	// handler, seeded at load from view/speed.json); this column is a READ-ONLY reflect so
	// the webview slider can show the persisted/live value instead of holding its own local
	// default (memory/feedback_reflect_dont_create_store.md) — it is never derived from this
	// column, only displayed from it.
	Speed float32 `buf:"f32"` // current playback-speed multiplier
}

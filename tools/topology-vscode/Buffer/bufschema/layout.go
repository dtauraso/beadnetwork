package bufschema

const BufLayoutVersion = 48

const BufInteriorSlotsPerNode = 4

var _ = [...]any{
	bufLayoutNode{},
	bufLayoutInterior{},
	bufLayoutEdge{},
	bufLayoutEdgeBead{},
	bufLayoutNodeRingPoint{},
	bufLayoutBeadRingPoint{},
	bufLayoutTiltArrow{},
	bufLayoutChannelVector{},
	bufLayoutCamera{},
	bufLayoutOverlay{},
	bufLayoutPanel{},
	bufLayoutScene{},
	bufLayoutEvent{},
}

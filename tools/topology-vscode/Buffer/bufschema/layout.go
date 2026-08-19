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
	bufLayoutRecv{},
	bufLayoutFire{},
	bufLayoutSend{},
	bufLayoutArrive{},
	bufLayoutBreadcrumb{},
	bufLayoutSpeedPanel{},
	bufLayoutTiltPanel{},
	bufLayoutAnglePill{},
	bufLayoutNodesPill{},
	bufLayoutOverlaysPill{},
	bufLayoutFitChip{},
	bufLayoutTabStrip{},
	bufLayoutRulesPanel{},
	bufLayoutPointerTarget{},
}

package bufschema

const BufLayoutVersion = 48

const BufInteriorSlotsPerNode = 4

var _ = [...]any{
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
}

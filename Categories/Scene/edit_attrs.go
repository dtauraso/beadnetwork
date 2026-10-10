package Scene

var (
	attrSelected       = attrIndex("selected")
	attrLatticePoints  = attrIndex("latticePoints")
	attrCreate         = attrIndex("create")
	attrDelete         = attrIndex("delete")
	attrViewport       = attrIndex("viewport")
	attrTiltVectorPhi  = attrIndex("phi")
	attrTiltVectorRst  = attrIndex("reset")
	attrTiltVectorStrt = attrIndex("start")
)

func IsViewport(attr byte) bool { return attr == attrViewport }

package Scene

import "github.com/dtauraso/wirefold/src/Input/Codec"

var (
	attrSelected       = Codec.AttrIndex("selected")
	attrLatticePoints  = Codec.AttrIndex("latticePoints")
	attrCreate         = Codec.AttrIndex("create")
	attrDelete         = Codec.AttrIndex("delete")
	attrViewport       = Codec.AttrIndex("viewport")
	attrTiltVectorPhi  = Codec.AttrIndex("phi")
	attrTiltVectorRst  = Codec.AttrIndex("reset")
	attrTiltVectorStrt = Codec.AttrIndex("start")
)

package Node

import "github.com/dtauraso/wirefold/src/Input/Codec"

var (
	attrDragPhi      = Codec.AttrIndex("dragPhi")
	attrDragMaxTheta = Codec.AttrIndex("dragMaxTheta")
	attrDragActive   = Codec.AttrIndex("dragActive")
	attrKindActive   = Codec.AttrIndex("kindActive")

	attrSelfDragPhi      = Codec.AttrIndex("selfDragPhi")
	attrSelfDragMaxTheta = Codec.AttrIndex("selfDragMaxTheta")
	attrSelfDragActive   = Codec.AttrIndex("selfDragActive")

	attrDragR     = Codec.AttrIndex("dragR")
	attrSelfDragR = Codec.AttrIndex("selfDragR")
)

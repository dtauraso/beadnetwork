package Scene

var (
	attrSelected      = attrIndex("selected")
	attrLatticePoints = attrIndex("latticePoints")
	attrCreate        = attrIndex("create")
	attrDelete        = attrIndex("delete")
	attrViewport      = attrIndex("viewport")
)

func IsViewport(attr byte) bool { return attr == attrViewport }

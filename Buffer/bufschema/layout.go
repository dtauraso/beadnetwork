// staticcheck anchor (schemaTypes below) that references every block type so

package bufschema

const BufLayoutVersion = 43

const BufInteriorSlotsPerNode = 4

// staticcheck. They are schema sources: the generator reads them via AST at

var _ = [...]any{
	bufLayoutNode{},
	bufLayoutInterior{},
	bufLayoutEdge{},
	bufLayoutEdgeBead{},
	bufLayoutCamera{},
	bufLayoutOverlay{},
	bufLayoutPanel{},
	bufLayoutScene{},
	bufLayoutEvent{},
}

// staticcheck anchor (schemaTypes below) that references every block type so

package Buffer

const BufLayoutVersion = 41

const BufInteriorSlotsPerNode = 4

// staticcheck. They are schema sources: the generator reads them via AST at

var _ = [...]any{
	bufLayoutNode{},
	bufLayoutChainBead{},
	bufLayoutInterior{},
	bufLayoutEdge{},
	bufLayoutCamera{},
	bufLayoutOverlay{},
	bufLayoutScene{},
	bufLayoutEvent{},
}

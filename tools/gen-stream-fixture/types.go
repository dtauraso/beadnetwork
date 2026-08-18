package main

type chainBeadFixture struct {
	OX       float32 `json:"ox"`
	OY       float32 `json:"oy"`
	OZ       float32 `json:"oz"`
	Lit      uint8   `json:"lit"`
	LitValue int32   `json:"litValue"`
}

type nodeFrameFixture struct {
	Tick      uint32  `json:"tick"`
	NodeRow   int32   `json:"nodeRow"`
	NodeId    int32   `json:"nodeId"`
	CX        float32 `json:"cx"`
	CY        float32 `json:"cy"`
	CZ        float32 `json:"cz"`
	Radius    float32 `json:"radius"`
	PolePhi   float32 `json:"polePhi"`
	PoleTheta float32 `json:"poleTheta"`

	TopTiltVectorLen float32 `json:"topTiltVectorLen"`

	TopTiltVectorIdx int32 `json:"topTiltVectorIdx"`

	TopTiltVectorPhi float32 `json:"topTiltVectorPhi"`

	BottomTiltVectorPhi float32 `json:"bottomTiltVectorPhi"`

	CoplanarNormalPhi float32 `json:"coplanarNormalPhi"`

	ReceivedVectorLen float32 `json:"receivedVectorLen"`
	ReceivedVectorPhi float32 `json:"receivedVectorPhi"`
	Selected          uint8   `json:"selected"`
	KindID            uint8   `json:"kindId"`
	Hovered           uint8   `json:"hovered"`
	LatchedSel        uint8   `json:"latchedSel"`

	LatticePoints uint8 `json:"latticePoints"`

	RoundsToParallel int32 `json:"roundsToParallel"`

	MsgsToParallel int32              `json:"msgsToParallel"`
	ChainBeads     []chainBeadFixture `json:"chainBeads"`
	Label          string             `json:"label"`
	Hex            string             `json:"hex"`
}

type edgeBeadFixture struct {
	X       float32 `json:"x"`
	Y       float32 `json:"y"`
	Z       float32 `json:"z"`
	Value   int32   `json:"value"`
	EdgeRow int32   `json:"edgeRow"`
}

type beadFrameFixture struct {
	Tick    uint32            `json:"tick"`
	NodeRow int32             `json:"nodeRow"`
	Beads   []edgeBeadFixture `json:"beads"`
	Hex     string            `json:"hex"`
}

type edgeFrameFixture struct {
	Tick       uint32  `json:"tick"`
	SX         float32 `json:"sx"`
	SY         float32 `json:"sy"`
	SZ         float32 `json:"sz"`
	EX         float32 `json:"ex"`
	EY         float32 `json:"ey"`
	EZ         float32 `json:"ez"`
	SrcNodeRow int32   `json:"srcNodeRow"`
	DstNodeRow int32   `json:"dstNodeRow"`
	DeltaR     float32 `json:"deltaR"`
	Label      string  `json:"label"`
	Hex        string  `json:"hex"`
}

type interiorFrameFixture struct {
	Tick uint32 `json:"tick"`

	Present []int     `json:"present"`
	Value   []int32   `json:"value"`
	OX      []float32 `json:"ox"`
	OY      []float32 `json:"oy"`
	OZ      []float32 `json:"oz"`
	Hex     string    `json:"hex"`
}

type streamFixture struct {
	NodeFrame     nodeFrameFixture     `json:"nodeFrame"`
	EdgeFrame     edgeFrameFixture     `json:"edgeFrame"`
	BeadFrame     beadFrameFixture     `json:"beadFrame"`
	InteriorFrame interiorFrameFixture `json:"interiorFrame"`
}

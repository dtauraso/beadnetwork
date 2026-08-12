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
	SphereR   float32 `json:"sphereR"`
	VRX       float32 `json:"vrx"`
	VRY       float32 `json:"vry"`
	VRZ       float32 `json:"vrz"`
	FRX       float32 `json:"frx"`
	FRY       float32 `json:"fry"`
	FRZ       float32 `json:"frz"`
	PoleTheta float32 `json:"poleTheta"`
	PolePhi   float32 `json:"polePhi"`

	RingAxisTheta float32 `json:"ringAxisTheta"`
	RingAxisPhi   float32 `json:"ringAxisPhi"`

	TopTiltVectorLen float32 `json:"topTiltVectorLen"`

	TopTiltVectorTheta float32 `json:"topTiltVectorTheta"`

	BottomTiltVectorTheta float32 `json:"bottomTiltVectorTheta"`

	CoplanarNormalTheta float32 `json:"coplanarNormalTheta"`

	ReceivedVectorLen   float32 `json:"receivedVectorLen"`
	ReceivedVectorTheta float32 `json:"receivedVectorTheta"`
	Selected            uint8   `json:"selected"`
	KindID              uint8   `json:"kindId"`
	Hovered             uint8   `json:"hovered"`
	LatchedSel          uint8   `json:"latchedSel"`

	LatticePoints uint8 `json:"latticePoints"`

	RoundsToParallel int32 `json:"roundsToParallel"`

	MsgsToParallel int32              `json:"msgsToParallel"`
	ChainBeads     []chainBeadFixture `json:"chainBeads"`
	Label          string             `json:"label"`
	Hex            string             `json:"hex"`
}

type edgeFrameFixture struct {
	Tick     uint32  `json:"tick"`
	SX       float32 `json:"sx"`
	SY       float32 `json:"sy"`
	SZ       float32 `json:"sz"`
	EX       float32 `json:"ex"`
	EY       float32 `json:"ey"`
	EZ       float32 `json:"ez"`
	Selected uint8   `json:"selected"`
	Label    string  `json:"label"`
	Hex      string  `json:"hex"`
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
	InteriorFrame interiorFrameFixture `json:"interiorFrame"`
}

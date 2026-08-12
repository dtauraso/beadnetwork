// types.go holds the fixture's own JSON shapes — one struct per stream frame kind, mirroring
// the fields tools/topology-vscode/test/buffer/stream-fixture.test.ts decodes with the REAL
// TS decoders. See main.go's package doc for the fixture's purpose and regeneration command.
package main

// chainBeadFixture is ONE node-local chain-bead offset row (Buffer bufLayoutChainBead).
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
	// The DRAWN ring's axis, distinct from the navigation pole above — see
	// Buffer/layout.go's RingAxisTheta/RingAxisPhi for why they are two values.
	RingAxisTheta float32 `json:"ringAxisTheta"`
	RingAxisPhi   float32 `json:"ringAxisPhi"`
	// The node's own drawn vector length along that axis; 0 means it draws none.
	TopTiltVectorLen float32 `json:"topTiltVectorLen"`
	// The vector's OWN direction — separate from RingAxisTheta/Phi above (Buffer/layout.go's
	// TopTiltVectorTheta). There is no φ: the tilt-vector model is θ-only
	// (task/drop-tilt-vector-phi).
	TopTiltVectorTheta float32 `json:"topTiltVectorTheta"`
	// The BOTTOM tilt vector's direction — a half turn in θ from the top (Buffer/layout.go's
	// BottomTiltVectorTheta); it shares the top's length column.
	BottomTiltVectorTheta float32 `json:"bottomTiltVectorTheta"`
	// The SECOND vector's direction — a quarter turn from the first, in the ring's plane.
	CoplanarNormalTheta float32 `json:"coplanarNormalTheta"`
	// The THIRD vector: the direction last received on this node's tilt-vector channel
	// (Buffer/layout.go's ReceivedVectorLen/Theta); 0 length means nothing received yet.
	ReceivedVectorLen   float32 `json:"receivedVectorLen"`
	ReceivedVectorTheta float32 `json:"receivedVectorTheta"`
	Selected            uint8   `json:"selected"`
	KindID              uint8   `json:"kindId"`
	Hovered             uint8   `json:"hovered"`
	LatchedSel          uint8   `json:"latchedSel"`
	// LatticePoints is this node's own pair-lattice point count (Buffer/layout.go's
	// LatticePoints) — the N the four θ columns above were converted against.
	LatticePoints uint8 `json:"latticePoints"`
	// RoundsToParallel is this node's own rounds-to-rest count (Buffer/layout.go's
	// RoundsToParallel) — vector-exchange rounds between START and its rule settling.
	RoundsToParallel int32 `json:"roundsToParallel"`
	// MsgsToParallel is the same span in vector-channel messages (Buffer/layout.go).
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
	// Present is []int, not []uint8: Go's encoding/json marshals []uint8 as a base64
	// string (it treats it as []byte), which would silently strip the fixture's own
	// human-readable expected values. []int marshals as a plain JSON number array.
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

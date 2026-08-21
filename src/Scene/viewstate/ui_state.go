package viewstate

import (
	"fmt"
	"github.com/dtauraso/wirefold/src/Chrome/Panels/Panel"
	"github.com/dtauraso/wirefold/src/Overlay"
	"math"
	"os"

	"github.com/dtauraso/wirefold/src/Camera"
	"github.com/dtauraso/wirefold/src/Polar/polar"
	"github.com/dtauraso/wirefold/src/Input/gesturefsm"
	"github.com/dtauraso/wirefold/src/Input/inputcodec"
	"github.com/dtauraso/wirefold/src/Node/movemsg"
	"github.com/dtauraso/wirefold/src/Polar/polarindex"
	"github.com/dtauraso/wirefold/src/Chrome/Panels/PolarRulesPanel"
	"github.com/dtauraso/wirefold/src/Chrome/Pills/AngleDropdown"
	"github.com/dtauraso/wirefold/src/Chrome/Pills/NodesDropdown"
	"github.com/dtauraso/wirefold/src/Chrome/Tabs"
	"github.com/dtauraso/wirefold/src/Scene/selectionstate"
	"github.com/dtauraso/wirefold/src/spatial"
	"github.com/dtauraso/wirefold/src/Buffer/colstream"
)

type UIState struct {
	EditRefused uint32

	SceneEditable bool

	SceneKinds uint32

	SceneSphere polar.SceneSphere

	Constants polarindex.SceneConstants

	singletonCols *colstream.ColumnSet

	rulesValues     *PolarRulesPanel.ValueWriter
	nodesPillValues *NodesDropdown.ValueWriter
	anglePillValues *AngleDropdown.ValueWriter
	tabStripValues  *Tabs.ValueWriter

	OwnerCounts struct{ Nodes, Edges int32 }

	ClockDivisor float64

	VP gesturefsm.ViewpointState

	OV Overlay.OverlayState

	PN Panel.PanelState

	Gest gesturefsm.GestureState

	Sel selectionstate.SelectionState

	LatchedNode string

	LastDraggedNode string

	Speed float64

	LatticePoints int32

	NodeRowFor func(id string) (int32, bool)

	TiltRows   []int32
	TiltLabels []string

	ViewW, ViewH float64

	Pointer PointerTarget

	OverlaysScroll float32

	RulesScroll float32

	AngleOpen      bool
	AngleGroupOpen map[int32]bool

	NodesOpen    bool
	NodesRowOpen map[uint8]bool

	SceneTabNames    []string
	SceneTabSelected int

	RuleNodes     []PolarRulesPanel.Node
	RuleEdit      PolarRulesPanel.Edit
	RuleSharedRow int32

	PlacingKind    uint8
	PlacingPending bool

	viewOut        viewClaimedStream
	ViewBuildFrame ViewFrameBuilder
	viewTick       uint32
	viewClaimed    bool
}

func (ui *UIState) SetSelectionUI(sendMove func(id string, msg movemsg.Msg), node, edge string) {
	prevNode := ui.Sel.Selected
	ui.Sel.Selected = node
	ui.Sel.SelectedEdge = edge
	if prevNode != "" && prevNode != node {
		sendMove(prevNode, movemsg.Msg{Kind: movemsg.KindSelect, NodeID: prevNode, Bool: false})
	}
	if node != "" && node != prevNode {
		sendMove(node, movemsg.Msg{Kind: movemsg.KindSelect, NodeID: node, Bool: true})
	}
	if node != "" && node != ui.LatchedNode {
		prevLatched := ui.LatchedNode
		ui.LatchedNode = node
		if prevLatched != "" {
			sendMove(prevLatched, movemsg.Msg{Kind: movemsg.KindLatched, NodeID: prevLatched, Bool: false})
		}
		sendMove(node, movemsg.Msg{Kind: movemsg.KindLatched, NodeID: node, Bool: true})
	}
}

func (ui *UIState) DropPointFromNDC(ndcX, ndcY float64) (spatial.Vec3, bool) {
	vp := ui.VP.Viewpoint
	eye := Camera.EyeOf(vp)
	basis := Camera.BasisFromViewpoint(vp.Pos, vp.Up)
	dir := Camera.RayDirThroughNDC(ndcX, ndcY, basis, ui.Gest.Fov, ui.Gest.Rect.Aspect())
	forward := basis.Pole.Scale(-1)
	denom := dir.Dot(forward)
	if denom == 0 {
		return spatial.Vec3{}, false
	}
	t := ui.SceneSphere.Center.Sub(eye).Dot(forward) / denom
	hit := eye.Add(dir.Scale(t))
	if math.IsNaN(hit.X) || math.IsInf(hit.X, 0) {
		return spatial.Vec3{}, false
	}
	return hit, true
}

func (ui *UIState) SetHoverUI(sendMove func(id string, msg movemsg.Msg), node, port string, isInput bool) {
	prevNode := ui.Sel.HoverNode
	ui.Sel.HoverNode, ui.Sel.HoverPort, ui.Sel.HoverInput = node, port, isInput
	if prevNode != "" && prevNode != node {
		sendMove(prevNode, movemsg.Msg{Kind: movemsg.KindHover, NodeID: prevNode, Bool: false})
	}
	if node != "" {
		sendMove(node, movemsg.Msg{Kind: movemsg.KindHover, NodeID: node, Bool: true, Port: port, IsInput: isInput})
	}
}

func (ui *UIState) DragPlaneHit(ev inputcodec.RawInputMsg) (hit spatial.Vec3, ok bool) {
	g := &ui.Gest
	vp := ui.VP.Viewpoint
	eye := Camera.EyeOf(vp)
	basis := Camera.BasisFromViewpoint(vp.Pos, vp.Up)
	nx, ny := g.PixelToNDC(ev.X, ev.Y)
	dir := Camera.RayDirThroughNDC(nx, ny, basis, ui.FovDeg(), g.Rect.Aspect())
	forward := basis.Pole.Scale(-1)
	denom := dir.Dot(forward)
	if denom == 0 {
		return spatial.Vec3{}, false
	}
	t := g.DragStartCenter.Sub(eye).Dot(forward) / denom
	hit = eye.Add(dir.Scale(t))
	if math.IsNaN(hit.X) || math.IsInf(hit.X, 0) {
		return spatial.Vec3{}, false
	}
	return hit, true
}

func (ui *UIState) OrbitViewpoint(from, to Camera.Dir) {
	ui.VP.OrbitViewpoint(from, to)
}
func (ui *UIState) OrbitLockedViewpoint(from, to Camera.Dir) {
	ui.VP.OrbitLockedViewpoint(from, to)
}
func (ui *UIState) ZoomViewpoint(factor float64) {
	ui.VP.ZoomViewpoint(factor)
}

func (ui *UIState) RefuseStructuralEdit(why string) {
	fmt.Fprintf(os.Stderr, "structural edit refused: %s\n", why)

	ui.EditRefused++
}

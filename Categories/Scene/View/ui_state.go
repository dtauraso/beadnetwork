package View

import (
	"fmt"
	"math"
	"os"

	NodeDrag "github.com/dtauraso/beadnetwork/Categories/Node/Drag"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Panel"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/SliderPanel"
	Flags "github.com/dtauraso/beadnetwork/Categories/Scene/View/Flags"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/PolarRulesPanel"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/TiltPanel"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/AngleDropdown"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/FitButton"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/NodesDropdown"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Tabs"
	"github.com/dtauraso/beadnetwork/Categories/Node"
	"github.com/dtauraso/beadnetwork/Categories/Polar/polar"
	"github.com/dtauraso/beadnetwork/Categories/Polar/polarindex"
	bead "github.com/dtauraso/beadnetwork/Categories/Ring/Bead"
	NodeShape "github.com/dtauraso/beadnetwork/Categories/Ring/NodeShape"
	"github.com/dtauraso/beadnetwork/Categories/Scene"
	"github.com/dtauraso/beadnetwork/Categories/Scene/Camera"
	"github.com/dtauraso/beadnetwork/Categories/Scene/Drag"
)

type UIState struct {
	EditRefused uint32

	SceneEditable bool

	SceneKinds uint32

	SceneSphere polar.SceneSphere

	Constants polarindex.SceneConstants

	sceneRoot string

	Counts Scene.CountsState

	OwnerCounts struct{ Nodes, Edges int32 }

	ClockDivisor float64

	VP Camera.ViewpointState

	OV Flags.OverlayState

	PN Panel.PanelState

	Gest Drag.GestureState

	Sel selectionState

	PersistOverlays func(Flags.OverlayState)
	PersistPanels   func(Panel.PanelState)
	PersistSphere   func(polar.SceneSphere)
	PersistSpeed    func(float64)
	PersistLattice  func(int32)

	LatchedNode string

	LastDraggedNode string

	Speed float64

	LatticePoints int32

	NodeRowFor func(id string) (int32, bool)

	Tilt TiltPanel.State

	ViewW, ViewH float64

	Pointer Panels.PointerTarget

	OverlaysPill Pills.State

	Slider     SliderPanel.State
	Fit        FitButton.State
	NodeRingPoints NodeShape.RingPointState
	BeadRingPoints bead.RingPointState
	PointerBlk Panels.State

	Angle AngleDropdown.State

	Nodes NodesDropdown.State

	TabStrip Tabs.State

	Rules PolarRulesPanel.State

	PlacingKind    uint8
	PlacingPending bool

	ViewBuildFrame ViewFrameBuilder
	viewTick       uint32
}

func (ui *UIState) SetSelectionUI(sendMove func(id string, msg Node.Msg), node, edge string) {
	prevNode := ui.Sel.selected
	ui.Sel.selected = node
	ui.Sel.selectedEdge = edge
	if prevNode != "" && prevNode != node {
		sendMove(prevNode, Node.Msg{NodeID: prevNode, Body: Node.Select{On: false}})
	}
	if node != "" && node != prevNode {
		sendMove(node, Node.Msg{NodeID: node, Body: Node.Select{On: true}})
	}
	if node != "" && node != ui.LatchedNode {
		prevLatched := ui.LatchedNode
		ui.LatchedNode = node
		if prevLatched != "" {
			sendMove(prevLatched, Node.Msg{NodeID: prevLatched, Body: Node.Latched{On: false}})
		}
		sendMove(node, Node.Msg{NodeID: node, Body: Node.Latched{On: true}})
	}
}

func (ui *UIState) DropPointFromNDC(ndcX, ndcY float64) (Vec3, bool) {
	vp := ui.VP.Viewpoint
	eye := Camera.EyeOf(vp)
	basis := Camera.BasisFromViewpoint(vp.Pos, vp.Up)
	dir := Camera.RayDirThroughNDC(ndcX, ndcY, basis, ui.Gest.Fov, ui.Gest.Rect.Aspect())
	forward := basis.Pole.Scale(-1)
	denom := dir.Dot(forward)
	if denom == 0 {
		return Vec3{}, false
	}
	t := ui.SceneSphere.Center.Sub(polar.Vec3(eye)).Dot(polar.Vec3(forward)) / denom
	hit := eye.Add(dir.Scale(t))
	if math.IsNaN(hit.X) || math.IsInf(hit.X, 0) {
		return Vec3{}, false
	}
	return Vec3(hit), true
}

func (ui *UIState) SetHoverUI(sendMove func(id string, msg Node.Msg), node, port string, isInput bool) {
	prevNode := ui.Sel.hoverNode
	ui.Sel.hoverNode, ui.Sel.hoverPort, ui.Sel.hoverInput = node, port, isInput
	if prevNode != "" && prevNode != node {
		sendMove(prevNode, Node.Msg{NodeID: prevNode, Body: Node.Hover{On: false}})
	}
	if node != "" {
		sendMove(node, Node.Msg{NodeID: node, Body: Node.Hover{On: true, Port: port, IsInput: isInput}})
	}
}

func (ui *UIState) DragPlaneHit(ev Drag.RawInputMsg) (hit Vec3, ok bool) {
	g := &ui.Gest
	vp := ui.VP.Viewpoint
	eye := Camera.EyeOf(vp)
	basis := Camera.BasisFromViewpoint(vp.Pos, vp.Up)
	nx, ny := g.PixelToNDC(ev.X, ev.Y)
	dir := Camera.RayDirThroughNDC(nx, ny, basis, ui.FovDeg(), g.Rect.Aspect())
	forward := basis.Pole.Scale(-1)
	denom := dir.Dot(forward)
	if denom == 0 {
		return Vec3{}, false
	}
	t := g.NodeDrag.StartCenter.Sub(NodeDrag.Vec3(eye)).Dot(NodeDrag.Vec3(forward)) / denom
	hit = Vec3(eye.Add(dir.Scale(t)))
	if math.IsNaN(hit.X) || math.IsInf(hit.X, 0) {
		return Vec3{}, false
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

type selectionState struct {
	selected     string
	selectedEdge string

	hoverNode  string
	hoverPort  string
	hoverInput bool
}

func (ui *UIState) HoverIs(node, port string, isInput bool) bool {
	return node == ui.Sel.hoverNode && port == ui.Sel.hoverPort && isInput == ui.Sel.hoverInput
}

func (ui *UIState) SelectedNode() (string, bool) {
	return ui.Sel.selected, ui.Sel.selected != ""
}

func (ui *UIState) SelectedEdge() string { return ui.Sel.selectedEdge }

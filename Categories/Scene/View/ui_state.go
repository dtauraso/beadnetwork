package View

import (
	"fmt"
	"os"

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
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polar"
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polarindex"
	bead "github.com/dtauraso/beadnetwork/Categories/Ring/Bead"
	NodeShape "github.com/dtauraso/beadnetwork/Categories/Ring/NodeShape"
	"github.com/dtauraso/beadnetwork/Categories/Scene"
	"github.com/dtauraso/beadnetwork/Categories/Scene/Camera"
	"github.com/dtauraso/beadnetwork/Categories/Scene/Drag"
)

type UIState struct {
	EditRefused uint32

	SceneEditable bool

	SceneNodesDraggable bool

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
	lastStrip      Tabs.Rect
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

func (ui *UIState) rotationScale() float64 {
	return Camera.RotationScale(ui.Gest.PressVP, Camera.Vec3(ui.SceneSphere.Center), ui.SceneSphere.Radius)
}

func (ui *UIState) OrbitViewpoint(from, to Camera.Dir) {
	ui.VP.OrbitViewpoint(ui.Gest.PressVP, from, to, ui.rotationScale())
}
func (ui *UIState) OrbitLockedViewpoint(from, to Camera.Dir) {
	ui.VP.OrbitLockedViewpoint(ui.Gest.PressVP, from, to, Camera.Vec3(ui.SceneSphere.Center))
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

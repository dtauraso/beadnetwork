package scenepersist

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepaths"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewstate"
	"github.com/dtauraso/wirefold/nodes/rowevent"

	"github.com/dtauraso/wirefold/tools/topology-vscode/OverlaysDropdown"
	T "github.com/dtauraso/wirefold/tools/topology-vscode/Trace"
)

func InstallOverlays(ui *viewstate.UIState, topologyPath string, tr *T.Trace) {
	ov, _ := OverlaysDropdown.LoadSceneOverlays(scenepaths.OverlaysFilePath(topologyPath))
	ui.OV.SetGuideVisibility(ov)

	ui.EmitViewFrame([]rowevent.RowEvent{
		{Kind: T.KindSceneTori, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1},
		{Kind: T.KindScenePoles, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1},
		{Kind: T.KindNodePoles, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1},
		{Kind: T.KindHandholds, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1},
		{Kind: T.KindLabelsGlobal, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1},
		{Kind: T.KindOverlaysVis, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1},
		{Kind: T.KindNodeBody, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1},
		{Kind: T.KindNodeRing, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1},
		{Kind: T.KindRingPick, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1},
		{Kind: T.KindSelectionRing, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1},
		{Kind: T.KindHoverRing, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1},
		{Kind: T.KindSceneVectors, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1},
		{Kind: T.KindRuleChannels, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1},
	})
}

func InstallPanels(ui *viewstate.UIState, topologyPath string) {
	pn, _ := OverlaysDropdown.LoadScenePanels(scenepaths.PanelsFilePath(topologyPath))
	ui.PN.SetPanelState(pn)
}

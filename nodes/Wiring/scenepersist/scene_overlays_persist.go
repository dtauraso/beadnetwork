package scenepersist

import (
	"encoding/json"

	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepaths"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewstate"
	"github.com/dtauraso/wirefold/nodes/rowevent"

	T "github.com/dtauraso/wirefold/Trace"
)

func InstallOverlays(ui *viewstate.UIState, topologyPath string, tr *T.Trace) {
	ov, _ := LoadSceneOverlays(scenepaths.OverlaysFilePath(topologyPath))
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
		{Kind: T.KindReachSphere, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1},
		{Kind: T.KindSceneVectors, NodeRow: -1, PortRow: -1, TargetRow: -1, TargetPortRow: -1, EdgeRow: -1},
	})
}

func WriteSceneOverlays(overlaysPath string, ov viewstate.OverlayState) error {
	obj := map[string]json.RawMessage{}

	setVisible := func(key string, visible bool) {
		if !visible {
			obj[key] = json.RawMessage("false")
		}
	}

	setHidden := func(key string, visible bool) {
		if !visible {
			obj[key] = json.RawMessage("true")
		}
	}

	setVisible("sceneToriVisible", ov.SceneToriVisible)
	setVisible("scenePolesVisible", ov.ScenePolesVisible)
	setVisible("nodePolesVisible", ov.NodePolesVisible)
	setVisible("handholdsVisible", ov.HandholdsVisible)
	setVisible("overlaysActive", ov.OverlaysVisible)
	setHidden("labelsGlobalHidden", ov.LabelsGlobalVisible)
	setVisible("nodeBodyVisible", ov.NodeBodyVisible)
	setVisible("nodeRingVisible", ov.NodeRingVisible)
	setVisible("ringPickVisible", ov.RingPickVisible)
	setVisible("selectionRingVisible", ov.SelectionRingVisible)
	setVisible("hoverRingVisible", ov.HoverRingVisible)
	setVisible("reachSphereVisible", ov.ReachSphereVisible)
	setVisible("sceneVectorsVisible", ov.SceneVectorsVisible)
	return jsonpersist.WriteJSONAtomic(overlaysPath, obj)
}

type sceneOverlaysFile struct {
	SceneToriVisible     *bool `json:"sceneToriVisible"`
	ScenePolesVisible    *bool `json:"scenePolesVisible"`
	NodePolesVisible     *bool `json:"nodePolesVisible"`
	HandholdsVisible     *bool `json:"handholdsVisible"`
	OverlaysActive       *bool `json:"overlaysActive"`
	LabelsGlobalHidden   *bool `json:"labelsGlobalHidden"`
	NodeBodyVisible      *bool `json:"nodeBodyVisible"`
	NodeRingVisible      *bool `json:"nodeRingVisible"`
	RingPickVisible      *bool `json:"ringPickVisible"`
	SelectionRingVisible *bool `json:"selectionRingVisible"`
	HoverRingVisible     *bool `json:"hoverRingVisible"`
	ReachSphereVisible   *bool `json:"reachSphereVisible"`
	SceneVectorsVisible  *bool `json:"sceneVectorsVisible"`
}

func LoadSceneOverlays(overlaysPath string) (viewstate.OverlayState, bool) {
	ov := viewstate.DefaultOverlayState()
	var sf sceneOverlaysFile
	jsonpersist.ReadJSONBestEffort(overlaysPath, &sf)
	found := false
	if sf.SceneToriVisible != nil {
		ov.SceneToriVisible = *sf.SceneToriVisible
		found = true
	}
	if sf.ScenePolesVisible != nil {
		ov.ScenePolesVisible = *sf.ScenePolesVisible
		found = true
	}
	if sf.NodePolesVisible != nil {
		ov.NodePolesVisible = *sf.NodePolesVisible
		found = true
	}
	if sf.HandholdsVisible != nil {
		ov.HandholdsVisible = *sf.HandholdsVisible
		found = true
	}
	if sf.OverlaysActive != nil {
		ov.OverlaysVisible = *sf.OverlaysActive
		found = true
	}
	if sf.LabelsGlobalHidden != nil {
		ov.LabelsGlobalVisible = !*sf.LabelsGlobalHidden
		found = true
	}
	if sf.NodeBodyVisible != nil {
		ov.NodeBodyVisible = *sf.NodeBodyVisible
		found = true
	}
	if sf.NodeRingVisible != nil {
		ov.NodeRingVisible = *sf.NodeRingVisible
		found = true
	}
	if sf.RingPickVisible != nil {
		ov.RingPickVisible = *sf.RingPickVisible
		found = true
	}
	if sf.SelectionRingVisible != nil {
		ov.SelectionRingVisible = *sf.SelectionRingVisible
		found = true
	}
	if sf.HoverRingVisible != nil {
		ov.HoverRingVisible = *sf.HoverRingVisible
		found = true
	}
	if sf.ReachSphereVisible != nil {
		ov.ReachSphereVisible = *sf.ReachSphereVisible
		found = true
	}
	if sf.SceneVectorsVisible != nil {
		ov.SceneVectorsVisible = *sf.SceneVectorsVisible
		found = true
	}
	return ov, found
}

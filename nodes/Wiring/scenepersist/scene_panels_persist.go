package scenepersist

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/scenepaths"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewstate"
)

func InstallPanels(ui *viewstate.UIState, topologyPath string) {
	pn, _ := LoadScenePanels(scenepaths.PanelsFilePath(topologyPath))
	ui.PN.SetPanelState(pn)
}

type scenePanelsFile struct {
	Overlays bool `json:"overlays"`

	Node      bool `json:"node"`
	NodeShape bool `json:"nodeShape"`
	NodeState bool `json:"nodeState"`
	NodePoles bool `json:"nodePoles"`
	NodeRules bool `json:"nodeRules"`

	Scene        bool `json:"scene"`
	SceneGuides  bool `json:"sceneGuides"`
	ScenePoles   bool `json:"scenePoles"`
	SceneVectors bool `json:"sceneVectors"`
	SceneLabels  bool `json:"sceneLabels"`
}

func WriteScenePanels(panelsPath string, pn viewstate.PanelState) error {
	return jsonpersist.WriteJSONAtomic(panelsPath, scenePanelsFile{
		Overlays: pn.OverlaysOpen,

		Node:      pn.NodeOpen,
		NodeShape: pn.NodeShapeOpen,
		NodeState: pn.NodeStateOpen,
		NodePoles: pn.NodePolesOpen,
		NodeRules: pn.NodeRulesOpen,

		Scene:        pn.SceneOpen,
		SceneGuides:  pn.SceneGuidesOpen,
		ScenePoles:   pn.ScenePolesOpen,
		SceneVectors: pn.SceneVectorsOpen,
		SceneLabels:  pn.SceneLabelsOpen,
	})
}

func LoadScenePanels(panelsPath string) (viewstate.PanelState, bool) {
	pn := viewstate.DefaultPanelState()
	var sf scenePanelsFile
	jsonpersist.ReadJSONBestEffort(panelsPath, &sf)
	found := true
	pn.OverlaysOpen = sf.Overlays

	pn.NodeOpen = sf.Node
	pn.NodeShapeOpen = sf.NodeShape
	pn.NodeStateOpen = sf.NodeState
	pn.NodePolesOpen = sf.NodePoles
	pn.NodeRulesOpen = sf.NodeRules

	pn.SceneOpen = sf.Scene
	pn.SceneGuidesOpen = sf.SceneGuides
	pn.ScenePolesOpen = sf.ScenePoles
	pn.SceneVectorsOpen = sf.SceneVectors
	pn.SceneLabelsOpen = sf.SceneLabels
	return pn, found
}

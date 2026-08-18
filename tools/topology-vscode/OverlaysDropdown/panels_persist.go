package OverlaysDropdown

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
)

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

func WriteScenePanels(panelsPath string, pn PanelState) error {
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

func LoadScenePanels(panelsPath string) (PanelState, bool) {
	pn := DefaultPanelState()
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

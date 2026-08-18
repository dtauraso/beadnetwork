package OverlaysDropdown

type PanelState struct {
	OverlaysOpen bool

	NodeOpen      bool
	NodeShapeOpen bool
	NodeStateOpen bool
	NodePolesOpen bool
	NodeRulesOpen bool

	SceneOpen        bool
	SceneGuidesOpen  bool
	ScenePolesOpen   bool
	SceneVectorsOpen bool
	SceneLabelsOpen  bool
}

func DefaultPanelState() PanelState {
	return PanelState{}
}

func (p *PanelState) setFlag(field *bool) {
	*field = !*field
}

func (p *PanelState) TogglePanelOverlays() { p.setFlag(&p.OverlaysOpen) }

func (p *PanelState) TogglePanelNode() { p.setFlag(&p.NodeOpen) }

func (p *PanelState) TogglePanelNodeShape() { p.setFlag(&p.NodeShapeOpen) }

func (p *PanelState) TogglePanelNodeState() { p.setFlag(&p.NodeStateOpen) }

func (p *PanelState) TogglePanelNodePoles() { p.setFlag(&p.NodePolesOpen) }

func (p *PanelState) TogglePanelNodeRules() { p.setFlag(&p.NodeRulesOpen) }

func (p *PanelState) TogglePanelScene() { p.setFlag(&p.SceneOpen) }

func (p *PanelState) TogglePanelSceneGuides() { p.setFlag(&p.SceneGuidesOpen) }

func (p *PanelState) TogglePanelScenePoles() { p.setFlag(&p.ScenePolesOpen) }

func (p *PanelState) TogglePanelSceneVectors() { p.setFlag(&p.SceneVectorsOpen) }

func (p *PanelState) TogglePanelSceneLabels() { p.setFlag(&p.SceneLabelsOpen) }

// PANEL_TOGGLES_START
var PanelToggles = map[string]func(*PanelState){
	"overlays":     (*PanelState).TogglePanelOverlays,
	"node":         (*PanelState).TogglePanelNode,
	"nodeShape":    (*PanelState).TogglePanelNodeShape,
	"nodeState":    (*PanelState).TogglePanelNodeState,
	"nodePoles":    (*PanelState).TogglePanelNodePoles,
	"nodeRules":    (*PanelState).TogglePanelNodeRules,
	"scene":        (*PanelState).TogglePanelScene,
	"sceneGuides":  (*PanelState).TogglePanelSceneGuides,
	"scenePoles":   (*PanelState).TogglePanelScenePoles,
	"sceneVectors": (*PanelState).TogglePanelSceneVectors,
	"sceneLabels":  (*PanelState).TogglePanelSceneLabels,
}

// PANEL_TOGGLES_END

var PanelOpen = map[string]func(*PanelState) bool{
	"overlays":     func(p *PanelState) bool { return p.OverlaysOpen },
	"node":         func(p *PanelState) bool { return p.NodeOpen },
	"nodeShape":    func(p *PanelState) bool { return p.NodeShapeOpen },
	"nodeState":    func(p *PanelState) bool { return p.NodeStateOpen },
	"nodePoles":    func(p *PanelState) bool { return p.NodePolesOpen },
	"nodeRules":    func(p *PanelState) bool { return p.NodeRulesOpen },
	"scene":        func(p *PanelState) bool { return p.SceneOpen },
	"sceneGuides":  func(p *PanelState) bool { return p.SceneGuidesOpen },
	"scenePoles":   func(p *PanelState) bool { return p.ScenePolesOpen },
	"sceneVectors": func(p *PanelState) bool { return p.SceneVectorsOpen },
	"sceneLabels":  func(p *PanelState) bool { return p.SceneLabelsOpen },
}

func (p *PanelState) SetPanelState(v PanelState) {
	*p = v
}

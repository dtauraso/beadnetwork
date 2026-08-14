package viewstate

type PanelState struct {
	OverlaysOpen bool

	NodeOpen      bool
	NodeShapeOpen bool
	NodeStateOpen bool
	NodeReachOpen bool

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

func (p *PanelState) TogglePanelNodeReach() { p.setFlag(&p.NodeReachOpen) }

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
	"nodeReach":    (*PanelState).TogglePanelNodeReach,
	"scene":        (*PanelState).TogglePanelScene,
	"sceneGuides":  (*PanelState).TogglePanelSceneGuides,
	"scenePoles":   (*PanelState).TogglePanelScenePoles,
	"sceneVectors": (*PanelState).TogglePanelSceneVectors,
	"sceneLabels":  (*PanelState).TogglePanelSceneLabels,
}

// PANEL_TOGGLES_END

func (p *PanelState) SetPanelState(v PanelState) {
	*p = v
}

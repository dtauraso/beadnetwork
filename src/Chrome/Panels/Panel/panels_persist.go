package Panel

import (
	"path/filepath"
	"sort"

	"github.com/dtauraso/wirefold/src/Node/Wiring/jsonpersist"
)

var PanelFlagRead = map[string]func(*PanelState) bool{
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

var PanelFlagWrite = map[string]func(*PanelState, bool){
	"overlays":     func(p *PanelState, v bool) { p.OverlaysOpen = v },
	"node":         func(p *PanelState, v bool) { p.NodeOpen = v },
	"nodeShape":    func(p *PanelState, v bool) { p.NodeShapeOpen = v },
	"nodeState":    func(p *PanelState, v bool) { p.NodeStateOpen = v },
	"nodePoles":    func(p *PanelState, v bool) { p.NodePolesOpen = v },
	"nodeRules":    func(p *PanelState, v bool) { p.NodeRulesOpen = v },
	"scene":        func(p *PanelState, v bool) { p.SceneOpen = v },
	"sceneGuides":  func(p *PanelState, v bool) { p.SceneGuidesOpen = v },
	"scenePoles":   func(p *PanelState, v bool) { p.ScenePolesOpen = v },
	"sceneVectors": func(p *PanelState, v bool) { p.SceneVectorsOpen = v },
	"sceneLabels":  func(p *PanelState, v bool) { p.SceneLabelsOpen = v },
}

func PanelFlagFile(panelsDir, name string) string {
	return filepath.Join(panelsDir, name+".json")
}

func WriteScenePanels(panelsDir string, pn PanelState) error {
	names := make([]string, 0, len(PanelFlagRead))
	for name := range PanelFlagRead {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if err := jsonpersist.WriteJSONAtomic(PanelFlagFile(panelsDir, name), PanelFlagRead[name](&pn)); err != nil {
			return err
		}
	}
	return nil
}

func LoadScenePanels(panelsDir string) (PanelState, bool) {
	pn := DefaultPanelState()
	found := false
	for name, set := range PanelFlagWrite {
		var v bool
		if jsonpersist.ReadJSONIfExists(PanelFlagFile(panelsDir, name), &v) {
			set(&pn, v)
			found = true
		}
	}
	return pn, found
}

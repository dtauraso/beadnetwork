package Panel

import (
	"path/filepath"
)

var PanelFlagRead = map[string]func(*PanelState) bool{
	"overlays":     func(p *PanelState) bool { return p.OverlaysOpen },
	"node":         func(p *PanelState) bool { return p.NodeOpen },
	"nodeShape":    func(p *PanelState) bool { return p.NodeShapeOpen },
	"nodeState":    func(p *PanelState) bool { return p.NodeStateOpen },
	"nodePoles":    func(p *PanelState) bool { return p.NodePolesOpen },
	"nodeRules":    func(p *PanelState) bool { return p.NodeRulesOpen },
	"nodeVectors":  func(p *PanelState) bool { return p.NodeVectorsOpen },
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
	"nodeVectors":  func(p *PanelState, v bool) { p.NodeVectorsOpen = v },
	"scene":        func(p *PanelState, v bool) { p.SceneOpen = v },
	"sceneGuides":  func(p *PanelState, v bool) { p.SceneGuidesOpen = v },
	"scenePoles":   func(p *PanelState, v bool) { p.ScenePolesOpen = v },
	"sceneVectors": func(p *PanelState, v bool) { p.SceneVectorsOpen = v },
	"sceneLabels":  func(p *PanelState, v bool) { p.SceneLabelsOpen = v },
}

func BlockPath(sceneRoot string) string {
	return filepath.Join(sceneRoot, filepath.FromSlash(BlockRelPath))
}

func WriteScenePanels(sceneRoot string, pn PanelState) error {
	w := NewBlobWriter(BlockPath(sceneRoot), FlagNames)
	w.Begin()
	for _, name := range FlagNames {
		w.Bool(name, PanelFlagRead[name](&pn))
	}
	return w.Flush()
}

func LoadScenePanels(sceneRoot string) (PanelState, bool) {
	pn := DefaultPanelState()
	r, ok := ReadBlob(BlockPath(sceneRoot), FlagNames)
	if !ok {
		return pn, false
	}
	found := false
	for _, name := range FlagNames {
		v, got := r.Bool(name)
		if !got {
			continue
		}
		PanelFlagWrite[name](&pn, v)
		found = true
	}
	return pn, found
}

// scene_overlays_persist.go — pure read/write helpers for Go's OWN overlay-visibility
// state in view/overlays.json, lifted out of nodes/Wiring/scene_overlays_persist.go (which
// keeps the MoveDispatch-facing LoadOverlays method and the overlaysPersister type — see
// that file's own header).
//
// WHOLE-FILE write (one-file-per-writer): overlays.json holds ONLY these flags and has
// exactly one writer, so each flush marshals the current overlayState fresh — no
// read-modify-write. The key names + polarity + default-omission mirror TS's
// serializeSceneState (webview/state/viewer/types.ts): most flags are visible-sense
// written only when hidden (false); labelsGlobalHidden/badgesHidden are hidden-sense
// written only when hidden (true); a key at its default is deleted so the on-disk shape
// matches what the editor would have written.
package scenepersist

import (
	"encoding/json"

	"github.com/dtauraso/wirefold/nodes/Wiring/jsonpersist"
	"github.com/dtauraso/wirefold/nodes/Wiring/viewstate"
)

// WriteSceneOverlays writes the current overlay-visibility snapshot as the WHOLE content of
// overlaysPath (overlays.json) — the sole writer of that file, so each call builds a fresh
// object (no read-modify-write of any prior content).
func WriteSceneOverlays(overlaysPath string, ov viewstate.OverlayState) error {
	obj := map[string]json.RawMessage{}
	// visible-sense: default true — write `false` only when hidden, else drop the key.
	setVisible := func(key string, visible bool) {
		if !visible {
			obj[key] = json.RawMessage("false")
		}
	}
	// hidden-sense: default visible — write `true` only when hidden, else drop the key.
	setHidden := func(key string, visible bool) {
		if !visible {
			obj[key] = json.RawMessage("true")
		}
	}

	setVisible("sceneToriVisible", ov.SceneToriVisible)
	setVisible("scenePolesVisible", ov.ScenePolesVisible)
	setVisible("nodePolesVisible", ov.NodePolesVisible)
	setVisible("selSpherePolesVisible", ov.SelSpherePolesVisible)
	setVisible("handholdsVisible", ov.HandholdsVisible)
	setVisible("overlaysActive", ov.OverlaysVisible)
	setHidden("labelsGlobalHidden", ov.LabelsGlobalVisible)
	setVisible("nodeBodyVisible", ov.NodeBodyVisible)
	setVisible("nodeRingVisible", ov.NodeRingVisible)
	setVisible("ringPickVisible", ov.RingPickVisible)
	setVisible("selectionRingVisible", ov.SelectionRingVisible)
	setVisible("hoverRingVisible", ov.HoverRingVisible)
	setVisible("reachSphereVisible", ov.ReachSphereVisible)
	return jsonpersist.WriteJSONAtomic(overlaysPath, obj)
}

// sceneOverlaysFile is the on-disk shape of overlays.json the overlay loader reads. Pointer fields
// distinguish an ABSENT key (keep the code default) from a present false/true — the writer
// omits any key at its default, so absence must not be read as false. Key names + polarity
// mirror WriteSceneOverlays (setVisible / setHidden) exactly.
type sceneOverlaysFile struct {
	SceneToriVisible      *bool `json:"sceneToriVisible"`
	ScenePolesVisible     *bool `json:"scenePolesVisible"`
	NodePolesVisible      *bool `json:"nodePolesVisible"`
	SelSpherePolesVisible *bool `json:"selSpherePolesVisible"`
	HandholdsVisible      *bool `json:"handholdsVisible"`
	OverlaysActive        *bool `json:"overlaysActive"`
	LabelsGlobalHidden    *bool `json:"labelsGlobalHidden"`
	NodeBodyVisible       *bool `json:"nodeBodyVisible"`
	NodeRingVisible       *bool `json:"nodeRingVisible"`
	RingPickVisible       *bool `json:"ringPickVisible"`
	SelectionRingVisible  *bool `json:"selectionRingVisible"`
	HoverRingVisible      *bool `json:"hoverRingVisible"`
	ReachSphereVisible    *bool `json:"reachSphereVisible"`
}

// LoadSceneOverlays reads the persisted overlay-visibility snapshot from overlaysPath
// (overlays.json), applying the same key names + polarity the writer used (visible-sense
// keys straight through; the two *Hidden keys inverted back to visible-sense). Starts from
// viewstate.DefaultOverlayState so any key the writer omitted (because it was at its
// default) keeps the code default. The bool return is false when the file yields no overlay
// key (fresh topology) — the caller then keeps the code defaults.
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
	if sf.SelSpherePolesVisible != nil {
		ov.SelSpherePolesVisible = *sf.SelSpherePolesVisible
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
	return ov, found
}

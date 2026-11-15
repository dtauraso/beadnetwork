package View

import Flags "github.com/dtauraso/beadnetwork/Categories/Scene/View/Flags"

func (ui *UIState) SceneOverlays(ov Flags.OverlayState) Flags.OverlayState {
	return Flags.Without(Flags.With(ov, ui.SceneFixedOverlays), ui.SceneRemovedOverlays)
}

func (ui *UIState) SceneUntoggledOverlays() []string {
	out := append([]string{}, ui.SceneHiddenOverlays...)
	out = append(out, ui.SceneFixedOverlays...)
	return append(out, ui.SceneRemovedOverlays...)
}

package Overlay

import "github.com/dtauraso/wirefold/src/Input/Codec"

func init() { Codec.RegisterUpdateDecoder("overlays", decodeUpdate) }

func decodeUpdate(r *Codec.Reader, attr byte) (Codec.StdinMsg, bool) {
	if attr != attrToggle {
		return Codec.StdinMsg{}, false
	}
	flagID, err := r.U8()
	if err != nil || int(flagID) >= len(Codec.InOverlayFlags) {
		return Codec.StdinMsg{}, false
	}
	return Codec.StdinMsg{Type: "edit", Op: "update", Kind: "overlays", Attr: "toggle", Flag: Codec.InOverlayFlags[flagID]}, true
}
